# Technical Specification: bujo CLI (typingincolor)

## 1. Project Overview
`bujo` is a macOS CLI tool for Bullet Journaling. It is built using **Go**, **SQLite**, and the **Gemini API**. It is designed to be extensible, maintainable, and context-aware.

### Architecture Principles
* **TDD-First:** 100% logic coverage before implementation.
* **Hexagonal Architecture:** Domain-driven design with interfaces for storage and AI services.
* **Persistence:** SQLite database stored at `~/.config/bujo/bujo.db`.

## 2. Core Functionality

### A. Hierarchical Bujo Entries
The app parses symbols and indentation to create parent-child relationships.
* **Symbols:** `.` (Task), `-` (Note), `o` (Event), `x` (Done), `>` (Migrated).
* **Nesting:** Indentation (2 spaces or 1 tab) indicates a sub-item.
* **Input:** Support for positional arguments AND piped input (e.g., `cat meeting.txt | bujo add`).

### B. Location & Context Tracking
Tracks the user's physical/work context for the day.
* **Command:** `bujo work "Manchester Office"`, `bujo work "Amsterdam Office"`, `bujo work "Home"`.
* **Behavior:** Every entry logged is associated with the active location.
* **Display:** `bujo ls` shows the current location in the header.

### C. Habit Tracker
Supports tracking habits that may occur multiple times per day.
* **Definition:** `bujo habit new "Water" --goal 8`.
* **Logging:** `bujo habit log "Water"` (increments count for today).
* **Visualization:** * `bujo habit`: Weekly sparkline for all habits.
  * `bujo habit map <name>`: Monthly ASCII heatmap (GitHub-style grid).

### D. Rolling AI Summaries (Gemini)
Summaries are hierarchical to manage context windows and provide long-term insights.
* **Daily:** Summarizes entries + location context.
* **Weekly:** Summarizes the 7 daily summaries.
* **Quarterly/Annual:** Summarizes the previous level's summaries.
* **Prompting:** AI is instructed to find patterns related to location and habits.

## 3. Data Schema

### **Table: entries**
`id`, `type`, `content`, `parent_id`, `location`, `scheduled_date`, `created_at`

### **Table: habits & habit_logs**
`habits`: `id`, `name`, `goal_per_day`
`habit_logs`: `id`, `habit_id`, `completed_at`

### **Table: day_context**
`date`, `location`, `mood`, `weather`

### **Table: summaries**
`id`, `horizon`, `content`, `start_date`, `end_date`

## 4. UI/UX Requirements
* **`bujo ls`**: Displays a tree view of today's items + overdue tasks.
* **`bujo week`**: A 7-day agenda view.
* **`bujo month`**: A monthly grid of events and tasks.
* **`bujo summary --horizon [day|week|quarter|year]`**: Triggers Gemini reflection.

## 5. Implementation Roadmap
1. **Domain:** Implement `Entry` and `Habit` structs with a `TreeParser` (TDD).
2. **Repository:** Implement SQLite storage with migrations.
3. **CLI:** Implement `cobra` or `urfave/cli` for command handling and `os.Stdin` support.
4. **Services:** Integrate Gemini API with a rolling summary logic.
