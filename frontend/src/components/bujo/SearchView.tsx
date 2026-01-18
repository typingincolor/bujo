import { useState, useCallback, useEffect } from 'react';
import { Search as SearchIcon, Check, X, RotateCcw, Trash2, Pencil, ArrowRight, Flag, RefreshCw, ChevronUp } from 'lucide-react';
import { Search, GetEntry, GetEntryAncestors, MarkEntryDone, MarkEntryUndone, CancelEntry, UncancelEntry, DeleteEntry, CyclePriority, RetypeEntry } from '@/wailsjs/go/wails/App';
import { format } from 'date-fns';
import { cn } from '@/lib/utils';
import { ENTRY_SYMBOLS, EntryType, Priority, PRIORITY_SYMBOLS } from '@/types/bujo';

function ActionPlaceholder() {
  return <span data-action-slot className="p-1 w-6 h-6" aria-hidden="true" />;
}

interface SearchResult {
  id: number;
  content: string;
  type: EntryType;
  priority: Priority;
  date: string;
  parentId: number | null;
}

interface AncestorEntry {
  id: number;
  content: string;
  type: EntryType;
}

export function SearchView() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [hasSearched, setHasSearched] = useState(false);
  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set());
  const [ancestorsMap, setAncestorsMap] = useState<Map<number, AncestorEntry[]>>(new Map());
  const [selectedIndex, setSelectedIndex] = useState(-1);

  const handleSearch = useCallback(async (searchQuery: string) => {
    setQuery(searchQuery);
    setExpandedIds(new Set());
    setAncestorsMap(new Map());
    setSelectedIndex(-1);

    if (!searchQuery.trim()) {
      setResults([]);
      setHasSearched(false);
      return;
    }

    try {
      const searchResults = await Search(searchQuery);
      setResults((searchResults || []).map(entry => ({
        id: entry.ID,
        content: entry.Content,
        type: entry.Type as EntryType,
        priority: ((entry.Priority as string)?.toLowerCase() || 'none') as Priority,
        date: (entry.CreatedAt as unknown as string) || '',
        parentId: entry.ParentID ?? null,
      })));
      setHasSearched(true);
    } catch (error) {
      console.error('Search failed:', error);
      setResults([]);
      setHasSearched(true);
    }
  }, []);

  const toggleExpanded = useCallback(async (result: SearchResult) => {
    const newExpanded = new Set(expandedIds);

    if (newExpanded.has(result.id)) {
      newExpanded.delete(result.id);
    } else {
      newExpanded.add(result.id);

      if (!ancestorsMap.has(result.id) && result.parentId !== null) {
        try {
          const ancestors = await GetEntryAncestors(result.id);
          setAncestorsMap(prev => {
            const next = new Map(prev);
            next.set(result.id, (ancestors || []).map(a => ({
              id: a.ID,
              content: a.Content,
              type: a.Type as EntryType,
            })));
            return next;
          });
        } catch (error) {
          console.error('Failed to load ancestors:', error);
        }
      }
    }

    setExpandedIds(newExpanded);
  }, [expandedIds, ancestorsMap]);

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

  const handleCancel = useCallback(async (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await CancelEntry(id);
      await refreshEntry(id);
    } catch (error) {
      console.error('Failed to cancel entry:', error);
    }
  }, [refreshEntry]);

  const handleUncancel = useCallback(async (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await UncancelEntry(id);
      await refreshEntry(id);
    } catch (error) {
      console.error('Failed to uncancel entry:', error);
    }
  }, [refreshEntry]);

  const handleDelete = useCallback(async (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await DeleteEntry(id);
      setResults(prev => prev.filter(r => r.id !== id));
    } catch (error) {
      console.error('Failed to delete entry:', error);
    }
  }, []);

  const handleCyclePriority = useCallback(async (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await CyclePriority(id);
      await refreshEntry(id);
    } catch (error) {
      console.error('Failed to cycle priority:', error);
    }
  }, [refreshEntry]);

  const handleCycleType = useCallback(async (id: number, currentType: EntryType, e: React.MouseEvent) => {
    e.stopPropagation();
    const cycleOrder: EntryType[] = ['task', 'note', 'event', 'question'];
    const currentIndex = cycleOrder.indexOf(currentType);
    if (currentIndex === -1) return;
    const nextType = cycleOrder[(currentIndex + 1) % cycleOrder.length];
    try {
      await RetypeEntry(id, nextType);
      await refreshEntry(id);
    } catch (error) {
      console.error('Failed to cycle type:', error);
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
        case 'Enter':
          e.preventDefault();
          if (selectedIndex >= 0 && selectedIndex < results.length) {
            toggleExpanded(results[selectedIndex]);
          }
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [results, selectedIndex, refreshEntry, toggleExpanded, handleCycleTypeKeyboard]);

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
          const isExpanded = expandedIds.has(result.id);
          const ancestors = ancestorsMap.get(result.id) || [];
          const isSelected = index === selectedIndex;

          return (
            <div
              key={result.id}
              onClick={() => toggleExpanded(result)}
              className={cn(
                'p-3 rounded-lg border border-border cursor-pointer',
                'bg-card transition-colors group',
                !isSelected && 'hover:bg-secondary/30',
                isSelected && 'ring-2 ring-primary'
              )}
            >
              {/* Ancestors context */}
              {isExpanded && ancestors.length > 0 && (
                <div className="mb-2 pb-2 border-b border-border/50 space-y-1">
                  {ancestors.map((ancestor, index) => (
                    <div
                      key={ancestor.id}
                      className="flex items-center gap-2 text-xs text-muted-foreground"
                      style={{ paddingLeft: `${index * 20}px` }}
                    >
                      <span className="font-mono">{getSymbol(ancestor.type)}</span>
                      <span>{ancestor.content}</span>
                    </div>
                  ))}
                </div>
              )}

              {/* Main result row */}
              <div
                className="flex items-start gap-3"
                data-result-id={result.id}
                style={{ paddingLeft: isExpanded && ancestors.length > 0 ? `${ancestors.length * 20}px` : undefined }}
              >
                {/* Context indicator - shows when entry has parent and isn't expanded */}
                {result.parentId !== null && !isExpanded && (
                  <span title="Has parent context">
                    <ChevronUp
                      className="w-4 h-4 text-muted-foreground flex-shrink-0"
                      aria-label="Has parent context"
                    />
                  </span>
                )}
                <span className="inline-flex items-center gap-1 flex-shrink-0">
                  <span className={cn(
                    'text-lg font-mono w-5 text-center',
                    result.type === 'done' && 'text-bujo-done',
                    result.type === 'task' && 'text-bujo-task',
                    result.type === 'note' && 'text-bujo-note',
                    result.type === 'event' && 'text-bujo-event',
                    result.type === 'cancelled' && 'text-bujo-cancelled',
                  )}>
                    {getSymbol(result.type)}
                  </span>
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

                {/* Action buttons */}
                {result.type === 'task' ? (
                  <button
                    data-action-slot
                    onClick={(e) => handleMarkDone(result.id, e)}
                    title="Mark done"
                    className="p-1 rounded hover:bg-secondary/50 text-muted-foreground hover:text-bujo-done"
                  >
                    <Check className="w-4 h-4" />
                  </button>
                ) : result.type === 'done' ? (
                  <button
                    data-action-slot
                    onClick={(e) => handleMarkUndone(result.id, e)}
                    title="Mark undone"
                    className="p-1 rounded hover:bg-orange-500/20 text-muted-foreground hover:text-orange-600"
                  >
                    <span className="text-sm font-bold leading-none">•</span>
                  </button>
                ) : (
                  <ActionPlaceholder />
                )}
                {result.type !== 'cancelled' ? (
                  <button
                    data-action-slot
                    onClick={(e) => handleCancel(result.id, e)}
                    title="Cancel entry"
                    className="p-1 rounded hover:bg-warning/20 text-muted-foreground hover:text-warning"
                  >
                    <X className="w-4 h-4" />
                  </button>
                ) : (
                  <button
                    data-action-slot
                    onClick={(e) => handleUncancel(result.id, e)}
                    title="Uncancel entry"
                    className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary"
                  >
                    <RotateCcw className="w-4 h-4" />
                  </button>
                )}
                <button
                  data-action-slot
                  onClick={(e) => handleCyclePriority(result.id, e)}
                  title="Cycle priority"
                  className="p-1 rounded hover:bg-warning/20 text-muted-foreground hover:text-warning"
                >
                  <Flag className="w-4 h-4" />
                </button>
                {(result.type === 'task' || result.type === 'note' || result.type === 'event' || result.type === 'question') ? (
                  <button
                    data-action-slot
                    onClick={(e) => handleCycleType(result.id, result.type, e)}
                    title="Change type"
                    className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary"
                  >
                    <RefreshCw className="w-4 h-4" />
                  </button>
                ) : (
                  <ActionPlaceholder />
                )}
                {result.type === 'task' ? (
                  <button
                    data-action-slot
                    onClick={(e) => e.stopPropagation()}
                    title="Migrate entry"
                    className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary"
                  >
                    <ArrowRight className="w-4 h-4" />
                  </button>
                ) : (
                  <ActionPlaceholder />
                )}
                <button
                  data-action-slot
                  onClick={(e) => e.stopPropagation()}
                  title="Edit entry"
                  className="p-1 rounded hover:bg-secondary text-muted-foreground hover:text-foreground"
                >
                  <Pencil className="w-4 h-4" />
                </button>
                <button
                  data-action-slot
                  onClick={(e) => handleDelete(result.id, e)}
                  title="Delete entry"
                  className="p-1 rounded hover:bg-destructive/20 text-muted-foreground hover:text-destructive"
                >
                  <Trash2 className="w-4 h-4" />
                </button>

                <span className="text-xs text-muted-foreground opacity-0 group-hover:opacity-100">
                  #{result.id}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
