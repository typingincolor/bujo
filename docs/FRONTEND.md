# Desktop App Guide

The bujo desktop app provides a native macOS experience built with Wails and React.

## Installation

The desktop app is distributed separately from the CLI. Install via Homebrew:

```bash
brew tap typingincolor/tap
brew install --cask bujo-desktop
```

## Views

Navigate between views using the sidebar:

| View | Description |
|------|-------------|
| Today | Today's entries with day navigation |
| Review | Past 7 days with week-by-week navigation |
| Outstanding | Overdue tasks requiring attention |
| Questions | Unanswered question entries |
| Habits | Habit tracker with calendar visualization |
| Lists | Collection lists with progress |
| Goals | Monthly goals |
| Search | Search through all entries |
| Stats | Usage statistics and insights |
| Settings | App configuration |

## Today View

The main view for daily journaling:

- **Day Navigation**: Use left/right arrows or date picker to navigate
- **Quick Stats**: Overview of today's tasks, habits, and goals
- **Entry List**: Hierarchical view of today's entries
- **Quick Capture**: Click the pen icon to add entries

### Entry Operations

Click on entries to select, then:
- **Space**: Mark task done/undone
- **Edit**: Modify entry content
- **Delete**: Remove entry (with confirmation)
- **Migrate**: Move task to a future date
- **Add Child**: Create nested entry

## Habits View

Visual habit tracking with multiple time periods:

| Period | Display |
|--------|---------|
| Week | Last 14 days with daily completion |
| Month | Calendar grid showing completion |
| Quarter | Compact 90-day overview |

### Logging Habits

- **Click** a day cell to log an occurrence
- **Cmd+Click** to remove an occurrence
- Press **w** to cycle through view periods

## Lists View

Manage collection lists (reading lists, shopping, etc.):

- Click a list to view its items
- Check items to mark as done
- Create new lists and items
- Track completion progress

## Goals View

Monthly goals with progress tracking:

- Add goals for the current month
- Mark goals as complete
- Track progress toward monthly objectives

## Questions View

Track questions that need answers:

- View all unanswered questions
- Click to answer a question
- Questions move to answered state when resolved

## Search

Full-text search across all entries:

- Type to search
- Filter by date range and entry type
- Click results to view context

## Keyboard Shortcuts

Press `?` to toggle the keyboard shortcuts panel.

### Today View

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `h` | Previous day |
| `l` | Next day |
| `Space` | Toggle done |
| `x` | Cancel/uncancel |
| `p` | Cycle priority |
| `t` | Cycle type |
| `e` | Edit entry |
| `d` | Delete entry |
| `Enter` | Expand context |
| `c` | Open capture modal |
| `r` | Add root entry |
| `a` | Add sibling (or answer question) |
| `A` | Add child entry |

### Habits View

| Key | Action |
|-----|--------|
| `w` | Cycle view (week/month/quarter) |
| `Click` | Log occurrence |
| `Cmd+Click` | Remove occurrence |

### Global

| Key | Action |
|-----|--------|
| `?` | Toggle keyboard shortcuts panel |

## Day Context

The header displays and allows editing:

- **Mood**: How you're feeling today
- **Weather**: Current conditions
- **Location**: Where you're working from

Click each to edit.

## Architecture

The desktop app uses:

- **Wails**: Go backend with native window frame
- **React**: Frontend UI with TypeScript
- **Tailwind CSS**: Styling with custom design tokens
- **Lucide Icons**: Consistent iconography

The Go backend exposes the same services used by the CLI, ensuring data consistency across interfaces.
