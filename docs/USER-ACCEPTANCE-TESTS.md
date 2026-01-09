# User Acceptance Tests

User acceptance tests for bujo defining expected behavior from a user's perspective.

These tests describe **what should happen**, not implementation details. They should fail when bugs exist.

---

## CLI Tests

### 1. Entry Management

#### Adding Entries

```gherkin
Scenario: Add a task to today's journal
  Given I want to record a task
  When I add ". Buy groceries"
  Then a task "Buy groceries" appears in today's journal

Scenario: Add multiple entries at once
  Given I have several things to log
  When I add ". Task one" "- Note one" "o Event one"
  Then all three entries appear in today's journal
  And each has the correct type (task, note, event)

Scenario: Add entries from a file
  Given I have a file with entries to import
  When I add entries from that file
  Then all entries from the file appear in my journal

Scenario: Add entries with a location
  Given I want to record where I'm working
  When I add an entry with location "Home Office"
  Then the entry is associated with that location

Scenario: Backfill entries for a past date
  Given I forgot to log something yesterday
  When I add an entry for yesterday
  Then the entry appears under yesterday's date

Scenario: Add hierarchical entries
  Given I have a task with sub-items
  When I add a parent entry with indented children
  Then the entries maintain their parent-child relationship
  And children appear nested under their parent
```

#### Viewing Entries

```gherkin
Scenario: See today's journal
  Given I have entries for today
  When I view today's journal
  Then I see all of today's entries
  And I see any overdue tasks from previous days
  And I see my current location if set

Scenario: See recent entries
  Given I have entries from the past week
  When I list recent entries
  Then I see entries grouped by date
  And I can see which tasks are done and which are pending

Scenario: See entries for a date range
  Given I want to review a specific period
  When I list entries from January 1 to January 7
  Then I see only entries from that date range

Scenario: See an entry with its context
  Given an entry has a parent and children
  When I view that entry
  Then I see it with its parent and siblings for context
```

#### Modifying Entries

```gherkin
Scenario: Mark a task as done
  Given I have completed a task
  When I mark it as done
  Then it shows as completed with a checkmark
  And it remains in my journal for that day

Scenario: Undo marking a task as done
  Given I accidentally marked a task done
  When I undo it
  Then it returns to an incomplete task

Scenario: Edit an entry's content
  Given I made a typo in an entry
  When I edit the entry with corrected text
  Then the entry shows the updated content

Scenario: Migrate a task to a future date
  Given I can't complete a task today
  When I migrate it to tomorrow
  Then the original shows as migrated
  And a new task appears on the future date

Scenario: Strikethrough/cancel an entry
  Given I no longer need to do a task
  When I cancel it
  Then it shows as cancelled (strikethrough)
  And it remains visible but clearly not active

Scenario: Change an entry's type
  Given I created a task that should be a note
  When I change its type to note
  Then it displays as a note instead of a task
```

#### Deleting Entries

```gherkin
Scenario: Delete an entry
  Given I want to remove an entry
  When I delete it
  Then it no longer appears in my journal

Scenario: Delete an entry with children
  Given an entry has child entries
  When I delete the parent
  Then I'm asked what to do with children
  And I can choose to delete all or keep children

Scenario: Recover a deleted entry
  Given I accidentally deleted an entry
  When I view deleted entries and restore it
  Then it reappears in my journal

Scenario: See deleted entries
  Given I have deleted some entries
  When I view deleted entries
  Then I see a list of recoverable entries
```

---

### 2. Habit Tracking

#### Viewing Habits

```gherkin
Scenario: See my habit tracker
  Given I have habits I'm tracking
  When I view my habits
  Then I see each habit with its name
  And I see my current streak for each
  And I see my progress toward the goal
  And I see today's completion count
  And I see a monthly history

Scenario: See detailed habit history
  Given I want to review a specific habit
  When I inspect that habit
  Then I see its complete log history
  And I see streak and completion statistics
```

#### Logging Habits

```gherkin
Scenario: Log a habit completion
  Given I completed a habit
  When I log it
  Then today's count increases
  And my streak updates if applicable

Scenario: Log multiple completions
  Given I drank 3 glasses of water
  When I log Water with count 3
  Then today's count shows 3

Scenario: Log a habit for a past date
  Given I forgot to log yesterday's workout
  When I log Gym for yesterday
  Then it appears in yesterday's history

Scenario: Create a new habit by logging it
  Given I want to track a new habit
  When I log a habit that doesn't exist
  Then I'm asked if I want to create it
  And it's created when I confirm
```

#### Managing Habits

```gherkin
Scenario: Rename a habit
  Given I misspelled a habit name
  When I rename it
  Then the new name appears everywhere
  And my history is preserved

Scenario: Set a daily goal for a habit
  Given I want to drink 8 glasses of water daily
  When I set Water's goal to 8
  Then my progress shows against that goal

Scenario: Delete a habit
  Given I no longer want to track a habit
  When I delete it
  Then it no longer appears in my tracker

Scenario: Undo a habit log
  Given I accidentally logged a habit
  When I undo the last log
  Then today's count decreases
```

---

### 3. List Management

#### Viewing Lists

```gherkin
Scenario: See all my lists
  Given I have multiple lists
  When I view my lists
  Then I see each list's name
  And I see completion progress (e.g., "3/5 done")

Scenario: See items in a list
  Given I have a shopping list
  When I view that list
  Then I see all items in the list
  And I can see which are done and which are pending
```

#### Managing Lists

```gherkin
Scenario: Create a new list
  Given I want a new shopping list
  When I create a list named "Shopping"
  Then it appears in my lists

Scenario: Rename a list
  Given I want to rename a list
  When I rename "Shopping" to "Groceries"
  Then it shows the new name

Scenario: Delete a list
  Given I no longer need a list
  When I delete it
  Then it no longer appears
  And all its items are removed
```

#### Managing List Items

```gherkin
Scenario: Add an item to a list
  Given I need to buy milk
  When I add "Milk" to my Shopping list
  Then it appears in that list

Scenario: Mark a list item as done
  Given I bought the milk
  When I mark it done
  Then it shows as completed

Scenario: Unmark a completed item
  Given I marked something done by mistake
  When I toggle it back
  Then it shows as incomplete again

Scenario: Remove an item from a list
  Given I no longer need an item
  When I remove it
  Then it no longer appears in the list

Scenario: Move an item to another list
  Given an item is in the wrong list
  When I move it to the correct list
  Then it appears in the new list
  And is removed from the old list
```

---

### 4. Day Context

```gherkin
Scenario: Set my work location
  Given I want to track where I'm working
  When I set today's location to "Home Office"
  Then it appears in today's journal view

Scenario: Record my mood
  Given I want to track how I'm feeling
  When I set today's mood
  Then it's recorded for that day

Scenario: Record the weather
  Given I want to note the weather
  When I set today's weather to "Sunny, 22C"
  Then it's recorded for that day
```

---

### 5. Backup

```gherkin
Scenario: See my backups
  Given backups have been created
  When I list backups
  Then I see each backup with its date and size

Scenario: Create a backup manually
  Given I want to ensure my data is safe
  When I create a backup
  Then a new backup file is created

Scenario: Automatic backup
  Given I haven't backed up recently
  When I use bujo
  Then a backup is created automatically
```

---

## TUI Tests

### 6. General Navigation

```gherkin
Scenario: Navigate between views using number keys
  Given I'm in the Journal view
  When I press the key for Habits
  Then I see the Habits view

Scenario: Navigate between views using the menu
  Given I want to switch views
  When I open the menu and select Lists
  Then I see the Lists view

Scenario: Open the command palette
  Given I want to find a command
  When I open the command palette
  Then I see a searchable list of all commands
  And I can type to filter commands

Scenario: Execute a command from the palette
  Given the command palette is open
  When I select a command
  Then that command is executed

Scenario: See available keyboard shortcuts
  Given I want to know what I can do
  Then I see major commands in the bottom bar
  When I toggle full help
  Then I see all available shortcuts

Scenario: Quit the application
  Given I'm done using bujo
  When I quit
  Then the application closes cleanly
```

---

### 7. Journal View

#### Viewing Entries

```gherkin
Scenario: See today's entries when opening TUI
  Given I have entries for today
  When I open the TUI
  Then I see today's entries in the Journal view
  And I see any overdue tasks

Scenario: Navigate through entries
  Given there are multiple entries
  When I move up and down
  Then I can select different entries
  And the selected entry is highlighted

Scenario: Scroll through many entries
  Given there are more entries than fit on screen
  When I navigate past the visible area
  Then the view scrolls to keep my selection visible
```

#### Adding Entries

```gherkin
Scenario: Add a quick inline entry
  Given I want to add a single entry
  When I add an entry inline
  Then I can type the entry content
  And it's added to today's journal when I confirm

Scenario: Add multiple entries via capture mode
  Given I have several things to log
  When I enter capture mode
  Then I see a multi-line editor
  When I type multiple entries with hierarchy
  Then I see a preview of parsed entries
  When I save
  Then all entries are added to my journal

Scenario: Capture mode shows syntax errors
  Given I'm in capture mode
  When I type invalid syntax
  Then I see an error indicator
  And I see what's wrong

Scenario: Resume a draft in capture mode
  Given I started capture mode but didn't save
  When I return to capture mode
  Then my previous draft is restored
```

#### Modifying Entries

```gherkin
Scenario: Mark selected entry as done
  Given I have a task selected
  When I mark it done
  Then it shows as completed immediately

Scenario: Edit selected entry
  Given I have an entry selected
  When I edit it
  Then I can modify the content
  And the change is saved when I confirm

Scenario: Delete selected entry
  Given I have an entry selected
  When I delete it
  Then I'm asked to confirm
  When I confirm
  Then it's removed from the view

Scenario: Migrate selected task
  Given I have a task selected
  When I migrate it
  Then I can enter a target date
  And the task shows as migrated
  And a new task appears on the target date

Scenario: Cancel/strikethrough selected entry
  Given I have an entry selected
  When I cancel it
  Then it shows as struck through

Scenario: Change entry type
  Given I have an entry selected
  When I change its type
  Then I can select the new type
  And the entry's symbol changes accordingly
```

#### View Context

```gherkin
Scenario: Capture mode is only available in Journal
  Given I'm in the Journal view
  Then capture mode is available
  When I switch to Habits or Lists view
  Then capture mode is not available
```

---

### 8. Habits View

#### Viewing Habits

```gherkin
Scenario: See all my habits
  Given I have habits being tracked
  When I view the Habits view
  Then I see each habit with:
    | name           |
    | current streak |
    | today's count  |
    | progress       |
    | monthly history|

Scenario: Only active habits are shown
  Given I deleted a habit
  When I view the Habits view
  Then the deleted habit does not appear

Scenario: Habit data is accurate
  Given I logged "Gym" for the last 5 consecutive days
  When I view the Habits view
  Then Gym shows a streak of 5
  And the monthly history shows those 5 days as completed
```

#### Logging Habits

```gherkin
Scenario: Log a habit from the TUI
  Given I have a habit selected
  When I log it
  Then today's count increases by 1
  And the display updates immediately

Scenario: Cannot log deleted habits
  Given a habit was deleted
  Then it does not appear in the view
  And I cannot log it
```

#### Navigation

```gherkin
Scenario: Navigate between habits
  Given there are multiple habits
  When I move up and down
  Then I can select different habits
```

#### Context-Appropriate Options

```gherkin
Scenario: Only habit-relevant commands are shown
  Given I'm in the Habits view
  When I view available commands
  Then I see habit commands (log, view details)
  And I do not see journal commands (capture, migrate)

Scenario: Help shows habit-relevant shortcuts
  Given I'm in the Habits view
  When I view help
  Then shortcuts are relevant to habits
  And capture mode shortcut is not shown
```

---

### 9. Lists View

#### Viewing Lists

```gherkin
Scenario: See all my lists
  Given I have multiple lists
  When I view the Lists view
  Then I see each list's name
  And I see accurate completion counts (e.g., "2/5 done")

Scenario: Completion count is accurate
  Given a list has 5 items with 2 marked done
  When I view the Lists view
  Then that list shows "2/5 done"

Scenario: Only active lists are shown
  Given I deleted a list
  When I view the Lists view
  Then the deleted list does not appear

Scenario: Open a list to see its items
  Given I have a list selected
  When I open it
  Then I see all items in that list
```

#### Context-Appropriate Options

```gherkin
Scenario: Only list-relevant commands are shown
  Given I'm in the Lists view
  When I view available commands
  Then I see list commands
  And I do not see capture mode
  And I do not see journal-specific commands
```

---

### 10. List Items View

#### Viewing Items

```gherkin
Scenario: See all items in a list
  Given a list has items
  When I view that list
  Then I see all items
  And done items show as completed
  And incomplete items show as pending

Scenario: Only active items are shown
  Given I deleted an item from a list
  When I view that list
  Then the deleted item does not appear

Scenario: Empty list shows helpful message
  Given a list has no items
  When I view that list
  Then I see a message that it's empty
  And I see how to add items
```

#### Managing Items

```gherkin
Scenario: Add an item to the list
  Given I'm viewing a list
  When I add a new item
  Then I can type the item content
  And it appears in the list when I confirm

Scenario: Mark an item as done
  Given I have an item selected
  When I toggle its done status
  Then it shows as completed
  And the list's completion count updates

Scenario: Unmark a done item
  Given I have a completed item selected
  When I toggle its done status
  Then it shows as incomplete again

Scenario: Edit an item
  Given I have an item selected
  When I edit it
  Then I can modify its content
  And the change is saved when I confirm

Scenario: Delete an item
  Given I have an item selected
  When I delete it
  Then it's removed from the list
  And the list's total count updates

Scenario: Navigate between items
  Given there are multiple items
  When I move up and down
  Then I can select different items

Scenario: Return to lists view
  Given I'm viewing a list's items
  When I go back
  Then I return to the Lists view
```

#### Context-Appropriate Options

```gherkin
Scenario: Only item-relevant commands are shown
  Given I'm viewing list items
  When I view available commands
  Then I see item commands (add, edit, done, delete)
  And I do not see capture mode
  And I do not see migrate (lists don't have dates)
```

---

### 11. Search View

```gherkin
Scenario: Search for entries
  Given I want to find entries containing "project"
  When I search for "project"
  Then I see matching entries from all dates

Scenario: Search results show context
  Given search results are displayed
  Then each result shows its date
  And each result shows its type
  And I can select a result to view it
```

---

### 12. Summary/Stats View

```gherkin
Scenario: See productivity overview
  Given I have entries and habits logged
  When I view the Summary
  Then I see statistics about my productivity
  And I see habit completion trends

Scenario: See AI reflections
  Given AI summaries are enabled
  When I view the Summary
  Then I see AI-generated reflections on my journal
```

---

### 13. Settings View

```gherkin
Scenario: View current settings
  Given I want to see my configuration
  When I open Settings
  Then I see current theme, default view, etc.

Scenario: Change a setting
  Given I want to change my theme
  When I select a different theme in Settings
  Then the change is applied immediately
```

---

### 14. Error Handling

```gherkin
Scenario: Handle errors gracefully
  Given something goes wrong (e.g., database issue)
  When an error occurs
  Then I see a clear error message
  And the app doesn't crash
  And I can continue using other features

Scenario: Handle empty states
  Given I have no entries/habits/lists yet
  When I view that section
  Then I see a helpful message
  And I see how to add my first item
```

---

### 15. Data Accuracy

```gherkin
Scenario: Deleted items never appear
  Given I deleted an entry, habit, or list item
  When I view any part of the application
  Then deleted items are never shown

Scenario: Counts are always accurate
  Given items have been added, completed, or deleted
  When I view counts (list items, habit logs, etc.)
  Then the counts reflect the actual current state

Scenario: Changes persist
  Given I made changes (add, edit, delete, done)
  When I close and reopen the application
  Then all my changes are still there

Scenario: Changes appear immediately
  Given I make a change
  Then the display updates immediately
  And I don't need to manually refresh
```

---

## Summary

| Category | Scenarios |
|----------|-----------|
| **CLI** | |
| Entry Management | 16 |
| Habit Tracking | 12 |
| List Management | 13 |
| Day Context | 3 |
| Backup | 3 |
| **TUI** | |
| General Navigation | 6 |
| Journal View | 14 |
| Habits View | 8 |
| Lists View | 5 |
| List Items View | 10 |
| Search View | 2 |
| Summary/Stats View | 2 |
| Settings View | 2 |
| Error Handling | 2 |
| Data Accuracy | 4 |
| **Total** | **102** |
