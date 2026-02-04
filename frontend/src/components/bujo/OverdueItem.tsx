import * as Tooltip from '@radix-ui/react-tooltip';
import { Entry, ENTRY_SYMBOLS, PRIORITY_SYMBOLS } from '@/types/bujo';
import { AttentionScore } from '@/hooks/useAttentionScores';
import { cn } from '@/lib/utils';

interface OverdueItemProps {
  entry: Entry;
  attentionScore?: AttentionScore;
  breadcrumb?: string;
  onSelect?: (entry: Entry) => void;
  isSelected?: boolean;
}

function getBadgeColor(score: number): string {
  if (score >= 80) return 'bg-red-500';
  if (score >= 50) return 'bg-orange-500';
  return 'bg-yellow-500';
}

function formatIndicator(indicator: string): string {
  switch (indicator) {
    case 'overdue':
      return 'Overdue';
    case 'priority':
      return 'Priority';
    case 'aging':
      return 'Aging';
    case 'migrated':
      return 'Migrated';
    default:
      return indicator;
  }
}

export function OverdueItem({
  entry,
  attentionScore,
  breadcrumb,
  onSelect,
  isSelected = false,
}: OverdueItemProps) {
  const score = attentionScore?.score ?? 0;
  const indicators = attentionScore?.indicators ?? [];
  const symbol = ENTRY_SYMBOLS[entry.type];
  const prioritySymbol = PRIORITY_SYMBOLS[entry.priority];
  const hasParent = entry.parentId !== null;

  const handleClick = () => {
    onSelect?.(entry);
  };

  return (
    <Tooltip.Provider>
      <button
        onClick={handleClick}
        className={cn(
          'w-full flex items-center gap-2 px-2 py-1.5 rounded-lg text-left text-sm transition-colors hover:bg-secondary/50',
          isSelected && 'bg-primary/10 ring-1 ring-primary/30'
        )}
      >
        <span
          data-testid="context-dot-container"
          className="w-2 flex-shrink-0 flex items-center justify-center"
        >
          {hasParent && (
            <span
              data-testid="context-dot"
              className="h-1.5 w-1.5 rounded-full bg-muted-foreground/50"
            />
          )}
        </span>

        <span data-testid="entry-symbol" className="text-muted-foreground flex-shrink-0">
          {symbol}
        </span>

        {prioritySymbol && (
          <span
            data-testid="priority-indicator"
            className="text-orange-500 font-medium flex-shrink-0"
          >
            {prioritySymbol}
          </span>
        )}

        <span className="flex-1 truncate">{entry.content}</span>

        {breadcrumb && (
          <span
            data-testid="breadcrumb"
            className="text-xs text-muted-foreground truncate flex-shrink-0 max-w-[120px]"
          >
            {breadcrumb}
          </span>
        )}

        <Tooltip.Root>
          <Tooltip.Trigger asChild>
            <span
              data-testid="attention-badge"
              className={cn(
                'px-1.5 py-0.5 rounded text-xs font-medium text-white flex-shrink-0',
                getBadgeColor(score)
              )}
            >
              {score}
            </span>
          </Tooltip.Trigger>
          <Tooltip.Portal>
            <Tooltip.Content
              role="tooltip"
              className="z-50 bg-popover border border-border rounded-lg shadow-lg p-2 text-sm"
              sideOffset={5}
            >
              <div className="font-medium mb-1">Attention Score: {score}</div>
              <ul className="text-muted-foreground text-xs space-y-0.5">
                {indicators.map((indicator) => (
                  <li key={indicator}>{formatIndicator(indicator)}</li>
                ))}
              </ul>
              <Tooltip.Arrow className="fill-popover" />
            </Tooltip.Content>
          </Tooltip.Portal>
        </Tooltip.Root>
      </button>
    </Tooltip.Provider>
  );
}
