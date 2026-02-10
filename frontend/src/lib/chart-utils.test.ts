import { describe, it, expect } from 'vitest'
import { flattenEntries, getWeekKey, formatWeekLabel } from './chart-utils'
import { Entry } from '@/types/bujo'

const makeEntry = (id: number, children?: Entry[]): Entry => ({
  id,
  content: `Entry ${id}`,
  type: 'task',
  priority: 'none',
  parentId: null,
  loggedDate: '2026-02-10',
  ...(children ? { children } : {}),
})

describe('flattenEntries', () => {
  it('returns empty array for empty input', () => {
    expect(flattenEntries([])).toEqual([])
  })

  it('returns flat entries unchanged', () => {
    const entries = [makeEntry(1), makeEntry(2)]
    const result = flattenEntries(entries)
    expect(result).toHaveLength(2)
    expect(result[0].id).toBe(1)
    expect(result[1].id).toBe(2)
  })

  it('flattens nested children', () => {
    const entries = [
      makeEntry(1, [makeEntry(2), makeEntry(3, [makeEntry(4)])]),
    ]
    const result = flattenEntries(entries)
    expect(result).toHaveLength(4)
    expect(result.map(e => e.id)).toEqual([1, 2, 3, 4])
  })
})

describe('getWeekKey', () => {
  it('returns Monday date for a Wednesday', () => {
    expect(getWeekKey('2026-02-11')).toBe('2026-02-09')
  })

  it('returns same date for a Monday', () => {
    expect(getWeekKey('2026-02-09')).toBe('2026-02-09')
  })

  it('returns previous Monday for a Sunday', () => {
    expect(getWeekKey('2026-02-15')).toBe('2026-02-09')
  })
})

describe('formatWeekLabel', () => {
  it('formats a date string as "Mon DD"', () => {
    expect(formatWeekLabel('2026-02-09')).toBe('Feb 9')
  })

  it('handles unknown type by converting to string', () => {
    expect(formatWeekLabel('2026-01-05')).toBe('Jan 5')
  })
})
