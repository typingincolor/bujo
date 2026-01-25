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

describe('filterWeekEntries - edge cases', () => {
  it('handles empty array', () => {
    expect(filterWeekEntries([])).toEqual([]);
  });

  it('includes all priority levels', () => {
    const entries: Entry[] = [
      {
        id: 1,
        content: 'Low priority',
        type: 'task',
        priority: 'low',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [],
      },
      {
        id: 2,
        content: 'Medium priority',
        type: 'task',
        priority: 'medium',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [],
      },
      {
        id: 3,
        content: 'High priority',
        type: 'task',
        priority: 'high',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [],
      },
    ];

    const filtered = filterWeekEntries(entries);

    expect(filtered).toHaveLength(3);
  });

  it('handles multiple qualifying entries', () => {
    const entries: Entry[] = [
      {
        id: 1,
        content: 'Event',
        type: 'event',
        priority: 'none',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [],
      },
      {
        id: 2,
        content: 'Priority task',
        type: 'task',
        priority: 'high',
        parentId: null,
        loggedDate: '2026-01-20',
        children: [],
      },
    ];

    const filtered = filterWeekEntries(entries);

    expect(filtered).toHaveLength(2);
  });
});
