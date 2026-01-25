import { describe, it, expect } from 'vitest';
import { filterWeekEntries, flattenEntries } from './weekView';
import { Entry } from '@/types/bujo';

describe('filterWeekEntries', () => {
  it('includes events', () => {
    const entries: Entry[] = [
      {
        id: 1,
        content: 'Team meeting',
        type: 'event',
        priority: 'none',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [],
      },
    ];

    const filtered = filterWeekEntries(entries);

    expect(filtered).toHaveLength(1);
    expect(filtered[0].content).toBe('Team meeting');
  });

  it('includes priority entries', () => {
    const entries: Entry[] = [
      {
        id: 2,
        content: 'Fix critical bug',
        type: 'task',
        priority: 'high',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [],
      },
    ];

    const filtered = filterWeekEntries(entries);

    expect(filtered).toHaveLength(1);
    expect(filtered[0].content).toBe('Fix critical bug');
  });

  it('excludes non-priority tasks', () => {
    const entries: Entry[] = [
      {
        id: 3,
        content: 'Regular task',
        type: 'task',
        priority: 'none',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [],
      },
    ];

    const filtered = filterWeekEntries(entries);

    expect(filtered).toHaveLength(0);
  });

  it('flattens nested entries', () => {
    const entries: Entry[] = [
      {
        id: 1,
        content: 'Parent',
        type: 'task',
        priority: 'none',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [
          {
            id: 2,
            content: 'Child event',
            type: 'event',
            priority: 'none',
            parentId: 1,
            loggedDate: '2026-01-20',
            children: [],
          },
        ],
      },
    ];

    const filtered = filterWeekEntries(entries);

    expect(filtered).toHaveLength(1);
    expect(filtered[0].content).toBe('Child event');
  });
});
