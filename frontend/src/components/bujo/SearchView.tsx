import { useState, useCallback } from 'react';
import { Search as SearchIcon } from 'lucide-react';
import { Search } from '@/wailsjs/go/wails/App';
import { format } from 'date-fns';
import { cn } from '@/lib/utils';
import { ENTRY_SYMBOLS, EntryType } from '@/types/bujo';

interface SearchResult {
  id: number;
  content: string;
  type: EntryType;
  date: string;
}

export function SearchView() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [hasSearched, setHasSearched] = useState(false);

  const handleSearch = useCallback(async (searchQuery: string) => {
    setQuery(searchQuery);

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
      })));
      setHasSearched(true);
    } catch (error) {
      console.error('Search failed:', error);
      setResults([]);
      setHasSearched(true);
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

        {results.map((result) => (
          <div
            key={result.id}
            className={cn(
              'flex items-start gap-3 p-3 rounded-lg border border-border',
              'bg-card hover:bg-secondary/30 transition-colors group'
            )}
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
            <span className="text-xs text-muted-foreground opacity-0 group-hover:opacity-100">
              #{result.id}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
