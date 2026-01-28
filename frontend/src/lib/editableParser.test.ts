import { describe, it, expect } from 'vitest'
import { parseLine, parseDocument } from './editableParser'

describe('parseLine', () => {
  describe('symbol recognition', () => {
    it('parses task symbol (.)', () => {
      const result = parseLine('. Buy groceries', 1)

      expect(result.symbol).toBe('.')
      expect(result.content).toBe('Buy groceries')
      expect(result.isValid).toBe(true)
    })

    it('parses note symbol (-)', () => {
      const result = parseLine('- Meeting went well', 1)

      expect(result.symbol).toBe('-')
      expect(result.content).toBe('Meeting went well')
      expect(result.isValid).toBe(true)
    })

    it('parses event symbol (o)', () => {
      const result = parseLine('o Team standup at 10am', 1)

      expect(result.symbol).toBe('o')
      expect(result.content).toBe('Team standup at 10am')
      expect(result.isValid).toBe(true)
    })

    it('parses done symbol (x)', () => {
      const result = parseLine('x Finished report', 1)

      expect(result.symbol).toBe('x')
      expect(result.content).toBe('Finished report')
      expect(result.isValid).toBe(true)
    })

    it('parses cancelled symbol (~)', () => {
      const result = parseLine('~ No longer needed', 1)

      expect(result.symbol).toBe('~')
      expect(result.content).toBe('No longer needed')
      expect(result.isValid).toBe(true)
    })

    it('parses question symbol (?)', () => {
      const result = parseLine('? How does auth work', 1)

      expect(result.symbol).toBe('?')
      expect(result.content).toBe('How does auth work')
      expect(result.isValid).toBe(true)
    })

    it('parses migrated symbol (>)', () => {
      const result = parseLine('> Moved task', 1)

      expect(result.symbol).toBe('>')
      expect(result.content).toBe('Moved task')
      expect(result.isValid).toBe(true)
    })

    it('marks unknown symbol as invalid', () => {
      const result = parseLine('^ Unknown', 1)

      expect(result.isValid).toBe(false)
      expect(result.errorMessage).toBe('Unknown entry type')
    })
  })

  describe('priority parsing', () => {
    it('parses high priority (!!!)', () => {
      const result = parseLine('. !!! Urgent task', 1)

      expect(result.symbol).toBe('.')
      expect(result.priority).toBe(1)
      expect(result.content).toBe('Urgent task')
    })

    it('parses medium priority (!!)', () => {
      const result = parseLine('. !! High priority', 1)

      expect(result.symbol).toBe('.')
      expect(result.priority).toBe(2)
      expect(result.content).toBe('High priority')
    })

    it('parses low priority (!)', () => {
      const result = parseLine('. ! Low priority', 1)

      expect(result.symbol).toBe('.')
      expect(result.priority).toBe(3)
      expect(result.content).toBe('Low priority')
    })

    it('defaults to no priority (0) when no markers', () => {
      const result = parseLine('. Normal task', 1)

      expect(result.priority).toBe(0)
      expect(result.content).toBe('Normal task')
    })
  })

  describe('indentation', () => {
    it('parses depth 0 (no indentation)', () => {
      const result = parseLine('. Root task', 1)

      expect(result.depth).toBe(0)
    })

    it('parses depth 1 (2 spaces)', () => {
      const result = parseLine('  . Child task', 1)

      expect(result.depth).toBe(1)
    })

    it('parses depth 2 (4 spaces)', () => {
      const result = parseLine('    . Grandchild', 1)

      expect(result.depth).toBe(2)
    })

    it('normalizes tab to spaces', () => {
      const result = parseLine('\t. Tabbed task', 1)

      expect(result.depth).toBe(1)
    })

    it('rounds odd indentation to nearest level', () => {
      const result = parseLine('   . Odd indent', 1)

      expect(result.depth).toBe(2) // 3 spaces rounds to 2 (depth 1)
    })
  })

  describe('migration syntax', () => {
    it('detects migration pattern with ISO date', () => {
      const result = parseLine('>[2026-01-29] Call dentist', 1)

      expect(result.symbol).toBe('>')
      expect(result.migrationTarget).toBe('2026-01-29')
      expect(result.content).toBe('Call dentist')
      expect(result.isValid).toBe(true)
    })

    it('detects migration pattern with natural language', () => {
      const result = parseLine('>[tomorrow] Review PR', 1)

      expect(result.symbol).toBe('>')
      expect(result.migrationTarget).toBe('tomorrow')
      expect(result.content).toBe('Review PR')
    })

    it('detects migration pattern with multi-word date', () => {
      const result = parseLine('>[next monday] Submit report', 1)

      expect(result.symbol).toBe('>')
      expect(result.migrationTarget).toBe('next monday')
      expect(result.content).toBe('Submit report')
    })

    it('handles indented migration', () => {
      const result = parseLine('  >[tomorrow] Indented migration', 1)

      expect(result.depth).toBe(1)
      expect(result.symbol).toBe('>')
      expect(result.migrationTarget).toBe('tomorrow')
      expect(result.content).toBe('Indented migration')
    })
  })

  describe('edge cases', () => {
    it('returns empty result for empty line', () => {
      const result = parseLine('', 1)

      expect(result.isEmpty).toBe(true)
      expect(result.isValid).toBe(true)
    })

    it('returns empty result for whitespace-only line', () => {
      const result = parseLine('   ', 1)

      expect(result.isEmpty).toBe(true)
      expect(result.isValid).toBe(true)
    })

    it('skips header lines (──)', () => {
      const result = parseLine('── Monday, Jan 27 ──────────────────', 1)

      expect(result.isHeader).toBe(true)
      expect(result.isValid).toBe(true)
    })

    it('marks missing content as error', () => {
      const result = parseLine('. ', 1)

      expect(result.isValid).toBe(false)
      expect(result.errorMessage).toBe('Entry content required')
    })

    it('preserves line number in result', () => {
      const result = parseLine('. Test', 42)

      expect(result.lineNumber).toBe(42)
    })

    it('handles symbol-only line as missing content', () => {
      const result = parseLine('.', 1)

      expect(result.isValid).toBe(false)
      expect(result.errorMessage).toBe('Entry content required')
    })
  })
})

describe('parseDocument', () => {
  it('parses multiple lines', () => {
    const doc = `. Buy groceries
- Meeting went well
x Finished report`

    const result = parseDocument(doc)

    expect(result.lines).toHaveLength(3)
    expect(result.lines[0].symbol).toBe('.')
    expect(result.lines[1].symbol).toBe('-')
    expect(result.lines[2].symbol).toBe('x')
  })

  it('assigns sequential line numbers', () => {
    const doc = `. First
. Second
. Third`

    const result = parseDocument(doc)

    expect(result.lines[0].lineNumber).toBe(1)
    expect(result.lines[1].lineNumber).toBe(2)
    expect(result.lines[2].lineNumber).toBe(3)
  })

  it('handles document with headers', () => {
    const doc = `── Monday, Jan 27 ──────────────────
. Buy groceries
- Meeting went well`

    const result = parseDocument(doc)

    expect(result.lines).toHaveLength(3)
    expect(result.lines[0].isHeader).toBe(true)
    expect(result.lines[1].symbol).toBe('.')
  })

  it('handles empty document', () => {
    const result = parseDocument('')

    expect(result.lines).toHaveLength(0)
    expect(result.isValid).toBe(true)
    expect(result.errors).toHaveLength(0)
  })

  it('collects errors from invalid lines', () => {
    const doc = `. Valid task
^ Invalid line
. Another valid`

    const result = parseDocument(doc)

    expect(result.isValid).toBe(false)
    expect(result.errors).toHaveLength(1)
    expect(result.errors[0].lineNumber).toBe(2)
    expect(result.errors[0].message).toBe('Unknown entry type')
  })

  it('handles hierarchical document', () => {
    const doc = `. Parent task
  . Child task
    . Grandchild
  - Sibling note`

    const result = parseDocument(doc)

    expect(result.lines[0].depth).toBe(0)
    expect(result.lines[1].depth).toBe(1)
    expect(result.lines[2].depth).toBe(2)
    expect(result.lines[3].depth).toBe(1)
  })

  it('preserves empty lines', () => {
    const doc = `. First

. Second`

    const result = parseDocument(doc)

    expect(result.lines).toHaveLength(3)
    expect(result.lines[1].isEmpty).toBe(true)
  })
})
