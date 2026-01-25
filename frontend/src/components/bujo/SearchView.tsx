import { useState, useCallback, useEffect } from 'react';
import { Search as SearchIcon } from 'lucide-react';
import { Search, GetEntry, MarkEntryDone, MarkEntryUndone, CancelEntry, UncancelEntry, CyclePriority, RetypeEntry } from '@/wailsjs/go/wails/App';
import { format } from 'date-fns';
import { cn } from '@/lib/utils';
import { ENTRY_SYMBOLS, EntryType, Priority, PRIORITY_SYMBOLS } from '@/types/bujo';
import { AnswerQuestionModal } from './AnswerQuestionModal';

export interface SearchResult {
  id: number;
  content: string;
  type: EntryType;
  priority: Priority;
  date: string;
  parentId: number | null;
}

interface SearchViewProps {
  onMigrate?: (entry: SearchResult) => void;
  onNavigateToEntry?: (entry: SearchResult) => void;
  onSelectEntry?: (entry: SearchResult) => void;
}

export function SearchView({ onMigrate, onNavigateToEntry, onSelectEntry }: SearchViewProps) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [hasSearched, setHasSearched] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(-1);
  const [answerModalOpen, setAnswerModalOpen] = useState(false);
  const [questionToAnswer, setQuestionToAnswer] = useState<SearchResult | null>(null);

  const handleSearch = useCallback(async (searchQuery: string) => {
    setQuery(searchQuery);
    setSelectedIndex(-1);

    if (!searchQuery.trim()) {
      setResults([]);
      setHasSearched(false);
      return;
    }

    try {
      const searchResults = await Search(searchQuery);
      const mappedResults = (searchResults || []).map(entry => ({
        id: entry.ID,
        content: entry.Content,
        type: (entry.Type as string).toLowerCase() as EntryType,
        priority: ((entry.Priority as string)?.toLowerCase() || 'none') as Priority,
        date: (entry.CreatedAt as unknown as string) || '',
        parentId: entry.ParentID ?? null,
      }));
      setResults(mappedResults);
      setHasSearched(true);
    } catch (error) {
      console.error('Search failed:', error);
      setResults([]);
      setHasSearched(true);
    }
  }, []);

  const refreshEntry = useCallback(async (oldId: number) => {
    const updated = await GetEntry(oldId);
    if (updated) {
      setResults(prev => prev.map(r =>
        r.id === oldId ? {
          id: updated.ID,
          content: updated.Content,
          type: updated.Type as EntryType,
          priority: ((updated.Priority as string)?.toLowerCase() || 'none') as Priority,
          date: (updated.CreatedAt as unknown as string) || r.date,
          parentId: updated.ParentID ?? null,
        } : r
      ));
    }
  }, []);

  const handleMarkDone = useCallback(async (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await MarkEntryDone(id);
      await refreshEntry(id);
    } catch (error) {
      console.error('Failed to mark done:', error);
    }
  }, [refreshEntry]);

  const handleMarkUndone = useCallback(async (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await MarkEntryUndone(id);
      await refreshEntry(id);
    } catch (error) {
      console.error('Failed to mark undone:', error);
    }
  }, [refreshEntry]);

  const handleCycleTypeKeyboard = useCallback(async (id: number, currentType: EntryType) => {
    const cycleOrder: EntryType[] = ['task', 'note', 'event', 'question'];
    const currentIdx = cycleOrder.indexOf(currentType);
    if (currentIdx === -1) return;
    const nextType = cycleOrder[(currentIdx + 1) % cycleOrder.length];
    try {
      await RetypeEntry(id, nextType);
      await refreshEntry(id);
    } catch (error) {
      console.error('Failed to cycle type:', error);
    }
  }, [refreshEntry]);

  const handleAnswer = useCallback((result: SearchResult) => {
    setQuestionToAnswer(result);
    setAnswerModalOpen(true);
  }, []);

  const handleAnswerSubmitted = useCallback(async () => {
    setAnswerModalOpen(false);
    const answeredId = questionToAnswer?.id;
    setQuestionToAnswer(null);
    if (answeredId) {
      await refreshEntry(answeredId);
    }
  }, [questionToAnswer, refreshEntry]);

  const handleResultClick = useCallback((result: SearchResult) => {
    onSelectEntry?.(result);
  }, [onSelectEntry]);

  useEffect(() => {
    const handleKeyDown = async (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      const isInputFocused = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA';
      if (isInputFocused) return;

      if (results.length === 0) return;

      switch (e.key) {
        case 'j':
        case 'ArrowDown':
          e.preventDefault();
          setSelectedIndex(prev => Math.min(prev + 1, results.length - 1));
          break;
        case 'k':
        case 'ArrowUp':
          e.preventDefault();
          setSelectedIndex(prev => Math.max(prev - 1, 0));
          break;
        case ' ':
          e.preventDefault();
          if (selectedIndex >= 0 && selectedIndex < results.length) {
            const selected = results[selectedIndex];
            if (selected.type === 'task') {
              await MarkEntryDone(selected.id);
              await refreshEntry(selected.id);
            } else if (selected.type === 'done') {
              await MarkEntryUndone(selected.id);
              await refreshEntry(selected.id);
            }
          }
          break;
        case 'x':
          e.preventDefault();
          if (selectedIndex >= 0 && selectedIndex < results.length) {
            const selected = results[selectedIndex];
            if (selected.type === 'cancelled') {
              await UncancelEntry(selected.id);
              await refreshEntry(selected.id);
            } else {
              await CancelEntry(selected.id);
              await refreshEntry(selected.id);
            }
          }
          break;
        case 'p':
          e.preventDefault();
          if (selectedIndex >= 0 && selectedIndex < results.length) {
            const selected = results[selectedIndex];
            await CyclePriority(selected.id);
            await refreshEntry(selected.id);
          }
          break;
        case 't':
          e.preventDefault();
          if (selectedIndex >= 0 && selectedIndex < results.length) {
            const selected = results[selectedIndex];
            await handleCycleTypeKeyboard(selected.id, selected.type);
          }
          break;
        case 'a':
          e.preventDefault();
          if (selectedIndex >= 0 && selectedIndex < results.length) {
            const selected = results[selectedIndex];
            if (selected.type === 'question') {
              handleAnswer(selected);
            }
          }
          break;
        case 'm':
          e.preventDefault();
          if (selectedIndex >= 0 && selectedIndex < results.length) {
            onMigrate?.(results[selectedIndex]);
          }
          break;
        case 'Enter':
          e.preventDefault();
          if (selectedIndex >= 0 && selectedIndex < results.length) {
            onNavigateToEntry?.(results[selectedIndex]);
          }
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [results, selectedIndex, refreshEntry, handleCycleTypeKeyboard, handleAnswer, onMigrate, onNavigateToEntry]);

  const formatDate = (dateStr: string) => {
    try {
      const date = new Date(dateStr);
      return format(date, 'MMM d, yyyy');
    } catch {
      return dateStr;
    }
  };

  const getSymbol = (type: EntryType): string => {
    return ENTRY_SYMBOLS[type] || '•';
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-2">
        <SearchIcon className="w-5 h-5 text-primary" data-testid="search-icon" />
        <h2 className="font-display text-xl font-semibold">Search</h2>
      </div>

      {/* Search Input */}
      <div className="relative">
        <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
        <input
          type="text"
          value={query}
          onChange={(e) => handleSearch(e.target.value)}
          placeholder="Search entries..."
          className="w-full pl-10 pr-4 py-3 text-sm rounded-lg border border-border bg-background focus:outline-none focus:ring-2 focus:ring-primary/50"
        />
      </div>

      {/* Results */}
      <div className="space-y-2">
        {!hasSearched && (
          <p className="text-sm text-muted-foreground italic py-6 text-center">
            Enter a search term to find entries
          </p>
        )}

        {hasSearched && results.length === 0 && (
          <p className="text-sm text-muted-foreground italic py-6 text-center">
            No results found for "{query}"
          </p>
        )}

        {results.map((result, index) => {
          const isSelected = index === selectedIndex;

          return (
            <div
              key={result.id}
              onClick={() => handleResultClick(result)}
              onDoubleClick={() => onNavigateToEntry?.(result)}
              className={cn(
                'p-3 rounded-lg border border-border cursor-pointer',
                'bg-card transition-colors group',
                !isSelected && 'hover:bg-secondary/30',
                isSelected && 'ring-2 ring-primary'
              )}
            >
              {/* Main result row */}
              <div
                className="flex items-start gap-3"
                data-result-id={result.id}
              >
                {/* Context dot - indicates entry has ancestors */}
                {result.parentId !== null ? (
                  <span
                    data-testid="context-dot"
                    className="w-1.5 h-1.5 mt-2 rounded-full bg-muted-foreground flex-shrink-0"
                    title="Has parent context"
                  />
                ) : (
                  <span className="w-1.5" />
                )}

                <span className="inline-flex items-center gap-1 flex-shrink-0">
                  {/* Symbol - clickable for task/done entries */}
                  {result.type === 'task' || result.type === 'done' ? (
                    <button
                      data-testid="entry-symbol"
                      onClick={(e) => result.type === 'task' ? handleMarkDone(result.id, e) : handleMarkUndone(result.id, e)}
                      title={result.type === 'task' ? 'Task' : 'Done'}
                      className={cn(
                        'text-lg w-5 text-center cursor-pointer hover:opacity-70 transition-opacity',
                        result.type === 'done' && 'text-bujo-done',
                        result.type === 'task' && 'text-bujo-task',
                      )}
                    >
                      {getSymbol(result.type)}
                    </button>
                  ) : (
                    <span
                      data-testid="entry-symbol"
                      className={cn(
                        'text-lg w-5 text-center',
                        result.type === 'note' && 'text-bujo-note',
                        result.type === 'event' && 'text-bujo-event',
                        result.type === 'cancelled' && 'text-bujo-cancelled',
                      )}
                    >
                      {getSymbol(result.type)}
                    </span>
                  )}
                  {result.priority !== 'none' && (
                    <span className={cn(
                      'text-xs font-bold',
                      result.priority === 'low' && 'text-priority-low',
                      result.priority === 'medium' && 'text-priority-medium',
                      result.priority === 'high' && 'text-priority-high',
                    )}>
                      {PRIORITY_SYMBOLS[result.priority]}
                    </span>
                  )}
                </span>
                <div className="flex-1 min-w-0">
                  <p className={cn(
                    'text-sm',
                    result.type === 'done' && 'text-bujo-done',
                    result.type === 'cancelled' && 'line-through text-muted-foreground'
                  )}>
                    {result.content}
                  </p>
                  <p className="text-xs text-muted-foreground mt-1">
                    {formatDate(result.date)}
                  </p>
                </div>

                <span className="text-xs text-muted-foreground opacity-0 group-hover:opacity-100">
                  #{result.id}
                </span>
              </div>
            </div>
          );
        })}
      </div>

      {/* Answer Question Modal */}
      {questionToAnswer && (
        <AnswerQuestionModal
          isOpen={answerModalOpen}
          questionId={questionToAnswer.id}
          questionContent={questionToAnswer.content}
          onClose={() => { setAnswerModalOpen(false); setQuestionToAnswer(null); }}
          onAnswered={handleAnswerSubmitted}
        />
      )}
    </div>
  );
}
