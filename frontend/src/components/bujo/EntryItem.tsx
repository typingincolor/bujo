import { Entry, EntryType } from '@/types/bujo';
import { EntrySymbol } from './EntrySymbol';
import { cn } from '@/lib/utils';
import { ChevronRight, ChevronDown, Pencil, Trash2, MessageCircle, X, RotateCcw, AlertTriangle, ArrowRight, Check } from 'lucide-react';

interface EntryItemProps {
  entry: Entry;
  depth?: number;
  isCollapsed?: boolean;
  hasChildren?: boolean;
  childCount?: number;
  isSelected?: boolean;
  onToggleCollapse?: () => void;
  onToggleDone?: () => void;
  onSelect?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
  onAnswer?: () => void;
  onCancel?: () => void;
  onUncancel?: () => void;
  onCyclePriority?: () => void;
  onMigrate?: () => void;
}

const contentStyles: Record<EntryType, string> = {
  task: '',
  note: 'text-muted-foreground italic',
  event: 'font-medium',
  done: 'line-through text-muted-foreground',
  migrated: 'text-muted-foreground',
  cancelled: 'line-through text-muted-foreground opacity-60',
  question: 'text-bujo-question font-medium',
  answered: 'text-bujo-answered',
  answer: 'text-muted-foreground pl-4',
};

export function EntryItem({
  entry,
  depth = 0,
  isCollapsed = false,
  hasChildren = false,
  childCount = 0,
  isSelected = false,
  onToggleCollapse,
  onToggleDone,
  onSelect,
  onEdit,
  onDelete,
  onAnswer,
  onCancel,
  onUncancel,
  onCyclePriority,
  onMigrate,
}: EntryItemProps) {
  const isToggleable = entry.type === 'task' || entry.type === 'done';

  const handleClick = () => {
    onSelect?.();
  };

  return (
    <div
      data-entry-id={entry.id}
      data-selected={isSelected}
      className={cn(
        'group flex items-center gap-2 py-1.5 px-2 rounded-md transition-colors',
        'animate-fade-in',
        !isSelected && 'hover:bg-secondary/50',
        isToggleable && 'cursor-pointer',
        isSelected && 'bg-primary/10 ring-1 ring-primary/30'
      )}
      style={{ paddingLeft: `${depth * 20 + 8}px` }}
      onClick={handleClick}
    >
      {/* Collapse indicator */}
      {hasChildren ? (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onToggleCollapse?.();
          }}
          className="w-4 h-4 flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors"
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

      {/* Action buttons (shown on hover) */}
      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
        {onToggleDone && entry.type === 'task' && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onToggleDone();
            }}
            title="Mark as done"
            className="p-1 rounded hover:bg-green-500/20 text-muted-foreground hover:text-green-600 transition-colors"
          >
            <Check className="w-3.5 h-3.5" />
          </button>
        )}
        {onToggleDone && entry.type === 'done' && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onToggleDone();
            }}
            title="Mark as not done"
            className="p-1 rounded hover:bg-orange-500/20 text-muted-foreground hover:text-orange-600 transition-colors"
          >
            <span className="text-sm font-bold leading-none">•</span>
          </button>
        )}
        {onAnswer && entry.type === 'question' && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onAnswer();
            }}
            title="Answer question"
            className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary transition-colors"
          >
            <MessageCircle className="w-3.5 h-3.5" />
          </button>
        )}
        {onCancel && entry.type !== 'cancelled' && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onCancel();
            }}
            title="Cancel entry"
            className="p-1 rounded hover:bg-warning/20 text-muted-foreground hover:text-warning transition-colors"
          >
            <X className="w-3.5 h-3.5" />
          </button>
        )}
        {onUncancel && entry.type === 'cancelled' && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onUncancel();
            }}
            title="Uncancel entry"
            className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary transition-colors"
          >
            <RotateCcw className="w-3.5 h-3.5" />
          </button>
        )}
        {onCyclePriority && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onCyclePriority();
            }}
            title="Cycle priority"
            className="p-1 rounded hover:bg-warning/20 text-muted-foreground hover:text-warning transition-colors"
          >
            <AlertTriangle className="w-3.5 h-3.5" />
          </button>
        )}
        {onMigrate && entry.type === 'task' && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onMigrate();
            }}
            title="Migrate entry"
            className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary transition-colors"
          >
            <ArrowRight className="w-3.5 h-3.5" />
          </button>
        )}
        {onEdit && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onEdit();
            }}
            title="Edit entry"
            className="p-1 rounded hover:bg-secondary text-muted-foreground hover:text-foreground transition-colors"
          >
            <Pencil className="w-3.5 h-3.5" />
          </button>
        )}
        {onDelete && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onDelete();
            }}
            title="Delete entry"
            className="p-1 rounded hover:bg-destructive/20 text-muted-foreground hover:text-destructive transition-colors"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        )}
      </div>

      {/* Entry ID (shown on hover) */}
      <span className="text-xs text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity">
        #{entry.id}
      </span>
    </div>
  );
}
