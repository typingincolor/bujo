import { describe, it, expect } from 'vitest';
import { startOfDay } from 'date-fns';

// Function to calculate review date range
// This mimics the logic in App.tsx
function calculateReviewDateRange(anchorDate: Date): { start: Date; end: Date; dayCount: number } {
  // Review view: past 7 days ending at reviewAnchorDate
  // Backend loop uses !d.After(to) which is INCLUSIVE of end date
  // So we pass anchorDate directly as end, not anchorDate + 24h
  const reviewEnd = anchorDate;
  const reviewStart = new Date(anchorDate.getTime() - 6 * 24 * 60 * 60 * 1000);

  // Calculate how many days this range represents
  // Backend loop is: for d := from; !d.After(to); d = d.AddDate(0, 0, 1)
  // This is INCLUSIVE of the end date
  const dayCount = Math.floor((reviewEnd.getTime() - reviewStart.getTime()) / (24 * 60 * 60 * 1000)) + 1;

  return { start: reviewStart, end: reviewEnd, dayCount };
}

describe('Review Date Range', () => {
  it('should produce exactly 7 days for weekly review', () => {
    // Anchor date is Jan 26, 2026 (today)
    const anchorDate = startOfDay(new Date(2026, 0, 26)); // Month is 0-indexed

    const { start, end, dayCount } = calculateReviewDateRange(anchorDate);

    // We expect 7 days: Jan 20-26
    expect(dayCount).toBe(7);

    // Start should be Jan 20
    expect(start.getDate()).toBe(20);
    expect(start.getMonth()).toBe(0); // January

    // End should be Jan 26 (not Jan 27)
    // Since backend loop is inclusive (!d.After(to)), end should be anchorDate
    expect(end.getDate()).toBe(26);
    expect(end.getMonth()).toBe(0); // January
  });

  it('calculates correct date range for different anchor dates', () => {
    // Test with Jan 15, 2026
    const anchorDate = startOfDay(new Date(2026, 0, 15));

    const { start, end, dayCount } = calculateReviewDateRange(anchorDate);

    expect(dayCount).toBe(7);
    expect(start.getDate()).toBe(9); // Jan 9
    expect(end.getDate()).toBe(15); // Jan 15
  });
});
