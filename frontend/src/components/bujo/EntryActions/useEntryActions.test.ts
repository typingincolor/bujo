import { describe, it, expect } from 'vitest';
import { getApplicableBarActions, getApplicableMenuActions } from './useEntryActions';
import { EntryLike, ActionContext } from './types';

describe('useEntryActions', () => {
  describe('getApplicableBarActions', () => {
    it('returns answer action only for question entries', () => {
      const question: EntryLike = { id: 1, type: 'question' };
      const task: EntryLike = { id: 2, type: 'task' };

      const questionActions = getApplicableBarActions(question);
      const taskActions = getApplicableBarActions(task);

      expect(questionActions.some(a => a.type === 'answer')).toBe(true);
      expect(taskActions.some(a => a.type === 'answer')).toBe(false);
    });

    it('returns cancel for non-cancelled entries', () => {
      const task: EntryLike = { id: 1, type: 'task' };
      const cancelled: EntryLike = { id: 2, type: 'cancelled' };

      const taskActions = getApplicableBarActions(task);
      const cancelledActions = getApplicableBarActions(cancelled);

      expect(taskActions.some(a => a.type === 'cancel')).toBe(true);
      expect(cancelledActions.some(a => a.type === 'cancel')).toBe(false);
    });

    it('returns uncancel only for cancelled entries', () => {
      const task: EntryLike = { id: 1, type: 'task' };
      const cancelled: EntryLike = { id: 2, type: 'cancelled' };

      const taskActions = getApplicableBarActions(task);
      const cancelledActions = getApplicableBarActions(cancelled);

      expect(taskActions.some(a => a.type === 'uncancel')).toBe(false);
      expect(cancelledActions.some(a => a.type === 'uncancel')).toBe(true);
    });

    it('returns cyclePriority for all entry types', () => {
      const types = ['task', 'note', 'event', 'done', 'cancelled', 'question'] as const;

      for (const type of types) {
        const entry: EntryLike = { id: 1, type };
        const actions = getApplicableBarActions(entry);
        expect(actions.some(a => a.type === 'cyclePriority')).toBe(true);
      }
    });

    it('returns cycleType only for task, note, event, question', () => {
      const allowed: EntryLike[] = [
        { id: 1, type: 'task' },
        { id: 2, type: 'note' },
        { id: 3, type: 'event' },
        { id: 4, type: 'question' },
      ];
      const notAllowed: EntryLike[] = [
        { id: 5, type: 'done' },
        { id: 6, type: 'migrated' },
        { id: 7, type: 'cancelled' },
        { id: 8, type: 'answered' },
        { id: 9, type: 'answer' },
      ];

      for (const entry of allowed) {
        const actions = getApplicableBarActions(entry);
        expect(actions.some(a => a.type === 'cycleType')).toBe(true);
      }

      for (const entry of notAllowed) {
        const actions = getApplicableBarActions(entry);
        expect(actions.some(a => a.type === 'cycleType')).toBe(false);
      }
    });

    it('returns migrate only for task entries', () => {
      const task: EntryLike = { id: 1, type: 'task' };
      const note: EntryLike = { id: 2, type: 'note' };

      const taskActions = getApplicableBarActions(task);
      const noteActions = getApplicableBarActions(note);

      expect(taskActions.some(a => a.type === 'migrate')).toBe(true);
      expect(noteActions.some(a => a.type === 'migrate')).toBe(false);
    });

    it('returns edit for non-cancelled entries', () => {
      const task: EntryLike = { id: 1, type: 'task' };
      const cancelled: EntryLike = { id: 2, type: 'cancelled' };

      const taskActions = getApplicableBarActions(task);
      const cancelledActions = getApplicableBarActions(cancelled);

      expect(taskActions.some(a => a.type === 'edit')).toBe(true);
      expect(cancelledActions.some(a => a.type === 'edit')).toBe(false);
    });

    it('returns delete for all entry types', () => {
      const types = ['task', 'note', 'event', 'done', 'cancelled', 'question'] as const;

      for (const type of types) {
        const entry: EntryLike = { id: 1, type };
        const actions = getApplicableBarActions(entry);
        expect(actions.some(a => a.type === 'delete')).toBe(true);
      }
    });

    it('returns actions in correct order', () => {
      const task: EntryLike = { id: 1, type: 'task' };
      const actions = getApplicableBarActions(task);
      const types = actions.map(a => a.type);

      const expectedOrder = ['markDone', 'cancel', 'cyclePriority', 'cycleType', 'migrate', 'moveToList', 'navigateToEntry', 'edit', 'delete'];
      expect(types).toEqual(expectedOrder);
    });

    it('returns actions in correct order for question', () => {
      const question: EntryLike = { id: 1, type: 'question' };
      const actions = getApplicableBarActions(question);
      const types = actions.map(a => a.type);

      const expectedOrder = ['answer', 'cancel', 'cyclePriority', 'cycleType', 'navigateToEntry', 'edit', 'delete'];
      expect(types).toEqual(expectedOrder);
    });
  });

  describe('getApplicableMenuActions', () => {
    it('includes addChild for non-question entries', () => {
      const task: EntryLike = { id: 1, type: 'task' };
      const question: EntryLike = { id: 2, type: 'question' };

      const taskActions = getApplicableMenuActions(task);
      const questionActions = getApplicableMenuActions(question);

      expect(taskActions.some(a => a.type === 'addChild')).toBe(true);
      expect(questionActions.some(a => a.type === 'addChild')).toBe(false);
    });

    it('includes moveToRoot only when hasParent is true', () => {
      const entry: EntryLike = { id: 1, type: 'task' };
      const withParent: ActionContext = { hasParent: true };
      const withoutParent: ActionContext = { hasParent: false };

      const actionsWithParent = getApplicableMenuActions(entry, withParent);
      const actionsWithoutParent = getApplicableMenuActions(entry, withoutParent);

      expect(actionsWithParent.some(a => a.type === 'moveToRoot')).toBe(true);
      expect(actionsWithoutParent.some(a => a.type === 'moveToRoot')).toBe(false);
    });

    it('includes moveToList only for task entries', () => {
      const task: EntryLike = { id: 1, type: 'task' };
      const note: EntryLike = { id: 2, type: 'note' };

      const taskActions = getApplicableMenuActions(task);
      const noteActions = getApplicableMenuActions(note);

      expect(taskActions.some(a => a.type === 'moveToList')).toBe(true);
      expect(noteActions.some(a => a.type === 'moveToList')).toBe(false);
    });
  });
});
