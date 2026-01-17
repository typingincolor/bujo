import { format } from 'date-fns';
import { Calendar, Search, Command } from 'lucide-react';
import { cn } from '@/lib/utils';

interface HeaderProps {
  title: string;
}

export function Header({ title }: HeaderProps) {
  const today = new Date();
  
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
            type="text"
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
        </div>
      </div>
    </header>
  );
}
