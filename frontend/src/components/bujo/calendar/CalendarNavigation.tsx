import { ChevronLeft, ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface CalendarNavigationProps {
  label: string;
  onPrev: () => void;
  onNext: () => void;
  canGoPrev?: boolean;
  canGoNext?: boolean;
}

export function CalendarNavigation({
  label,
  onPrev,
  onNext,
  canGoPrev = true,
  canGoNext = true,
}: CalendarNavigationProps) {
  return (
    <div className="flex items-center justify-between gap-2">
      <button
        onClick={onPrev}
        disabled={!canGoPrev}
        aria-label="Previous"
        className={cn(
          'p-1 rounded-md transition-colors',
          canGoPrev
            ? 'hover:bg-secondary text-foreground'
            : 'text-muted-foreground cursor-not-allowed opacity-50'
        )}
      >
        <ChevronLeft className="w-4 h-4" />
      </button>

      <span className="text-sm font-medium">{label}</span>

      <button
        onClick={onNext}
        disabled={!canGoNext}
        aria-label="Next"
        className={cn(
          'p-1 rounded-md transition-colors',
          canGoNext
            ? 'hover:bg-secondary text-foreground'
            : 'text-muted-foreground cursor-not-allowed opacity-50'
        )}
      >
        <ChevronRight className="w-4 h-4" />
      </button>
    </div>
  );
}
