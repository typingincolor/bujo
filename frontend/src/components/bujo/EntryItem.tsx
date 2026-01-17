import { Entry, EntryType } from '@/types/bujo';
import { EntrySymbol } from './EntrySymbol';
import { cn } from '@/lib/utils';
import { ChevronRight, ChevronDown } from 'lucide-react';

interface EntryItemProps {
  entry: Entry;
  depth?: number;
  isCollapsed?: boolean;
  hasChildren?: boolean;
  childCount?: number;
  onToggleCollapse?: () => void;
  onToggleDone?: () => void;
}

const contentStyles: Record<EntryType, string> = {
  task: '',
  note: 'text-muted-foreground italic',
  event: 'font-medium',
  done: 'line-through text-muted-foreground',
  migrated: 'text-muted-foreground',
  cancelled: 'line-through text-muted-foreground opacity-60',
};

export function EntryItem({
  entry,
  depth = 0,
  isCollapsed = false,
  hasChildren = false,
  childCount = 0,
  onToggleCollapse,
  onToggleDone,
}: EntryItemProps) {
  return (
    <div
      className={cn(
        'group flex items-start gap-2 py-1.5 px-2 rounded-md transition-colors',
        'hover:bg-secondary/50 cursor-pointer animate-fade-in'
      )}
      style={{ paddingLeft: `${depth * 20 + 8}px` }}
      onClick={onToggleDone}
    >
      {/* Collapse indicator */}
      {hasChildren ? (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onToggleCollapse?.();
          }}
          className="w-4 h-4 mt-0.5 flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors"
        >
          {isCollapsed ? (
            <ChevronRight className="w-3.5 h-3.5" />
          ) : (
            <ChevronDown className="w-3.5 h-3.5" />
          )}
        </button>
      ) : (
        <span className="w-4" />
      )}

      {/* Symbol */}
      <EntrySymbol type={entry.type} priority={entry.priority} />

      {/* Content */}
      <span className={cn('flex-1 text-sm', contentStyles[entry.type])}>
        {entry.content}
      </span>

      {/* Hidden child count when collapsed */}
      {isCollapsed && childCount > 0 && (
        <span className="text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
          {childCount} hidden
        </span>
      )}

      {/* Entry ID (shown on hover) */}
      <span className="text-xs text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity">
        #{entry.id}
      </span>
    </div>
  );
}
