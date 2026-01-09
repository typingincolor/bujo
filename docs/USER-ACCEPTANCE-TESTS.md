# User Acceptance Tests

Comprehensive acceptance tests for bujo in Given/When/Then format covering both CLI and TUI interfaces.

---

## CLI Tests

### 1. Entry Management

#### 1.1 Adding Entries

```gherkin
Scenario: Add a single task
  Given the user has bujo installed
  When the user runs `bujo add ". Buy groceries"`
  Then the entry ID is printed to stdout
  And the message "Added 1 entry(s)" is printed to stderr
  And a task with content "Buy groceries" exists for today

Scenario: Add multiple entries at once
  Given the user has bujo installed
  When the user runs `bujo add ". Task one" "- Note one" "o Event one"`
  Then three entry IDs are printed to stdout
  And the message "Added 3 entry(s)" is printed to stderr
  And a task, note, and event are created for today

Scenario: Add entries from a file
  Given the user has a file "tasks.txt" containing:
    """
    . Task from file
    - Note from file
    """
  When the user runs `bujo add --file tasks.txt`
  Then two entries are created
  And the message "Added 2 entry(s)" is printed to stderr

Scenario: Add entries with location context
  Given the user has bujo installed
  When the user runs `bujo add --at "Home Office" ". Work on project"`
  Then a task is created with location "Home Office"

Scenario: Add entries for a past date
  Given the user has bujo installed
  When the user runs `bujo add --date yesterday ". Forgot to log"`
  Then a task is created for yesterday's date

Scenario: Add entries via stdin pipe
  Given the user has bujo installed
  When the user runs `echo ". Piped task" | bujo add`
  Then a task with content "Piped task" is created

Scenario: Add hierarchical entries
  Given the user has bujo installed
  When the user runs:
    """
    bujo add ". Parent task
      - Child note
      - Another child"
    """
  Then three entries are created
  And the child entries have the parent task as their parent
```

#### 1.2 Viewing Entries

```gherkin
Scenario: View today's entries
  Given the user has entries for today
  When the user runs `bujo today`
  Then today's entries are displayed
  And overdue tasks are shown in the overdue section
  And the current location is shown if set

Scenario: List entries for last 7 days
  Given the user has entries for the past week
  When the user runs `bujo ls`
  Then entries for the last 7 days are displayed
  And each day's entries are grouped under a date header

Scenario: List entries for custom date range
  Given the user has entries from January 1 to January 15
  When the user runs `bujo ls --from 2026-01-01 --to 2026-01-07`
  Then only entries from January 1 to January 7 are displayed

Scenario: View entry with context
  Given an entry with ID 42 has a parent and children
  When the user runs `bujo view 42`
  Then the entry is displayed with its parent and siblings
  And the requested entry is highlighted

Scenario: View entry with ancestor context
  Given an entry with ID 42 has a grandparent
  When the user runs `bujo view 42 --up 1`
  Then the grandparent context is also displayed
```

#### 1.3 Modifying Entries

```gherkin
Scenario: Mark entry as done
  Given a task with ID 42 exists
  When the user runs `bujo done 42`
  Then the entry type changes to "done"
  And the message "Marked entry 42 as done" is displayed

Scenario: Undo completion
  Given a completed entry with ID 42 exists
  When the user runs `bujo undo 42`
  Then the entry type changes back to "task"
  And the message "Marked entry 42 as incomplete" is displayed

Scenario: Edit entry content
  Given an entry with ID 42 and content "Old content"
  When the user runs `bujo edit 42 "New content"`
  Then the entry content is updated to "New content"
  And the message "Updated entry #42" is displayed

Scenario: Migrate task to future date
  Given a task with ID 42 scheduled for today
  When the user runs `bujo migrate 42 --to tomorrow`
  Then the original entry is marked as migrated
  And a new task is created for tomorrow
  And the message shows the new entry ID
```

#### 1.4 Deleting Entries

```gherkin
Scenario: Delete entry without children
  Given an entry with ID 42 has no children
  When the user runs `bujo delete 42`
  Then the entry is soft-deleted
  And the message "Deleted entry #42" is displayed

Scenario: Delete entry with children - cascade
  Given an entry with ID 42 has child entries
  When the user runs `bujo delete 42`
  And the user selects option 1 "Delete entry and all children"
  Then the entry and all children are soft-deleted

Scenario: Delete entry with children - reparent
  Given an entry with ID 42 has child entries and a parent
  When the user runs `bujo delete 42`
  And the user selects option 2 "Delete entry and reparent children"
  Then the entry is soft-deleted
  And children are moved to the grandparent

Scenario: Force delete without prompt
  Given an entry with ID 42 has children
  When the user runs `bujo delete 42 --force`
  Then the entry and children are deleted without prompting

Scenario: View deleted entries
  Given entries have been soft-deleted
  When the user runs `bujo deleted`
  Then a list of deleted entries is displayed
  And each entry shows its entity ID for restoration

Scenario: Restore deleted entry
  Given an entry with entity ID "abc123" was deleted
  When the user runs `bujo restore abc123`
  Then the entry is restored with a new internal ID
  And the message "Restored entry #<new-id>" is displayed
```

---

### 2. Habit Tracking

#### 2.1 Viewing Habits

```gherkin
Scenario: View habit tracker
  Given habits exist with logged completions
  When the user runs `bujo habit`
  Then a 7-day sparkline view is displayed
  And each habit shows name, streak, and completion rate

Scenario: View habit tracker monthly
  Given habits exist with logged completions
  When the user runs `bujo habit --month`
  Then a 30-day calendar view is displayed

Scenario: Inspect habit details
  Given a habit "Gym" exists with logs
  When the user runs `bujo habit inspect Gym`
  Then detailed habit information is displayed
  And individual log entries are shown

Scenario: Inspect habit by ID
  Given a habit with ID 1 exists
  When the user runs `bujo habit inspect #1`
  Then the habit details are displayed

Scenario: Inspect habit with date range
  Given a habit "Gym" exists
  When the user runs `bujo habit inspect Gym --from "last month" --to today`
  Then only logs within the date range are shown
```

#### 2.2 Logging Habits

```gherkin
Scenario: Log existing habit
  Given a habit "Gym" exists
  When the user runs `bujo habit log Gym`
  Then a log entry is created for today with count 1
  And the message "Logged: Gym" is displayed

Scenario: Log habit with count
  Given a habit "Water" exists
  When the user runs `bujo habit log Water 8`
  Then a log entry is created with count 8
  And the message "Logged: Water (x8)" is displayed

Scenario: Log new habit with confirmation
  Given no habit "Meditation" exists
  When the user runs `bujo habit log Meditation`
  And the user confirms "y" to create the habit
  Then the habit is created
  And a log entry is created

Scenario: Log new habit auto-create
  Given no habit "Running" exists
  When the user runs `bujo habit log Running --yes`
  Then the habit is created without prompting
  And a log entry is created

Scenario: Log habit for past date
  Given a habit "Gym" exists
  When the user runs `bujo habit log Gym --date yesterday`
  Then a log entry is created for yesterday
  And the message includes "for yesterday"

Scenario: Log habit by ID
  Given a habit with ID 1 exists
  When the user runs `bujo habit log #1`
  Then a log entry is created for habit ID 1
```

#### 2.3 Managing Habits

```gherkin
Scenario: Rename habit
  Given a habit "Excercise" exists (misspelled)
  When the user runs `bujo habit rename Excercise Exercise`
  Then the habit name is updated to "Exercise"

Scenario: Set habit goal
  Given a habit "Water" exists with goal 1
  When the user runs `bujo habit set-goal Water 8`
  Then the habit goal is updated to 8 per day

Scenario: Delete habit
  Given a habit "OldHabit" exists
  When the user runs `bujo habit delete OldHabit`
  And the user confirms deletion
  Then the habit is deleted

Scenario: Undo last habit log
  Given a habit "Gym" has a log for today
  When the user runs `bujo habit undo Gym`
  Then the most recent log is deleted
  And the message confirms the undo

Scenario: Delete specific habit log
  Given a habit log with ID 42 exists
  When the user runs `bujo habit delete-log 42`
  Then the specific log entry is deleted
```

---

### 3. List Management

#### 3.1 Viewing Lists

```gherkin
Scenario: View all lists
  Given lists "Shopping" and "Work Tasks" exist
  When the user runs `bujo list`
  Then both lists are displayed
  And each list shows completion progress (e.g., "3/5 done")

Scenario: Show list items
  Given a list "Shopping" exists with items
  When the user runs `bujo list show Shopping`
  Then all items in the list are displayed
  And completed items are marked appropriately
```

#### 3.2 Creating and Managing Lists

```gherkin
Scenario: Create new list
  Given no list "Groceries" exists
  When the user runs `bujo list create Groceries`
  Then a new list is created
  And the message "Created list #<id>: Groceries" is displayed

Scenario: Create list with spaces in name
  Given no list "Shopping List" exists
  When the user runs `bujo list create "Shopping List"`
  Then a list named "Shopping List" is created

Scenario: Rename list
  Given a list "Shopping" exists
  When the user runs `bujo list rename Shopping Groceries`
  Then the list name is updated to "Groceries"

Scenario: Delete list
  Given a list "OldList" exists
  When the user runs `bujo list delete OldList`
  And the user confirms deletion
  Then the list and all its items are deleted
```

#### 3.3 Managing List Items

```gherkin
Scenario: Add item to list
  Given a list "Shopping" exists
  When the user runs `bujo list add Shopping "Buy milk"`
  Then a task item is added to the list
  And the item ID is printed

Scenario: Add item with type prefix
  Given a list "Notes" exists
  When the user runs `bujo list add Notes "- Important info"`
  Then a note item is added to the list

Scenario: Add item by list ID
  Given a list with ID 1 exists
  When the user runs `bujo list add #1 "Item content"`
  Then an item is added to list ID 1

Scenario: Mark list item done
  Given a list item with ID 42 exists
  When the user runs `bujo list done 42`
  Then the item is marked as complete

Scenario: Undo list item completion
  Given a completed list item with ID 42 exists
  When the user runs `bujo list undo 42`
  Then the item is marked as incomplete

Scenario: Remove item from list
  Given a list item with ID 42 exists
  When the user runs `bujo list remove 42`
  Then the item is removed from the list

Scenario: Move item between lists
  Given list item ID 42 is in list "Shopping"
  And list "Groceries" exists
  When the user runs `bujo list move 42 Groceries`
  Then the item is moved to "Groceries"
```

---

### 4. Day Context

#### 4.1 Work Location

```gherkin
Scenario: Set work location
  Given today has no location set
  When the user runs `bujo work set "Home Office"`
  Then today's location is set to "Home Office"

Scenario: Set location for past date
  Given yesterday has no location set
  When the user runs `bujo work set "Client Site" --date yesterday`
  Then yesterday's location is set to "Client Site"

Scenario: View work location
  Given today's location is "Home Office"
  When the user runs `bujo work inspect`
  Then "Home Office" is displayed

Scenario: Clear work location
  Given today has a location set
  When the user runs `bujo work clear`
  Then today's location is cleared
```

#### 4.2 Mood Tracking

```gherkin
Scenario: Set mood
  Given today has no mood set
  When the user runs `bujo mood set "Feeling great"`
  Then today's mood is recorded

Scenario: View mood
  Given today's mood is "Feeling great"
  When the user runs `bujo mood inspect`
  Then "Feeling great" is displayed

Scenario: Clear mood
  Given today has a mood set
  When the user runs `bujo mood clear`
  Then today's mood is cleared
```

#### 4.3 Weather

```gherkin
Scenario: Set weather
  Given today has no weather set
  When the user runs `bujo weather set "Sunny, 22C"`
  Then today's weather is recorded

Scenario: View weather
  Given today's weather is "Sunny, 22C"
  When the user runs `bujo weather inspect`
  Then "Sunny, 22C" is displayed
```

---

### 5. Backup Management

```gherkin
Scenario: List backups
  Given backups exist in ~/.bujo/backups/
  When the user runs `bujo backup`
  Then all backups are listed with filename, date, and size

Scenario: Create manual backup
  Given the database has data
  When the user runs `bujo backup create`
  Then a new backup file is created
  And the backup path is displayed

Scenario: Automatic backup on startup
  Given no backup exists within 7 days
  When the user runs any bujo command
  Then a backup is automatically created
  And "Creating backup..." is shown on stderr

Scenario: Verify backup integrity
  Given a backup file exists
  When the user runs `bujo backup verify <backup-file>`
  Then the backup is validated
  And integrity status is reported
```

---

### 6. Utility Commands

```gherkin
Scenario: View tomorrow's entries
  Given entries are scheduled for tomorrow
  When the user runs `bujo tomorrow`
  Then tomorrow's entries are displayed

Scenario: View next 7 days
  Given entries exist for the coming week
  When the user runs `bujo next`
  Then entries for the next 7 days are displayed

Scenario: View open tasks
  Given open tasks exist across multiple dates
  When the user runs `bujo tasks`
  Then all incomplete tasks are listed

Scenario: View history
  Given entries exist from past dates
  When the user runs `bujo history`
  Then historical entries are displayed

Scenario: View archived entries
  Given archived entries exist
  When the user runs `bujo archive`
  Then archived entries are displayed

Scenario: Check version
  Given bujo is installed
  When the user runs `bujo version`
  Then the version, commit, and build date are displayed
```

---

## TUI Tests

### 7. Application Startup

```gherkin
Scenario: Launch TUI
  Given the user has bujo installed
  When the user runs `bujo tui`
  Then the TUI application starts
  And the journal view is displayed by default
  And today's entries are loaded

Scenario: Launch TUI with theme
  Given a config file with theme "dark" exists
  When the user runs `bujo tui`
  Then the TUI uses the dark theme colors

Scenario: Launch TUI with default view
  Given a config file with default_view "habits" exists
  When the user runs `bujo tui`
  Then the habits view is displayed by default
```

---

### 8. Journal View

#### 8.1 Navigation

```gherkin
Scenario: Navigate down through entries
  Given the journal view is active with multiple entries
  When the user presses 'j' or Down arrow
  Then the selection moves to the next entry
  And the cursor indicator moves down

Scenario: Navigate up through entries
  Given the selection is on the second entry
  When the user presses 'k' or Up arrow
  Then the selection moves to the previous entry

Scenario: Jump to top
  Given the selection is in the middle of the list
  When the user presses 'g'
  Then the selection jumps to the first entry

Scenario: Jump to bottom
  Given the selection is in the middle of the list
  When the user presses 'G'
  Then the selection jumps to the last entry

Scenario: Scroll maintains visibility
  Given there are more entries than fit on screen
  When the user navigates past the visible area
  Then the view scrolls to keep the selection visible
```

#### 8.2 View Modes

```gherkin
Scenario: Toggle between day and week view
  Given the journal is in day view
  When the user presses 'w'
  Then the view switches to week view
  And entries for 7 days are displayed

Scenario: Toggle back to day view
  Given the journal is in week view
  When the user presses 'w'
  Then the view switches to day view
  And only today's entries are displayed

Scenario: Go to specific date
  Given the journal view is active
  When the user presses '/'
  And enters "2026-01-15"
  Then the view navigates to January 15, 2026
  And entries for that date are displayed
```

#### 8.3 Entry Actions

```gherkin
Scenario: Mark entry as done
  Given a task entry is selected
  When the user presses Space
  Then the entry is marked as done
  And the entry symbol changes to checkmark
  And the entry appears green

Scenario: Delete entry with confirmation
  Given an entry is selected
  When the user presses 'd'
  Then a confirmation prompt appears
  When the user presses 'y'
  Then the entry is deleted
  And the entry list is refreshed

Scenario: Cancel delete
  Given the delete confirmation is shown
  When the user presses 'n' or Esc
  Then the deletion is cancelled
  And the entry remains

Scenario: Edit entry inline
  Given an entry is selected
  When the user presses 'e'
  Then edit mode activates
  And the entry content is editable
  When the user modifies content and presses Enter
  Then the entry is updated

Scenario: Add sibling entry
  Given an entry is selected
  When the user presses 'a'
  Then add mode activates
  When the user types ". New task" and presses Enter
  Then a new task is created at the same level

Scenario: Add child entry
  Given an entry is selected
  When the user presses 'A'
  Then add mode activates for child entry
  When the user types "- Child note" and presses Enter
  Then a child note is created under the selected entry

Scenario: Add root entry
  Given the journal view is active
  When the user presses 'r'
  Then add mode activates for root entry
  When the user types content and presses Enter
  Then a root-level entry is created for the current date

Scenario: Migrate entry
  Given a task is selected
  When the user presses 'm'
  Then migrate mode activates
  When the user enters "tomorrow"
  Then the task is migrated to tomorrow
  And the original shows migration symbol
```

---

### 9. Capture Mode

```gherkin
Scenario: Enter capture mode
  Given the journal view is active
  When the user presses 'c'
  Then capture mode activates
  And a multi-line editor appears

Scenario: Type entries in capture mode
  Given capture mode is active
  When the user types:
    """
    . Task one
    - Note here
      - Sub-note
    """
  Then a real-time preview shows parsed entries
  And indentation creates hierarchy

Scenario: Save capture content
  Given capture mode has content
  When the user presses Ctrl+S
  Then all entries are saved
  And capture mode closes
  And the journal view refreshes

Scenario: Cancel capture with confirmation
  Given capture mode has content
  When the user presses Escape
  Then a confirmation prompt appears "Discard changes?"
  When the user presses 'y'
  Then capture mode closes without saving

Scenario: Resume capture draft
  Given capture mode was exited with unsaved content
  When the user presses 'c' to enter capture mode
  Then the previous draft content is restored

Scenario: Show capture mode help
  Given capture mode is active
  When the user presses F1
  Then the help overlay is displayed
  And available keybindings are shown

Scenario: Detect syntax errors
  Given capture mode is active
  When the user types "Missing symbol"
  Then an error indicator appears
  And the error message explains the issue

Scenario: Emacs navigation in capture mode
  Given capture mode is active with text
  When the user presses Ctrl+A
  Then the cursor moves to beginning of line
  When the user presses Ctrl+E
  Then the cursor moves to end of line
  When the user presses Ctrl+U
  Then text from cursor to line start is deleted
```

---

### 10. Habits View

```gherkin
Scenario: Switch to habits view
  Given the journal view is active
  When the user presses '2'
  Then the habits view is displayed
  And all habits are loaded and shown

Scenario: View habit list
  Given the habits view is active
  And habits "Gym" and "Meditation" exist
  Then both habits are displayed
  And each shows name, streak, and completion indicator

Scenario: Navigate habits
  Given multiple habits are displayed
  When the user presses 'j' or 'k'
  Then the selection moves between habits

Scenario: Log habit completion
  Given a habit is selected
  When the user presses Space
  Then a log entry is created for today
  And the habit display updates immediately

Scenario: View empty habits state
  Given no habits exist
  When the user switches to habits view
  Then a message "No habits yet" is displayed
  And instructions for creating habits are shown
```

---

### 11. Lists View

```gherkin
Scenario: Switch to lists view
  Given the journal view is active
  When the user presses '3'
  Then the lists view is displayed
  And all lists are loaded and shown

Scenario: View list overview
  Given lists "Shopping" and "Work" exist
  When the lists view is displayed
  Then both lists show name and completion progress
  And progress shows "X/Y done" format

Scenario: Navigate lists
  Given multiple lists are displayed
  When the user presses 'j' or 'k'
  Then the selection moves between lists

Scenario: Enter list items view
  Given a list is selected
  When the user presses Enter
  Then the list items view is displayed
  And all items in the list are shown

Scenario: Navigate list items
  Given the list items view is active
  When the user presses 'j' or 'k'
  Then the selection moves between items

Scenario: Toggle list item done
  Given a list item is selected
  When the user presses Space
  Then the item done status toggles
  And the display updates immediately

Scenario: Return to lists from items
  Given the list items view is active
  When the user presses Escape
  Then the view returns to lists overview

Scenario: View empty list
  Given a list with no items is selected
  When the user presses Enter
  Then a message "No items" is displayed
```

---

### 12. Command Palette

```gherkin
Scenario: Open command palette
  Given any view is active
  When the user presses Ctrl+P or ':'
  Then the command palette overlay appears
  And a text input is focused

Scenario: Search commands
  Given the command palette is open
  When the user types "hab"
  Then only commands matching "hab" are shown
  And "Switch to Habits" appears in results

Scenario: Fuzzy search commands
  Given the command palette is open
  When the user types "swh"
  Then "Switch to Habits" matches via fuzzy search

Scenario: Navigate command results
  Given command results are displayed
  When the user presses 'j' or Down arrow
  Then the selection moves to next command

Scenario: Execute command
  Given a command is selected
  When the user presses Enter
  Then the command is executed
  And the palette closes

Scenario: Close command palette
  Given the command palette is open
  When the user presses Escape
  Then the palette closes
  And the previous view is restored

Scenario: View command keybindings
  Given the command palette shows results
  Then each command displays its keybinding
  And the user can learn shortcuts
```

---

### 13. Help System

```gherkin
Scenario: Toggle help display
  Given any view is active
  When the user presses '?'
  Then the help panel toggles visibility
  And available keybindings are shown

Scenario: View short help
  Given the help panel is collapsed
  Then the bottom bar shows essential keybindings
  And "?" shows how to get more help

Scenario: View full help
  Given the user presses '?'
  Then full keybinding reference is displayed
  And keybindings are grouped by function
```

---

### 14. General TUI Behavior

```gherkin
Scenario: Quit application
  Given the TUI is running
  When the user presses 'q' or Ctrl+C
  Then the application exits cleanly
  And the terminal is restored

Scenario: Handle window resize
  Given the TUI is running
  When the terminal window is resized
  Then the UI adapts to new dimensions
  And content remains properly laid out

Scenario: Show loading state
  Given data is being loaded
  Then a loading indicator is displayed
  And the UI remains responsive

Scenario: Show error state
  Given an error occurs during data load
  Then an error message is displayed
  And the user can retry or continue

Scenario: Maintain state between views
  Given entries exist in journal view
  When the user switches to habits and back
  Then the journal data is reloaded
  And selection state is maintained
```

---

### 15. Theming

```gherkin
Scenario: Use default theme
  Given no theme is configured
  When the TUI starts
  Then the default color scheme is applied

Scenario: Use dark theme
  Given theme "dark" is configured
  When the TUI starts
  Then dark theme colors are applied

Scenario: Use light theme
  Given theme "light" is configured
  When the TUI starts
  Then light theme colors are applied

Scenario: Use solarized theme
  Given theme "solarized" is configured
  When the TUI starts
  Then solarized colors are applied
```

---

## Cross-Cutting Tests

### 16. Data Persistence

```gherkin
Scenario: Data persists between sessions
  Given entries were created in a previous session
  When the user starts bujo again
  Then the previous entries are still available

Scenario: Changes reflect immediately
  Given the TUI is open
  When an entry is created via CLI in another terminal
  And the user refreshes the TUI view
  Then the new entry appears

Scenario: Database migrations run automatically
  Given a database from an older version exists
  When the user runs bujo
  Then migrations are applied automatically
  And data is preserved
```

### 17. Date Handling

```gherkin
Scenario: Natural language dates
  Given the user wants to add an entry for "last monday"
  When the user runs `bujo add -d "last monday" ". Task"`
  Then the task is created for the previous Monday

Scenario: Relative dates
  Given the user wants to add an entry for yesterday
  When the user runs `bujo add -d yesterday ". Task"`
  Then the task is created for yesterday's date

Scenario: ISO date format
  Given the user wants a specific date
  When the user runs `bujo add -d 2026-01-15 ". Task"`
  Then the task is created for January 15, 2026

Scenario: Invalid date handling
  Given the user enters an invalid date
  When the user runs `bujo add -d "not a date" ". Task"`
  Then an error message explains the issue
  And no entry is created
```

### 18. Error Handling

```gherkin
Scenario: Handle missing database
  Given the database file does not exist
  When the user runs bujo
  Then the database is created automatically
  And the command succeeds

Scenario: Handle invalid entry ID
  Given no entry with ID 99999 exists
  When the user runs `bujo done 99999`
  Then an error message indicates entry not found

Scenario: Handle permission errors
  Given the database directory is not writable
  When the user runs bujo
  Then a clear error message is displayed
```

---

## Summary

| Category | Test Count |
|----------|------------|
| Entry Management | 18 |
| Habit Tracking | 16 |
| List Management | 14 |
| Day Context | 8 |
| Backup Management | 4 |
| Utility Commands | 6 |
| TUI Startup | 3 |
| Journal View | 17 |
| Capture Mode | 9 |
| Habits View | 5 |
| Lists View | 8 |
| Command Palette | 7 |
| Help System | 3 |
| General TUI | 5 |
| Theming | 4 |
| Cross-Cutting | 8 |
| **Total** | **125** |
