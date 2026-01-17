import { useState, useEffect, useRef } from 'react';
import { format } from 'date-fns';
import { Calendar, Search, Command } from 'lucide-react';
import { cn } from '@/lib/utils';

const SEARCH_DEBOUNCE_MS = 200;

interface SearchResult {
  id: number;
  content: string;
  type: string;
  date: string;
}

interface HeaderProps {
  title: string;
  searchResults?: SearchResult[];
  onSearch?: (query: string) => void;
  onSelectResult?: (result: SearchResult) => void;
}

export function Header({ title, searchResults = [], onSearch, onSelectResult }: HeaderProps) {
  const today = new Date();
  const [query, setQuery] = useState('');
  const [showResults, setShowResults] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const searchRequestRef = useRef(0);

  useEffect(() => {
    const currentRequest = ++searchRequestRef.current;

    const debounceTimer = setTimeout(() => {
      if (currentRequest !== searchRequestRef.current) return;

      if (query.length > 0) {
        onSearch?.(query);
        setShowResults(true);
      } else {
        setShowResults(false);
      }
    }, SEARCH_DEBOUNCE_MS);

    return () => clearTimeout(debounceTimer);
  }, [query, onSearch]);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node) &&
          inputRef.current && !inputRef.current.contains(e.target as Node)) {
        setShowResults(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleResultClick = (result: SearchResult) => {
    onSelectResult?.(result);
    setQuery('');
    setShowResults(false);
  };

  return (
    <header className="flex items-center justify-between px-6 py-4 border-b border-border bg-card/50">
      <div className="flex items-center gap-4">
        <h2 className="font-display text-2xl font-semibold">{title}</h2>
        <span className="flex items-center gap-1.5 text-sm text-muted-foreground">
          <Calendar className="w-4 h-4" />
          {format(today, 'EEEE, MMMM d, yyyy')}
        </span>
      </div>

      <div className="flex items-center gap-3">
        {/* Search */}
        <div className="relative">
          <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search entries..."
            className={cn(
              'pl-9 pr-4 py-2 w-64 rounded-lg text-sm',
              'bg-secondary/50 border border-transparent',
              'placeholder:text-muted-foreground',
              'focus:bg-background focus:border-border focus:outline-none focus:ring-1 focus:ring-ring',
              'transition-all'
            )}
          />
          <div className="absolute right-3 top-1/2 -translate-y-1/2 flex items-center gap-0.5 text-xs text-muted-foreground">
            <Command className="w-3 h-3" />
            <span>K</span>
          </div>

          {/* Search Results Dropdown */}
          {showResults && searchResults.length > 0 && (
            <div
              ref={dropdownRef}
              className="absolute top-full left-0 right-0 mt-1 bg-card border border-border rounded-lg shadow-lg z-50 max-h-64 overflow-y-auto"
            >
              {searchResults.map((result) => (
                <button
                  key={result.id}
                  onClick={() => handleResultClick(result)}
                  className="w-full px-3 py-2 text-left text-sm hover:bg-secondary/50 transition-colors flex items-center justify-between"
                >
                  <span className="truncate">{result.content}</span>
                  <span className="text-xs text-muted-foreground ml-2">{result.date}</span>
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
