import { cn } from '@/lib/utils';
import { priorityColors } from './insights-constants';

export function LevelBadge({ level }: { level: string }) {
  return (
    <span className={cn('px-1.5 py-0.5 rounded text-xs', priorityColors[level] || priorityColors.low)}>
      {level}
    </span>
  );
}
