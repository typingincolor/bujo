import { describe, it, expect } from 'vitest';
import { startOfDay, startOfWeek, endOfWeek } from 'date-fns';

// Function to calculate review date range
// This mimics the logic in App.tsx
function calculateReviewDateRange(anchorDate: Date): { start: Date; end: Date; dayCount: number } {
  // Review view: show the full week (Mon-Sun) containing reviewAnchorDate
  // Use date-fns to get proper week boundaries with Monday as first day
  const reviewStart = startOfWeek(anchorDate, { weekStartsOn: 1 });
  const reviewEnd = endOfWeek(anchorDate, { weekStartsOn: 1 });

  // Calculate how many days this range represents
  // Backend loop is: for d := from; !d.After(to); d = d.AddDate(0, 0, 1)
  // This is INCLUSIVE of the end date
  const dayCount = Math.floor((reviewEnd.getTime() - reviewStart.getTime()) / (24 * 60 * 60 * 1000)) + 1;

  return { start: reviewStart, end: reviewEnd, dayCount };
}

describe('Review Date Range', () => {
  it('should produce exactly 7 days for weekly review (Mon-Sun)', () => {
    // Anchor date is Jan 26, 2026 (Monday - today)
    const anchorDate = startOfDay(new Date(2026, 0, 26)); // Month is 0-indexed

    const { start, end, dayCount } = calculateReviewDateRange(anchorDate);

    // We expect 7 days: Mon Jan 26 - Sun Feb 1
    expect(dayCount).toBe(7);

    // Start should be Monday Jan 26
    expect(start.getDate()).toBe(26);
    expect(start.getMonth()).toBe(0); // January
    expect(start.getDay()).toBe(1); // Monday

    // End should be Sunday Feb 1
    expect(end.getDate()).toBe(1);
    expect(end.getMonth()).toBe(1); // February
    expect(end.getDay()).toBe(0); // Sunday
  });

  it('calculates correct date range for mid-week anchor date', () => {
    // Test with Wed Jan 15, 2026 (mid-week)
    const anchorDate = startOfDay(new Date(2026, 0, 15));

    const { start, end, dayCount } = calculateReviewDateRange(anchorDate);

    // Week containing Jan 15 is Mon Jan 12 - Sun Jan 18
    expect(dayCount).toBe(7);
    expect(start.getDate()).toBe(12); // Monday Jan 12
    expect(start.getDay()).toBe(1); // Monday
    expect(end.getDate()).toBe(18); // Sunday Jan 18
    expect(end.getDay()).toBe(0); // Sunday
  });
});
