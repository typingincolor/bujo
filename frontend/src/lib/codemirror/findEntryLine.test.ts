import { describe, it, expect } from 'vitest'
import { findEntryLine } from './findEntryLine'

describe('findEntryLine', () => {
  it('returns null when content is empty', () => {
    expect(findEntryLine('', 'some task')).toBeNull()
  })

  it('returns null when search text is empty', () => {
    expect(findEntryLine('.task one\n-note two', '')).toBeNull()
  })

  it('finds a task line matching the content', () => {
    const doc = '.task one\n-note two\n.task three'
    expect(findEntryLine(doc, 'task one')).toEqual({ line: 1, from: 0, to: 9 })
  })

  it('finds a note line matching the content', () => {
    const doc = '.task one\n-note two\n.task three'
    expect(findEntryLine(doc, 'note two')).toEqual({ line: 2, from: 10, to: 19 })
  })

  it('matches indented entries', () => {
    const doc = '.parent\n  .child task'
    expect(findEntryLine(doc, 'child task')).toEqual({ line: 2, from: 8, to: 21 })
  })

  it('returns first match when multiple lines contain the same text', () => {
    const doc = '.duplicate\n-other\n.duplicate'
    expect(findEntryLine(doc, 'duplicate')).toEqual({ line: 1, from: 0, to: 10 })
  })

  it('matches with different entry type symbols', () => {
    const doc = 'oMy event\nxDone task\n>Migrated thing'
    expect(findEntryLine(doc, 'Done task')).toEqual({ line: 2, from: 10, to: 20 })
  })

  it('ignores lines without entry symbols', () => {
    const doc = 'random text\n.actual task'
    expect(findEntryLine(doc, 'random text')).toBeNull()
  })

  it('handles priority markers in the line', () => {
    const doc = '.!important task\n-note'
    expect(findEntryLine(doc, 'important task')).toEqual({ line: 1, from: 0, to: 16 })
  })
})
