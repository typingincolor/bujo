# Capture Bar & Week Summary Design

**Date:** 2026-01-23
**Status:** Approved

## Overview

This design addresses friction in the current capture experience and adds value to the weekly view. The goal is to make note capture effortless across all use cases (quick single entries, batch capture, and contextual notes under meetings) while providing actionable insights in the weekly review.

## Problem Statement

**Current capture issues:**
- CaptureModal blocks context - can't see journal while typing
- Must manually type entry prefixes (`. - o ?`)
- Two disconnected flows (modal for batch, inline for quick)
- Hidden "Add entry" button is easy to miss

**Weekly view limitations:**
- Can view but not add entries
- No high-level summary of the week
- No prioritization of what needs attention

## Solution

### Part 1: Always-Ready Capture Bar

A persistent input at the bottom of the day view that's always visible and ready for capture.

#### Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  [Task] [Note] [Event] [Question]              [📎]            │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │ Add a task...                                             │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

With parent context:

```
┌─────────────────────────────────────────────────────────────────┐
│  Adding to: Team standup                               [×]      │
│  [Task] [Note] [Event] [Question]              [📎]            │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │ Add a task...                                             │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

#### Type Selection

| Method | Description |
|--------|-------------|
| Mouse click | Click type button to select |
| Tab (empty input) | Cycles: Task → Note → Event → Question → Task |
| Prefix (power user) | Type `. ` `- ` `o ` `? ` at start of input |
| Memory | Last-used type persists in localStorage |

#### Interaction Behavior

| Action | Behavior |
|--------|----------|
| Enter | Submit entry, clear input, keep focus for next entry |
| Shift+Enter | New line (input grows, max ~4 lines before scroll) |
| Escape | Clear input (or blur if already empty) |
| Escape (with parent) | First clears parent context, second clears input |

#### Multi-line & Hierarchy

- Supports multiple lines via Shift+Enter
- Indented lines become children (same parsing as current modal)
- Example:
  ```
  o Team standup
    - John blocked on API
    . Help John debug
  ```

#### Contextual Adding (Child Entries)

- **Default:** Adding at root (no context bar shown)
- **`A` on selected entry:** Sets that entry as parent, shows "Adding to: [entry]"
- **Click [×] or Escape twice:** Clears parent, back to root mode
- After submit, stays in child mode for rapid note-taking under same parent
- **`r`:** Explicitly add at root even if something is selected

#### Draft Persistence

- Content saves to localStorage on every keystroke
- On page load, restore draft with subtle "Draft restored" indicator
- Parent context also persisted
- Clear draft only on successful submit

#### File Import

- Paperclip icon in capture bar (right side)
- Click to open file picker, or drag-drop onto capture bar
- File contents append to current input

#### Visual States

| State | Visual Treatment |
|-------|------------------|
| Empty, blurred | Subtle placeholder, muted border |
| Empty, focused | Brighter border, placeholder visible |
| Has content | Normal text, subtle "draft" indicator |
| Has parent context | "Adding to: [entry]" bar above input |
| Submitting | Brief disabled state |
| Error | Red border, error message below, content preserved |

#### Type Button States

| State | Visual |
|-------|--------|
| Active type | Filled background, primary color |
| Inactive | Outline/ghost style, muted |
| Hover | Slight highlight |

#### Feedback

- Input clears instantly on submit
- No toast/notification (too noisy for rapid entry)
- Entry appears in journal above (immediate visual confirmation)
- On error: input keeps content, error message appears, retry on Enter

### Part 2: Keyboard Shortcuts

#### From Journal View (capture bar blurred)

| Key | Action |
|-----|--------|
| `i` | Focus capture bar |
| `A` | Focus capture bar + set selected entry as parent |
| `r` | Focus capture bar in root mode |
| `↑` / `↓` | Navigate entries |
| `j` / `k` | Navigate entries (vim alternative) |

#### In Capture Bar (focused)

| Key | Action |
|-----|--------|
| Enter | Submit entry |
| Shift+Enter | New line |
| Tab (empty) | Cycle entry type |
| Escape | Clear input (or blur if empty) |

#### Prefix Shortcuts (in input)

| Prefix | Type |
|--------|------|
| `. ` | Task |
| `- ` | Note |
| `o ` | Event |
| `? ` | Question |

### Part 3: Week Summary

A summary panel at the top of the weekly view providing actionable insights.

#### Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  Week of Jan 13 - Jan 19                                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  TASK FLOW                                                          │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐          │
│  │ Created │ →  │  Done   │    │Migrated │    │  Open   │          │
│  │   12    │    │    7    │    │    2    │    │    3    │          │
│  └─────────┘    └─────────┘    └─────────┘    └─────────┘          │
│                                                                     │
│  MEETINGS                           NEEDS ATTENTION                 │
│  ┌────────────────────────────┐    ┌────────────────────────────┐  │
│  │ Team standup        4 items│    │ . Review PR #42       🔴   │  │
│  │ 1:1 with Sarah      2 items│    │ . Send invoice        ⚡   │  │
│  │ Sprint planning     7 items│    │ ? Deploy date?        🔄   │  │
│  └────────────────────────────┘    └────────────────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

#### Task Flow

- Visual pipeline: Created → Done / Migrated / Open
- Numbers update live as entries change
- Click a number to filter day views to just those entries

#### Meetings (Events with children)

- Lists events that have child entries
- Shows count of children (notes + tasks from that meeting)
- Click to jump to that entry in the day view

#### Needs Attention

- Open tasks (not done, not migrated, not canceled)
- Unanswered questions
- Sorted by attention score (see below)
- Click to jump to entry
- Show top 5, with "Show all (N)" to expand

### Part 4: Attention Scoring System

Rule-based scoring to surface items that need attention.

#### Scoring Rules

| Condition | Points | Rationale |
|-----------|--------|-----------|
| Past scheduled date | +50 | Explicitly overdue |
| Has priority set (any level) | +30 | User marked as important |
| Priority high/urgent | +20 (additional) | Extra weight for top priority |
| Age > 7 days | +25 | Week old, likely forgotten |
| Age > 3 days | +15 | Getting stale |
| Migrated before | +15 per migration | Keeps slipping |
| Content contains "urgent", "asap", "blocker", "waiting" | +20 | Language signals |
| Is a question (not task) | +10 | Questions often block decisions |
| Has children | +10 | Bigger scope item |
| Parent is an event | +5 | Meeting action item |

#### Display

- Show top 5 items by score
- No numeric score shown to user
- Subtle indicators for *why* it surfaced:
  - 🔴 "Overdue" (past scheduled date)
  - 🔄 "Migrated 2x" (repeated migration)
  - ⚡ "3+ days old" (aging)
- "Show all (N)" link expands to full list sorted by score

## Migration Notes

- **CaptureModal:** Can be removed or kept as legacy access. The capture bar handles all use cases including bulk import via multi-line and file drag-drop.
- **InlineEntryInput:** Replaced by capture bar.
- **"+ Add entry" button:** Replaced by always-visible capture bar.

## Implementation Sequence

1. **Capture Bar component** - New `CaptureBar.tsx` with type selection, input, draft persistence
2. **Parent context mode** - Add "Adding to:" UI and child entry logic
3. **File import in capture bar** - Move file handling from modal
4. **Keyboard shortcuts** - Update App.tsx handlers for new shortcuts
5. **Week Summary component** - New `WeekSummary.tsx` with task flow and meetings
6. **Attention scoring** - Implement scoring logic and "Needs Attention" section
7. **Remove/deprecate old components** - CaptureModal, InlineEntryInput, add button

## Open Questions

None - design approved.
