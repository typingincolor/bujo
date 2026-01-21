import { describe, it, expect } from 'vitest'
import { formatDateForInput, parseDateFromInput } from './dateUtils'

describe('formatDateForInput', () => {
  it('returns local date string in YYYY-MM-DD format', () => {
    // Create a date at midnight local time on Jan 21, 2026
    const date = new Date(2026, 0, 21, 0, 0, 0)

    const result = formatDateForInput(date)

    expect(result).toBe('2026-01-21')
  })

  it('handles dates near midnight correctly regardless of timezone', () => {
    // Create a date at 11:59 PM local time on Jan 21, 2026
    const date = new Date(2026, 0, 21, 23, 59, 0)

    const result = formatDateForInput(date)

    expect(result).toBe('2026-01-21')
  })
})

describe('parseDateFromInput', () => {
  it('parses YYYY-MM-DD string to local midnight date', () => {
    const result = parseDateFromInput('2026-01-21')

    expect(result).not.toBeNull()
    // Should be midnight local time
    expect(result!.getFullYear()).toBe(2026)
    expect(result!.getMonth()).toBe(0) // January is 0
    expect(result!.getDate()).toBe(21)
    expect(result!.getHours()).toBe(0)
    expect(result!.getMinutes()).toBe(0)
  })

  it('returns null for invalid input', () => {
    const result = parseDateFromInput('')

    expect(result).toBeNull()
  })

  it('returns null for malformed input', () => {
    expect(parseDateFromInput('invalid')).toBeNull()
    expect(parseDateFromInput('not-a-date')).toBeNull()
  })
})
