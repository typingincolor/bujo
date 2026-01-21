import { Entry } from '@/types/bujo';
import {
  ActionConfig,
  ActionContext,
  EntryLike,
  EntryActionType,
  ACTION_REGISTRY,
  BAR_ACTION_ORDER,
  MENU_ACTION_ORDER,
} from './types';

export function getApplicableBarActions(
  entry: Entry | EntryLike,
  context?: ActionContext
): ActionConfig[] {
  return BAR_ACTION_ORDER
    .map(type => ACTION_REGISTRY[type])
    .filter(config => config.showInBar && config.appliesTo(entry, context));
}

export function getApplicableMenuActions(
  entry: Entry | EntryLike,
  context?: ActionContext
): ActionConfig[] {
  return MENU_ACTION_ORDER
    .map(type => ACTION_REGISTRY[type])
    .filter(config => config.showInMenu && config.appliesTo(entry, context));
}

export function getActionConfig(type: EntryActionType): ActionConfig {
  return ACTION_REGISTRY[type];
}

export function isActionApplicable(
  type: EntryActionType,
  entry: Entry | EntryLike,
  context?: ActionContext
): boolean {
  return ACTION_REGISTRY[type].appliesTo(entry, context);
}
