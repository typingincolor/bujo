import { describe, it, expect } from 'vitest'
import { toWailsTime } from './wailsTime'

describe('toWailsTime', () => {
  it('preserves the local date when serializing midnight', () => {
    // Create a date at midnight local time on January 19, 2026
    const localMidnight = new Date(2026, 0, 19, 0, 0, 0, 0) // Jan 19, 2026 00:00:00 local

    const result = toWailsTime(localMidnight)

    // The result should contain "2026-01-19" regardless of timezone
    // This test will fail if toISOString() is used because it converts to UTC,
    // which shifts the date backwards for positive UTC offsets (e.g., GMT+1)
    expect(result).toContain('2026-01-19')
  })

  it('preserves the local date for end of day times', () => {
    // Create a date at 11:59 PM local time on January 19, 2026
    const localEndOfDay = new Date(2026, 0, 19, 23, 59, 59, 0)

    const result = toWailsTime(localEndOfDay)

    // Should still be January 19th, not shift to January 20th
    expect(result).toContain('2026-01-19')
  })

  it('returns a string that Go can parse as RFC3339', () => {
    const date = new Date(2026, 0, 19, 14, 30, 0, 0) // 2:30 PM local

    const result = toWailsTime(date)

    // Should be a valid ISO 8601 / RFC3339 format string
    // Go's time.Time expects this format for JSON unmarshaling
    expect(result).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/)
  })
})
