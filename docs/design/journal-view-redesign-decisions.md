# Journal View Redesign Decisions

> Implementation decisions for aligning the Journal view with the bujo-canvas mockup.
> Feature branch: `feature/journal-view-redesign`
> Date: 2026-01-25

## Summary

This document captures decisions made during planning that differ from or clarify the mockup specifications.

---

## QuickStats Cards

**Decision:** Make existing QuickStats cards smaller to match mockup styling.

**Rationale:** Current implementation is functionally correct but visually larger than the mockup design.

---

## Right Sidebar Layout (Journal View)

**Decision:** Show both Overdue Items and ContextPanel.

**Structure:**
```
┌─────────────────────────┐
│ Overdue Items           │ ← Collapsible section
│ (with attention scoring)│
│ ├─ Entry with score     │
│ └─ Entry with score     │
├─────────────────────────┤
│ Context                 │ ← Always visible
│ ├─ Ancestor 1           │
│ ├─ Ancestor 2           │
│ └─ Selected Entry ◀     │
└─────────────────────────┘
```

**Behavior:**
- Overdue Items section is collapsible
- ContextPanel is always visible
- Selecting an overdue item updates the ContextPanel with its ancestor hierarchy
- Context dot indicator shown on overdue entries that have parents
- Overdue items display with attention scoring badges

---

## AI Summary

**Decision:** Keep as-is.

- Toggle button (Sparkles icon) remains in day header
- Summary block appears below header when toggled on
- No changes needed

---

## Attention Scoring

**Decision:** Use attention scoring with color-coded badges in Overdue Items section.

**Score factors:**
- Days overdue
- Priority level
- Migration count
- Age

**Display:** Color-coded badge with tooltip showing score breakdown.

---

## CaptureBar Behavior

### Type Detection

**Decision:** User specifies type via prefix characters. **Prefix is kept in content.**

| Prefix | Type | Example |
|--------|------|---------|
| `.` | task | `. Buy groceries` |
| `-` | note | `- Meeting notes from today` |
| `o` | event | `o Doctor appointment at 2pm` |
| `?` | question | `? Should we use Redux?` |

**Rationale:** Matches traditional bullet journal notation where the symbol is part of the written entry. Provides visual consistency between input and display.

### No Tab Cycling

**Decision:** Remove Tab cycling feature from mockup design.

**Rationale:**
- Tab could conflict with indentation in multi-line mode
- Prefix characters are more explicit and match bullet journal conventions
- Simpler mental model for users

### Priority Detection

**Decision:** Priority indicated with `!` characters after type prefix.

| Pattern | Priority |
|---------|----------|
| `!` | low |
| `!!` | medium |
| `!!!` | high |

### Multi-line Entries

**Decision:**
- `Enter` submits the entry
- `Shift+Enter` adds a new line
- User is responsible for their own indentation
- No automatic indentation based on parent depth

**Rationale:** Keeps the CaptureBar simple. Users who want aligned multi-line entries can manually add spaces.

### Typography

**Decision:** Use monospaced font in CaptureBar textarea.

**Rationale:** Helps users align indentation for multi-line entries and matches the journal aesthetic.

### Parent Context

**Decision:** Show "Adding to: [parent content]" with clear button when adding child entry.

- Displayed above or below textarea (implementation detail)
- Clear button (×) removes parent context
- No automatic indentation in textarea

---

---

## Header Layout

**Decision:** Keep all pickers (mood, weather, location, date) and capture mode button in Header only. No duplication in DayView.

**Styling:**
- Match mockup's clean aesthetic
- Picker buttons styled subtly (icon-only until clicked)
- Date display matches mockup format (EEEE, MMMM d, yyyy)
- Background: `bg-card/50` with `border-b`

**Structure:**
```
┌─────────────────────────────────────────────────────────────┐
│ [← Back] Title    📅 Date    😊 ☁️ 📍    [Capture Mode]    │
└─────────────────────────────────────────────────────────────┘
```

---

## CaptureBar Position

**Decision:** CaptureBar anchored to bottom of window (fixed position).

**Implementation:**
- `position: fixed` with `bottom: 0`
- Full width of main content area
- Main content has padding-bottom to prevent overlap

---

## Changes from Mockup

| Mockup Feature | Our Decision | Reason |
|----------------|--------------|--------|
| Tab cycles entry type | Removed | Prefix characters are clearer |
| Type prefix consumed | Prefix kept in content | Matches bullet journal notation |
| Type selector button | Removed | Prefix characters are sufficient |
| Automatic indentation | User responsibility | Simpler implementation |
| Context in DayView header | Removed | Pickers stay in Header only |
| Minimal Header | Keep pickers | Need editing capability |
| CaptureModal for batch entry | FileUploadButton | Mockup updated to use upload popover |

---

## DateNavigator

**Decision:** Implement DateNavigator component for journal day navigation.

**Features:**
- Previous/Next day buttons (ChevronLeft/Right icons)
- Date picker button with calendar popover
- Shows "Today" when viewing current date, otherwise formatted date
- Today button to jump to current date (invisible when already viewing today)

**Placement:** In Header, passed as `actions` prop.

---

## FileUploadButton

**Decision:** Implement FileUploadButton for file attachments. Replaces batch capture modal.

**Features:**
- Popover with drag-and-drop zone
- File list with name, size, and remove button
- Badge showing file count on button
- Files tracked in memory (storage integration later)

**Note:** This removes the need for the previous CaptureModal component.

---

## Implementation Order

1. QuickStats card sizing (smaller)
2. Header styling update (cleaner, mockup aesthetic)
3. DateNavigator component (day navigation with calendar)
4. CaptureBar updates (prefixes, monospace, multi-line, fixed bottom)
5. Right sidebar restructure (Overdue + ContextPanel)
6. Attention scoring for overdue items
7. Selection interaction (overdue → context update)
8. FileUploadButton component (file attachments)
