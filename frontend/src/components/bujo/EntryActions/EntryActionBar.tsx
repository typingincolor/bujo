import { cn } from '@/lib/utils';
import { Entry } from '@/types/bujo';
import { EntryActionButton, ActionButtonSize } from './EntryActionButton';
import { ActionPlaceholder } from './ActionPlaceholder';
import {
  ActionCallbacks,
  ActionConfig,
  ActionContext,
  EntryLike,
  BAR_ACTION_ORDER,
  ACTION_REGISTRY,
  EntryActionType,
} from './types';

export type EntryActionBarVariant = 'hover-reveal' | 'always-visible';

interface EntryActionBarProps {
  entry: Entry | EntryLike;
  callbacks: ActionCallbacks;
  context?: ActionContext;
  variant?: EntryActionBarVariant;
  size?: ActionButtonSize;
  usePlaceholders?: boolean;
  isSelected?: boolean;
  className?: string;
}

const callbackMap: Record<EntryActionType, keyof ActionCallbacks> = {
  answer: 'onAnswer',
  cancel: 'onCancel',
  uncancel: 'onUncancel',
  cyclePriority: 'onCyclePriority',
  cycleType: 'onCycleType',
  migrate: 'onMigrate',
  edit: 'onEdit',
  delete: 'onDelete',
  addChild: 'onAddChild',
  moveToRoot: 'onMoveToRoot',
  moveToList: 'onMoveToList',
  navigateToEntry: 'onNavigateToEntry',
};

export function EntryActionBar({
  entry,
  callbacks,
  context,
  variant = 'hover-reveal',
  size = 'md',
  usePlaceholders = false,
  isSelected = false,
  className,
}: EntryActionBarProps) {
  const variantClasses =
    variant === 'hover-reveal' && !isSelected
      ? 'opacity-0 group-hover:opacity-100 focus-within:opacity-100 transition-opacity'
      : '';

  const renderAction = (actionType: EntryActionType): React.ReactNode => {
    const config: ActionConfig = ACTION_REGISTRY[actionType];
    const callbackKey = callbackMap[actionType];
    const callback = callbacks[callbackKey];

    const isApplicable = config.showInBar && config.appliesTo(entry, context);
    const hasCallback = callback !== undefined;

    if (isApplicable && hasCallback) {
      return (
        <EntryActionButton
          key={actionType}
          config={config}
          onClick={callback}
          size={size}
        />
      );
    }

    if (usePlaceholders) {
      return <ActionPlaceholder key={actionType} size={size} />;
    }

    return null;
  };

  return (
    <div
      data-testid="entry-action-bar"
      className={cn('flex items-center gap-0.5', variantClasses, className)}
    >
      {BAR_ACTION_ORDER.map(renderAction)}
    </div>
  );
}
