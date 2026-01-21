import { ActionButtonSize } from './EntryActionButton';

interface ActionPlaceholderProps {
  size?: ActionButtonSize;
}

const sizeClasses: Record<ActionButtonSize, string> = {
  sm: 'p-1 w-[22px] h-[22px]',
  md: 'p-1 w-6 h-6',
};

export function ActionPlaceholder({ size = 'md' }: ActionPlaceholderProps) {
  return <span data-action-slot className={sizeClasses[size]} aria-hidden="true" />;
}
