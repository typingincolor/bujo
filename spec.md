Technical Specification: bujo (typingincolor)
1. Project Vision
bujo is a high-performance, Go-based command-line Bullet Journal for macOS. It acts as a "Life Log" that captures tasks, notes, events, habits, and locations. It uses TDD, Hexagonal Architecture, and Gemini AI to provide deep, rolling reflections on a user's life.

2. Architectural Principles
Web-Ready Hexagonal Design: Business logic is isolated in internal/domain. The CLI and a future Web Server are merely "adapters" to this shared logic.

Stateless Service Layer: All "rules" (parsing, habit logic, AI prompts) reside in service structs. The app state persists entirely in SQLite.

12-Factor Lite: Configuration via environment variables (GEMINI_API_KEY, DB_PATH), and strict dependency management via go.mod.

Logging and Streams: Diagnostic messages (errors, status updates) MUST go to stderr. Only actual data (the journal tree, habits, summaries) should go to stdout.

TDD Mandate: 100% logic coverage in internal/domain before implementation.

3. UX & CLI Interface Examples
A. The Daily Agenda (bujo ls)
Displays current location, overdue items, and today's hierarchical notes.

Plaintext

📅 Tuesday, Jan 6, 2026 | 📍 Manchester Office
---------------------------------------------------------
OVERDUE
 [ ]  12  . Finish project proposal (from 2026-01-04)

TODAY
 [ ]  45  . Call the bank regarding mortgage
 [x]  46  x Morning meditation
 [ ]  47  o Team Sync @ 10:00 AM
 [ ]  48  o Project Alpha Kickoff
          ├── - Attendees: Alice, Bob, Charlie
          └── [ ] . Send follow-up email to Alice
              └── - Include the PDF attachment
---------------------------------------------------------
B. Habit Tracking (bujo habit)
Displays a 7-day sparkline status and the monthly "GitHub-style" heatmap.

Bash

# Weekly Sparkline View
$ bujo habit
🔥 HABIT TRACKER (Last 7 Days)
---------------------------------------------------------
Gym    [X] [ ] [X] [X] [ ] [X] [X]  (5/7)  STREAK: 2
Water  [8] [7] [8] [8] [6] [8] [4]  (Avg: 7.1)
Med    [X] [X] [X] [X] [X] [X] [X]  (7/7)  STREAK: 14
---------------------------------------------------------

# Monthly Heatmap
$ bujo habit map Gym
GYM: JANUARY 2026
      S M T W T F S
W1      X . X X . X
W2    X X . X . . . 
W3    . . . . . . .
(X = Goal Met, . = Missed)
C. AI Reflections (bujo summary --weekly)
Plaintext

🤖 WEEKLY REFLECTION (Jan 5 - Jan 11)
---------------------------------------------------------
CORE THEMES:
• Heavy focus on financial admin and software architecture.
• High correlation between "Home" location and "Deep Work" notes.

AI INSIGHT:
"You've completed 80% of tasks this week. However, when working from the 
Manchester Office, your 'Note' volume increases by 40% while 'Task' 
completion drops. Consider Manchester for brainstorming and Home for execution."
---------------------------------------------------------
4. Core Features
A. Hierarchical Tree Parser
Support for nested items using indentation (2 spaces or 1 tab).

Symbols: . (Task), - (Note), o (Event), x (Done), > (Migrated).

Logic: Piped input or multi-line arguments are parsed into a parent-child tree structure.

Storage: entries table uses parent_id and depth for recursion.

B. Location & Context Tracking
Every day has a primary location context.

Command: bujo work <location> updates the day_context table.

Metadata: Entries logged inherit the active location unless overridden via --at.

C. Multi-Log Habit Tracker
A quantitative tracker for recurring habits.

Command: bujo habit log <name> [--count N] [--date YYYY-MM-DD]

Feedback: Status updates (e.g., [LOG] Habit 'Water' recorded (3/8 today)) are sent to stderr.

D. Rolling AI Summaries (Gemini)
Hierarchical summarization to manage context windows and find long-term trends.

Flow: Daily -> Weekly -> Quarterly -> Annual.

Storage: Summaries are cached in a dedicated summaries table to minimize API usage.

5. Data Schema (SQLite)
entries: id, type, content, parent_id (FK), location, scheduled_date, created_at

habits & habit_logs: habits (goal_per_day), habit_logs (completed_at)

day_context: date (PK), location, mood, weather

summaries: id, horizon, content, start_date, end_date

6. Implementation Roadmap
Domain Layer: Define core types (Entry, Habit, Summary) and the TreeParser with 100% TDD.

Service Layer: Build BujoService and HabitService to coordinate domain logic.

Infrastructure: SQLite repository implementation using golang-migrate.

Adapter (CLI): Build the CLI using cobra, ensuring stderr for logging and stdout for data.

Adapter (AI): Implement the Gemini integration with the hierarchical rolling logic.
