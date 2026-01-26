/**
 * Entry Action Validation Logic
 *
 * IMPORTANT: The validation rules in this file (appliesTo functions) must stay
 * in sync with the domain layer source of truth: internal/domain/entry.go
 *
 * The Go domain layer defines Can* methods on the Entry type:
 * - CanAnswer() -> answer.appliesTo
 * - CanCancel() -> cancel.appliesTo
 * - CanUncancel() -> uncancel.appliesTo
 * - CanCycleType() -> cycleType.appliesTo
 * - CanEdit() -> edit.appliesTo
 * - CanMigrate() -> migrate.appliesTo
 * - CanAddChild() -> addChild.appliesTo
 * - CanMoveToList() -> moveToList.appliesTo
 * - CanMoveToRoot() -> moveToRoot.appliesTo
 * - CanCyclePriority() -> cyclePriority.appliesTo
 * - CanDelete() -> delete.appliesTo
 *
 * This duplication exists because:
 * 1. Different tech stacks (Go backend, TypeScript frontend)
 * 2. Frontend needs instant validation for UX (hiding/showing buttons)
 * 3. Backend validates regardless (defense-in-depth)
 *
 * When modifying validation rules, update BOTH files.
 */
import { Entry, EntryType } from '@/types/bujo';
import { LucideIcon, MessageCircle, X, RotateCcw, Flag, RefreshCw, ArrowRight, Pencil, Trash2, CornerDownRight, ArrowUpToLine, ListPlus, Calendar, Check } from 'lucide-react';

export type EntryActionType =
  | 'answer'
  | 'markDone'
  | 'cancel'
  | 'uncancel'
  | 'cyclePriority'
  | 'cycleType'
  | 'migrate'
  | 'edit'
  | 'delete'
  | 'addChild'
  | 'moveToRoot'
  | 'moveToList'
  | 'navigateToEntry';

export interface ActionContext {
  hasParent?: boolean;
}

export interface ActionConfig {
  type: EntryActionType;
  icon: LucideIcon;
  label: string;
  title: string;
  appliesTo: (entry: Entry | EntryLike, context?: ActionContext) => boolean;
  showInBar: boolean;
  showInMenu: boolean;
  hoverClass: string;
}

export interface EntryLike {
  id: number;
  type: EntryType;
  priority?: string;
}

export interface ActionCallbacks {
  onAnswer?: () => void;
  onMarkDone?: () => void;
  onCancel?: () => void;
  onUncancel?: () => void;
  onCyclePriority?: () => void;
  onCycleType?: () => void;
  onMigrate?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
  onAddChild?: () => void;
  onMoveToRoot?: () => void;
  onMoveToList?: () => void;
  onNavigateToEntry?: () => void;
}

const CYCLEABLE_TYPES: EntryType[] = ['task', 'note', 'event', 'question'];

export const ACTION_REGISTRY: Record<EntryActionType, ActionConfig> = {
  answer: {
    type: 'answer',
    icon: MessageCircle,
    label: 'Answer',
    title: 'Answer question',
    appliesTo: (entry) => entry.type === 'question',
    showInBar: true,
    showInMenu: true,
    hoverClass: 'hover:bg-bujo-question/20 hover:text-bujo-question',
  },
  markDone: {
    type: 'markDone',
    icon: Check,
    label: 'Mark done',
    title: 'Mark as done',
    appliesTo: (entry) => entry.type === 'task',
    showInBar: true,
    showInMenu: true,
    hoverClass: 'hover:bg-bujo-done/20 hover:text-bujo-done',
  },
  cancel: {
    type: 'cancel',
    icon: X,
    label: 'Cancel',
    title: 'Cancel entry',
    appliesTo: (entry) => entry.type !== 'cancelled',
    showInBar: true,
    showInMenu: true,
    hoverClass: 'hover:bg-warning/20 hover:text-warning',
  },
  uncancel: {
    type: 'uncancel',
    icon: RotateCcw,
    label: 'Uncancel',
    title: 'Uncancel entry',
    appliesTo: (entry) => entry.type === 'cancelled',
    showInBar: true,
    showInMenu: true,
    hoverClass: 'hover:bg-primary/20 hover:text-primary',
  },
  cyclePriority: {
    type: 'cyclePriority',
    icon: Flag,
    label: 'Cycle priority',
    title: 'Cycle priority',
    appliesTo: () => true,
    showInBar: true,
    showInMenu: true,
    hoverClass: 'hover:bg-warning/20 hover:text-warning',
  },
  cycleType: {
    type: 'cycleType',
    icon: RefreshCw,
    label: 'Change type',
    title: 'Change type',
    appliesTo: (entry) => CYCLEABLE_TYPES.includes(entry.type) && entry.type !== 'cancelled',
    showInBar: true,
    showInMenu: true,
    hoverClass: 'hover:bg-primary/20 hover:text-primary',
  },
  migrate: {
    type: 'migrate',
    icon: ArrowRight,
    label: 'Migrate',
    title: 'Migrate entry',
    appliesTo: (entry) => entry.type === 'task',
    showInBar: true,
    showInMenu: true,
    hoverClass: 'hover:bg-primary/20 hover:text-primary',
  },
  edit: {
    type: 'edit',
    icon: Pencil,
    label: 'Edit',
    title: 'Edit entry',
    appliesTo: (entry) => entry.type !== 'cancelled',
    showInBar: true,
    showInMenu: true,
    hoverClass: 'hover:bg-secondary hover:text-foreground',
  },
  delete: {
    type: 'delete',
    icon: Trash2,
    label: 'Delete',
    title: 'Delete entry',
    appliesTo: () => true,
    showInBar: true,
    showInMenu: true,
    hoverClass: 'hover:bg-destructive/20 hover:text-destructive',
  },
  addChild: {
    type: 'addChild',
    icon: CornerDownRight,
    label: 'Add child',
    title: 'Add child entry',
    appliesTo: (entry) => entry.type !== 'question',
    showInBar: false,
    showInMenu: true,
    hoverClass: 'hover:bg-secondary hover:text-foreground',
  },
  moveToRoot: {
    type: 'moveToRoot',
    icon: ArrowUpToLine,
    label: 'Move to root',
    title: 'Move to root level',
    appliesTo: (_, context) => context?.hasParent === true,
    showInBar: false,
    showInMenu: true,
    hoverClass: 'hover:bg-secondary hover:text-foreground',
  },
  moveToList: {
    type: 'moveToList',
    icon: ListPlus,
    label: 'Move to list',
    title: 'Move to list',
    appliesTo: (entry) => entry.type === 'task',
    showInBar: true,
    showInMenu: true,
    hoverClass: 'hover:bg-primary/20 hover:text-primary',
  },
  navigateToEntry: {
    type: 'navigateToEntry',
    icon: Calendar,
    label: 'Go to date',
    title: 'Go to date',
    appliesTo: () => true,
    showInBar: true,
    showInMenu: true,
    hoverClass: 'hover:bg-primary/20 hover:text-primary',
  },
};

export const BAR_ACTION_ORDER: EntryActionType[] = [
  'answer',
  'markDone',
  'cancel',
  'uncancel',
  'cyclePriority',
  'cycleType',
  'migrate',
  'moveToList',
  'navigateToEntry',
  'edit',
  'delete',
];

export const MENU_ACTION_ORDER: EntryActionType[] = [
  'answer',
  'markDone',
  'cancel',
  'uncancel',
  'migrate',
  'moveToList',
  'navigateToEntry',
  'cycleType',
  'cyclePriority',
  'moveToRoot',
  'addChild',
  'edit',
  'delete',
];
