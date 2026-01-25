/**
 * Shared layout constants for consistent component positioning.
 * These values must stay in sync across components that share layouts.
 */

/** Width of the navigation sidebar on the left (w-56 = 14rem = 224px) */
export const NAV_SIDEBAR_WIDTH_CLASS = 'w-56';
export const NAV_SIDEBAR_LEFT_CLASS = 'left-56';

/** Default width of the journal sidebar on the right (32rem = 512px)
 * Note: Actual width is dynamic and controlled by JournalSidebar component state.
 * This constant is kept for reference but is no longer used in layout calculations.
 */
export const JOURNAL_SIDEBAR_WIDTH_CLASS = 'w-[32rem]';
export const JOURNAL_SIDEBAR_RIGHT_CLASS = 'right-[32rem]';
