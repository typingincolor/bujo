import { useState, useCallback } from 'react';
import { Search as SearchIcon, Check, X, RotateCcw, Trash2, Pencil, ArrowRight } from 'lucide-react';
import { Search, GetEntryAncestors, MarkEntryDone, MarkEntryUndone, CancelEntry, UncancelEntry, DeleteEntry, CyclePriority } from '@/wailsjs/go/wails/App';
import { format } from 'date-fns';
import { cn } from '@/lib/utils';
import { ENTRY_SYMBOLS, EntryType } from '@/types/bujo';

interface SearchResult {
  id: number;
  content: string;
  type: EntryType;
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

  const handleSearch = useCallback(async (searchQuery: string) => {
    setQuery(searchQuery);
    setExpandedIds(new Set());
    setAncestorsMap(new Map());

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

  const handleMarkDone = useCallback(async (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await MarkEntryDone(id);
      setResults(prev => prev.map(r =>
        r.id === id ? { ...r, type: 'done' as EntryType } : r
      ));
    } catch (error) {
      console.error('Failed to mark done:', error);
    }
  }, []);

  const handleMarkUndone = useCallback(async (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await MarkEntryUndone(id);
      setResults(prev => prev.map(r =>
        r.id === id ? { ...r, type: 'task' as EntryType } : r
      ));
    } catch (error) {
      console.error('Failed to mark undone:', error);
    }
  }, []);

  const handleCancel = useCallback(async (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await CancelEntry(id);
      setResults(prev => prev.map(r =>
        r.id === id ? { ...r, type: 'cancelled' as EntryType } : r
      ));
    } catch (error) {
      console.error('Failed to cancel entry:', error);
    }
  }, []);

  const handleUncancel = useCallback(async (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await UncancelEntry(id);
      setResults(prev => prev.map(r =>
        r.id === id ? { ...r, type: 'task' as EntryType } : r
      ));
    } catch (error) {
      console.error('Failed to uncancel entry:', error);
    }
  }, []);

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
    } catch (error) {
      console.error('Failed to cycle priority:', error);
    }
  }, []);

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

        {results.map((result) => {
          const isExpanded = expandedIds.has(result.id);
          const ancestors = ancestorsMap.get(result.id) || [];

          return (
            <div
              key={result.id}
              onClick={() => toggleExpanded(result)}
              className={cn(
                'p-3 rounded-lg border border-border cursor-pointer',
                'bg-card hover:bg-secondary/30 transition-colors group'
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
                <span className={cn(
                  'text-lg font-mono flex-shrink-0 w-5 text-center',
                  result.type === 'done' && 'text-bujo-done',
                  result.type === 'task' && 'text-bujo-task',
                  result.type === 'note' && 'text-bujo-note',
                  result.type === 'event' && 'text-bujo-event',
                )}>
                  {getSymbol(result.type)}
                </span>
                <div className="flex-1 min-w-0">
                  <p className={cn(
                    'text-sm',
                    result.type === 'done' && 'line-through text-muted-foreground'
                  )}>
                    {result.content}
                  </p>
                  <p className="text-xs text-muted-foreground mt-1">
                    {formatDate(result.date)}
                  </p>
                </div>

                {/* Action buttons */}
                {result.type === 'task' && (
                  <button
                    onClick={(e) => handleMarkDone(result.id, e)}
                    title="Mark done"
                    className="p-1 rounded hover:bg-secondary/50 text-muted-foreground hover:text-bujo-done"
                  >
                    <Check className="w-4 h-4" />
                  </button>
                )}
                {result.type === 'done' && (
                  <button
                    onClick={(e) => handleMarkUndone(result.id, e)}
                    title="Mark undone"
                    className="p-1 rounded hover:bg-orange-500/20 text-muted-foreground hover:text-orange-600"
                  >
                    <span className="text-sm font-bold leading-none">•</span>
                  </button>
                )}
                {result.type !== 'cancelled' && (
                  <button
                    onClick={(e) => handleCancel(result.id, e)}
                    title="Cancel entry"
                    className="p-1 rounded hover:bg-warning/20 text-muted-foreground hover:text-warning"
                  >
                    <X className="w-4 h-4" />
                  </button>
                )}
                {result.type === 'cancelled' && (
                  <button
                    onClick={(e) => handleUncancel(result.id, e)}
                    title="Uncancel entry"
                    className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary"
                  >
                    <RotateCcw className="w-4 h-4" />
                  </button>
                )}
                <button
                  onClick={(e) => handleCyclePriority(result.id, e)}
                  title="Cycle priority"
                  className="p-1 rounded hover:bg-warning/20 text-muted-foreground hover:text-warning"
                >
                  <span className="w-4 h-4 inline-flex items-center justify-center text-sm font-bold leading-none">!</span>
                </button>
                {result.type === 'task' && (
                  <button
                    onClick={(e) => e.stopPropagation()}
                    title="Migrate entry"
                    className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary"
                  >
                    <ArrowRight className="w-4 h-4" />
                  </button>
                )}
                <button
                  onClick={(e) => e.stopPropagation()}
                  title="Edit entry"
                  className="p-1 rounded hover:bg-secondary text-muted-foreground hover:text-foreground"
                >
                  <Pencil className="w-4 h-4" />
                </button>
                <button
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
