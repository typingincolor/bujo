import { describe, it, expect } from 'vitest';
import { filterWeekEntries, flattenEntries } from './weekView';
import { Entry } from '@/types/bujo';

describe('filterWeekEntries', () => {
  it('includes parent entries (parentId is null)', () => {
    const entries: Entry[] = [
      {
        id: 1,
        content: 'Top-level task',
        type: 'task',
        priority: 'none',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [],
      },
    ];

    const filtered = filterWeekEntries(entries);

    expect(filtered).toHaveLength(1);
    expect(filtered[0].content).toBe('Top-level task');
  });

  it('excludes child entries (parentId is set)', () => {
    const entries: Entry[] = [
      {
        id: 1,
        content: 'Parent',
        type: 'task',
        priority: 'none',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [],
      },
      {
        id: 2,
        content: 'Child',
        type: 'task',
        priority: 'high',
        parentId: 1,
        loggedDate: '2026-01-20',
        children: [],
      },
    ];

    const filtered = filterWeekEntries(entries);

    expect(filtered).toHaveLength(1);
    expect(filtered[0].content).toBe('Parent');
  });

  it('includes all entry types when they are parents', () => {
    const entries: Entry[] = [
      {
        id: 1,
        content: 'A task',
        type: 'task',
        priority: 'none',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [],
      },
      {
        id: 2,
        content: 'A note',
        type: 'note',
        priority: 'none',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [],
      },
      {
        id: 3,
        content: 'An event',
        type: 'event',
        priority: 'none',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [],
      },
      {
        id: 4,
        content: 'A question',
        type: 'question',
        priority: 'none',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [],
      },
    ];

    const filtered = filterWeekEntries(entries);

    expect(filtered).toHaveLength(4);
  });

  it('handles empty array', () => {
    expect(filterWeekEntries([])).toEqual([]);
  });
});

describe('flattenEntries', () => {
  it('returns empty array for empty input', () => {
    expect(flattenEntries([])).toEqual([]);
  });

  it('flattens deeply nested entries', () => {
    const entries: Entry[] = [
      {
        id: 1,
        content: 'Level 1',
        type: 'task',
        priority: 'none',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [
          {
            id: 2,
            content: 'Level 2',
            type: 'task',
            priority: 'none',
            parentId: 1,
            loggedDate: '2026-01-20',
            children: [
              {
                id: 3,
                content: 'Level 3',
                type: 'event',
                priority: 'none',
                parentId: 2,
                loggedDate: '2026-01-20',
                children: [],
              },
            ],
          },
        ],
      },
    ];

    const flattened = flattenEntries(entries);

    expect(flattened).toHaveLength(3);
    expect(flattened[0].content).toBe('Level 1');
    expect(flattened[1].content).toBe('Level 2');
    expect(flattened[2].content).toBe('Level 3');
  });
});
