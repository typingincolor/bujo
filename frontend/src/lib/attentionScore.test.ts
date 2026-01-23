import { describe, it, expect } from 'vitest'
import { calculateAttentionScore, AttentionIndicator } from './attentionScore'
import { Entry } from '@/types/bujo'

const createEntry = (overrides: Partial<Entry> = {}): Entry => ({
  id: 1,
  entityId: 'e1',
  type: 'task',
  content: 'Test task',
  priority: '',
  parentId: null,
  depth: 0,
  loggedDate: new Date().toISOString(),
  scheduledDate: undefined,
  migrationCount: 0,
  ...overrides,
})

describe('calculateAttentionScore', () => {
  it('returns 0 for a new task with no special conditions', () => {
    const entry = createEntry()
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBe(0)
  })

  it('adds 50 points for past scheduled date', () => {
    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)

    const entry = createEntry({
      scheduledDate: yesterday.toISOString(),
    })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBeGreaterThanOrEqual(50)
    expect(result.indicators).toContain('overdue')
  })

  it('adds 30 points for any priority set', () => {
    const entry = createEntry({ priority: 'low' })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBeGreaterThanOrEqual(30)
  })

  it('adds additional 20 points for high priority', () => {
    const entry = createEntry({ priority: 'high' })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBeGreaterThanOrEqual(50) // 30 + 20
    expect(result.indicators).toContain('priority')
  })

  it('adds 25 points for items older than 7 days', () => {
    const eightDaysAgo = new Date()
    eightDaysAgo.setDate(eightDaysAgo.getDate() - 8)

    const entry = createEntry({
      loggedDate: eightDaysAgo.toISOString(),
    })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBeGreaterThanOrEqual(25)
    expect(result.indicators).toContain('aging')
  })

  it('adds 15 points for items older than 3 days but less than 7', () => {
    const fourDaysAgo = new Date()
    fourDaysAgo.setDate(fourDaysAgo.getDate() - 4)

    const entry = createEntry({
      loggedDate: fourDaysAgo.toISOString(),
    })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBe(15)
  })

  it('adds 15 points per migration', () => {
    const entry = createEntry({ migrationCount: 2 })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBe(30) // 15 * 2
    expect(result.indicators).toContain('migrated')
  })

  it('adds 20 points for urgent keywords in content', () => {
    const entry = createEntry({ content: 'This is urgent!' })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBeGreaterThanOrEqual(20)
  })

  it('adds 10 points for questions', () => {
    const entry = createEntry({ type: 'question' })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBe(10)
  })

  it('adds 5 points for items with event parent', () => {
    const entry = createEntry({ parentId: 1 })
    const result = calculateAttentionScore(entry, new Date(), 'event')
    expect(result.score).toBe(5)
  })

  it('combines multiple conditions', () => {
    const fourDaysAgo = new Date()
    fourDaysAgo.setDate(fourDaysAgo.getDate() - 4)

    const entry = createEntry({
      priority: 'high',
      loggedDate: fourDaysAgo.toISOString(),
      migrationCount: 1,
    })
    const result = calculateAttentionScore(entry, new Date())
    // 30 (priority) + 20 (high) + 15 (age) + 15 (migration) = 80
    expect(result.score).toBe(80)
  })
})

describe('AttentionIndicator formatting', () => {
  it('returns overdue indicator for past scheduled date', () => {
    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)

    const entry = createEntry({ scheduledDate: yesterday.toISOString() })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.indicators).toContain('overdue')
  })

  it('returns migrated indicator with count', () => {
    const entry = createEntry({ migrationCount: 2 })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.indicators).toContain('migrated')
    expect(result.migrationCount).toBe(2)
  })

  it('returns aging indicator for old items', () => {
    const fourDaysAgo = new Date()
    fourDaysAgo.setDate(fourDaysAgo.getDate() - 4)

    const entry = createEntry({ loggedDate: fourDaysAgo.toISOString() })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.indicators).toContain('aging')
  })
})
