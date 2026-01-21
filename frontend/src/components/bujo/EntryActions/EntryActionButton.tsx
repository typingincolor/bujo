import { cn } from '@/lib/utils';
import { ActionConfig } from './types';

export type ActionButtonSize = 'sm' | 'md';

interface EntryActionButtonProps {
  config: ActionConfig;
  onClick: () => void;
  size?: ActionButtonSize;
  className?: string;
}

const iconSizeClasses: Record<ActionButtonSize, string> = {
  sm: 'w-3.5 h-3.5',
  md: 'w-4 h-4',
};

export function EntryActionButton({
  config,
  onClick,
  size = 'md',
  className,
}: EntryActionButtonProps) {
  const Icon = config.icon;

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    onClick();
  };

  return (
    <button
      type="button"
      data-action-slot
      onClick={handleClick}
      title={config.title}
      className={cn(
        'p-1 rounded text-muted-foreground transition-colors',
        config.hoverClass,
        className
      )}
    >
      <Icon className={iconSizeClasses[size]} />
    </button>
  );
}
