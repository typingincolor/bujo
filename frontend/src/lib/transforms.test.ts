import { describe, it, expect } from 'vitest'
import {
  transformEntry,
  transformDayEntries,
  transformHabit,
  transformList,
  transformGoal,
} from './transforms'
import { domain, service, wails } from '../wailsjs/go/models'

describe('transformEntry', () => {
  it('transforms a task entry correctly', () => {
    const input = {
      ID: 1,
      EntityID: 'abc123',
      Type: 'Task',
      Content: 'Buy groceries',
      Priority: 'High',
      ParentID: undefined,
      Depth: 0,
      CreatedAt: '2026-01-15T10:00:00Z',
    } as unknown as domain.Entry

    const result = transformEntry(input)

    expect(result).toEqual({
      id: 1,
      content: 'Buy groceries',
      type: 'task',
      priority: 'high',
      parentId: null,
      loggedDate: '2026-01-15T10:00:00Z',
    })
  })

  it('handles null ParentID by converting to null', () => {
    const input = {
      ID: 2,
      Type: 'Note',
      Content: 'Remember this',
      Priority: '',
      ParentID: null,
      CreatedAt: '2026-01-15T10:00:00Z',
    } as unknown as domain.Entry

    const result = transformEntry(input)

    expect(result.parentId).toBeNull()
    expect(result.priority).toBe('none')
  })

  it('preserves parent ID when present', () => {
    const input = {
      ID: 3,
      Type: 'Task',
      Content: 'Sub-task',
      Priority: 'Low',
      ParentID: 1,
      CreatedAt: '2026-01-15T10:00:00Z',
    } as unknown as domain.Entry

    const result = transformEntry(input)

    expect(result.parentId).toBe(1)
  })

  it('handles missing CreatedAt with current date fallback', () => {
    const input = {
      ID: 4,
      Type: 'Event',
      Content: 'Meeting',
      Priority: '',
      ParentID: undefined,
      CreatedAt: undefined,
    } as unknown as domain.Entry

    const result = transformEntry(input)

    expect(result.loggedDate).toBeDefined()
    expect(new Date(result.loggedDate).getTime()).not.toBeNaN()
  })
})

describe('transformDayEntries', () => {
  it('transforms day entries with entries', () => {
    const input = {
      Date: '2026-01-15T00:00:00Z',
      Location: 'Office',
      Mood: 'Happy',
      Weather: 'Sunny',
      Entries: [
        {
          ID: 1,
          Type: 'Task',
          Content: 'Work item',
          Priority: 'Medium',
          CreatedAt: '2026-01-15T10:00:00Z',
        },
      ],
    } as unknown as service.DayEntries

    const result = transformDayEntries(input)

    expect(result.date).toBe('2026-01-15')
    expect(result.location).toBe('Office')
    expect(result.mood).toBe('Happy')
    expect(result.weather).toBe('Sunny')
    expect(result.entries).toHaveLength(1)
    expect(result.entries[0].content).toBe('Work item')
  })

  it('handles empty entries array', () => {
    const input = {
      Date: '2026-01-15T00:00:00Z',
      Entries: [],
    } as unknown as service.DayEntries

    const result = transformDayEntries(input)

    expect(result.entries).toEqual([])
  })

  it('handles null entries by defaulting to empty array', () => {
    const input = {
      Date: '2026-01-15T00:00:00Z',
      Entries: null,
    } as unknown as service.DayEntries

    const result = transformDayEntries(input)

    expect(result.entries).toEqual([])
  })
})

describe('transformHabit', () => {
  it('transforms habit status correctly', () => {
    const input = {
      ID: 1,
      Name: 'Exercise',
      GoalPerDay: 1,
      CurrentStreak: 5,
      CompletionPercent: 80,
      TodayCount: 1,
      DayHistory: [
        { Date: '2024-01-01T00:00:00Z', Completed: true, Count: 1 },
        { Date: '2024-01-02T00:00:00Z', Completed: true, Count: 1 },
        { Date: '2024-01-03T00:00:00Z', Completed: false, Count: 0 },
        { Date: '2024-01-04T00:00:00Z', Completed: true, Count: 1 },
        { Date: '2024-01-05T00:00:00Z', Completed: true, Count: 2 },
        { Date: '2024-01-06T00:00:00Z', Completed: true, Count: 1 },
        { Date: '2024-01-07T00:00:00Z', Completed: false, Count: 0 },
      ],
    } as unknown as service.HabitStatus

    const result = transformHabit(input)

    expect(result).toEqual({
      id: 1,
      name: 'Exercise',
      streak: 5,
      completionRate: 80,
      goal: 1,
      dayHistory: [
        { date: '2024-01-01', completed: true, count: 1 },
        { date: '2024-01-02', completed: true, count: 1 },
        { date: '2024-01-03', completed: false, count: 0 },
        { date: '2024-01-04', completed: true, count: 1 },
        { date: '2024-01-05', completed: true, count: 2 },
        { date: '2024-01-06', completed: true, count: 1 },
        { date: '2024-01-07', completed: false, count: 0 },
      ],
      todayLogged: true,
      todayCount: 1,
    })
  })

  it('sets todayLogged to false when TodayCount is 0', () => {
    const input = {
      ID: 2,
      Name: 'Read',
      GoalPerDay: 1,
      CurrentStreak: 0,
      CompletionPercent: 50,
      TodayCount: 0,
      DayHistory: [],
    } as unknown as service.HabitStatus

    const result = transformHabit(input)

    expect(result.todayLogged).toBe(false)
  })

  it('handles null DayHistory by defaulting to empty array', () => {
    const input = {
      ID: 3,
      Name: 'Meditate',
      GoalPerDay: 1,
      CurrentStreak: 0,
      CompletionPercent: 0,
      TodayCount: 0,
      DayHistory: null,
    } as unknown as service.HabitStatus

    const result = transformHabit(input)

    expect(result.dayHistory).toEqual([])
  })
})

describe('transformList', () => {
  it('transforms list with items correctly', () => {
    const input = {
      ID: 1,
      Name: 'Shopping List',
      Items: [
        { RowID: 1, Content: 'Milk', Type: 'Task' },
        { RowID: 2, Content: 'Bread', Type: 'Done' },
        { RowID: 3, Content: 'Eggs', Type: 'Task' },
      ],
    } as unknown as wails.ListWithItems

    const result = transformList(input)

    expect(result.id).toBe(1)
    expect(result.name).toBe('Shopping List')
    expect(result.items).toHaveLength(3)
    expect(result.doneCount).toBe(1)
    expect(result.totalCount).toBe(3)
  })

  it('correctly identifies done items', () => {
    const input = {
      ID: 2,
      Name: 'Todo',
      Items: [
        { RowID: 1, Content: 'Item 1', Type: 'DONE' },
        { RowID: 2, Content: 'Item 2', Type: 'done' },
      ],
    } as unknown as wails.ListWithItems

    const result = transformList(input)

    expect(result.items[0].done).toBe(true)
    expect(result.items[1].done).toBe(true)
    expect(result.doneCount).toBe(2)
  })

  it('handles empty items array', () => {
    const input = {
      ID: 3,
      Name: 'Empty List',
      Items: [],
    } as unknown as wails.ListWithItems

    const result = transformList(input)

    expect(result.items).toEqual([])
    expect(result.doneCount).toBe(0)
    expect(result.totalCount).toBe(0)
  })

  it('handles null items by defaulting to empty array', () => {
    const input = {
      ID: 4,
      Name: 'Null List',
      Items: null,
    } as unknown as wails.ListWithItems

    const result = transformList(input)

    expect(result.items).toEqual([])
    expect(result.doneCount).toBe(0)
    expect(result.totalCount).toBe(0)
  })
})

describe('transformGoal', () => {
  it('transforms goal correctly', () => {
    const input = {
      ID: 1,
      EntityID: 'goal-123',
      Content: 'Learn TypeScript',
      Month: '2026-01-01T00:00:00Z',
      Status: 'active',
    } as unknown as domain.Goal

    const result = transformGoal(input)

    expect(result.id).toBe(1)
    expect(result.content).toBe('Learn TypeScript')
    expect(result.month).toBe('2026-01')
    expect(result.completed).toBe(false)
  })

  it('marks goal as completed when status is done', () => {
    const input = {
      ID: 2,
      Content: 'Finish project',
      Month: '2026-01-01T00:00:00Z',
      Status: 'done',
    } as unknown as domain.Goal

    const result = transformGoal(input)

    expect(result.completed).toBe(true)
  })

  it('handles missing Month by using current month', () => {
    const input = {
      ID: 3,
      Content: 'Goal without month',
      Month: null,
      Status: 'active',
    } as unknown as domain.Goal

    const result = transformGoal(input)

    expect(result.month).toMatch(/^\d{4}-\d{2}$/)
  })
})
