# Technical Specification: bujo CLI (typingincolor)

## 1. Project Overview

**bujo** is a macOS CLI tool for Bullet Journaling. It is built using Go, SQLite, and the Gemini API. It is designed to be extensible, maintainable, and context-aware.

### Architecture Principles

- **TDD-First**: 100% logic coverage before implementation.
- - **Web-Ready Hexagonal Design**: The core logic has been refactored from CLI-specific to a shared service layer, enabling dual entry points (CLI and future Web Server).
  -   - **Decoupled Logic**: The "Brain" (parsing, habit calculations, AI logic) lives in `internal/domain`.
      -   - **Dual Entry Points**: Two `cmd/` packages support both CLI and Web Server without code duplication.
          -   - **Dependency Injection**: Interfaces for Database and AI services make it trivial to swap components or add new frontends.
              - - **Persistence**: SQLite database stored at `~/.config/bujo/bujo.db`.
                - - **12-Factor Lite**: API keys and DB paths managed through environment variables; stateless architecture ensures CLI and Web Server can run side-by-side.
                 
                  - ## 2. Core Functionality
                 
                  - ### A. Hierarchical Bujo Entries with Deep Nesting
                 
                  - The app parses symbols and indentation to create parent-child relationships, with support for complex meeting notes and deep task nesting.
                 
                  - - **Symbols**: `.` (Task), `-` (Note), `o` (Event), `x` (Done), `>` (Migrated).
                    - - **Nesting**: Indentation (2 spaces or 1 tab) indicates a sub-item.
                      - - **Parent-Child Relationships**: Database schema includes `parent_id` and `depth` columns for hierarchical support.
                        - - **Recursive Parser**: The app preserves tree structure when parsing indented text blocks (e.g., from meeting notes).
                          - - **Tree Rendering**: UI displays notes using visual branch connectors (e.g., `├──`).
                            - - **Input**: Support for positional arguments AND piped input (e.g., `cat meeting.txt | bujo add`).
                             
                              - ### B. Location & Context Tracking (Where-Aware)
                             
                              - Tracks the user's physical/work context for the day, adding a new dimension to journals.
                             
                              - - **Command**: `bujo work "Manchester"`, `bujo work "Amsterdam"`, `bujo work "Home"`.
                                - - **Day Context**: A new `day_context` table tracks location, mood, and weather data.
                                  - - **Behavior**: Every entry logged is associated with the active location.
                                    - - **Display**: `bujo ls` shows the current location in the header.
                                      - - **AI Metadata**: Location is passed to Gemini, allowing AI to notice patterns like "20% more productive on coding tasks when working from Home."
                                       
                                        - ### C. Habit Tracker with Multi-Log & Visuals
                                       
                                        - Upgraded from simple "Yes/No" to a quantitative system supporting multiple logs per day.
                                       
                                        - - **Definition**: `bujo habit new "Water" --goal 8`.
                                          - - **Multi-Logging**: Log the same habit multiple times per day (e.g., tracking 8 glasses of water).
                                            - - **Logging**: `bujo habit log "Water"` (increments count for today).
                                              - - **Visualization**:
                                                -   - `bujo habit`: Weekly sparkline for all habits.
                                                    -   - `bujo habit map <name>`: Monthly ASCII heatmap (GitHub-style contribution grid).
                                                     
                                                        - ### D. Rolling AI Summaries (Gemini)
                                                     
                                                        - Hierarchical summary logic solves the "too much data" problem for LLMs.
                                                     
                                                        - - **Tiered Processing**: Instead of sending 365 days of notes to Gemini for annual summaries, the app summarizes: days → weeks → quarters → year.
                                                          - - **Persistence**: Summaries are stored in the DB, making them instant to retrieve and cheaper to generate.
                                                            - - **Daily**: Summarizes entries + location context.
                                                              - - **Weekly**: Summarizes the 7 daily summaries.
                                                                - - **Quarterly/Annual**: Summarizes the previous level's summaries.
                                                                  - - **Prompting**: AI is instructed to find patterns related to location and habits.
                                                                   
                                                                    - ## 3. Data Schema
                                                                   
                                                                    - ### Table: `entries`
                                                                    - ```
                                                                      id, type, content, parent_id, depth, location, scheduled_date, created_at
                                                                      ```
                                                                      - **parent_id**: Enables hierarchical relationships.
                                                                      - - **depth**: Tracks nesting level for efficient tree traversal.
                                                                        - - **location**: Associates each entry with a context.
                                                                         
                                                                          - ### Table: `day_context`
                                                                          - ```
                                                                            date, location, mood, weather
                                                                            ```

                                                                            ### Table: `habits` & `habit_logs`
                                                                            ```
                                                                            habits: id, name, goal_per_day
                                                                            habit_logs: id, habit_id, count, completed_at
                                                                            ```

                                                                            ### Table: `summaries`
                                                                            ```
                                                                            id, horizon, content, start_date, end_date
                                                                            ```

                                                                            ## 4. UI/UX Requirements

                                                                            - **`bujo ls`**: Displays a tree view of today's items + overdue tasks, with branch connectors (├──).
                                                                            - - **`bujo week`**: A 7-day agenda view.
                                                                              - - **`bujo month`**: A monthly grid of events and tasks.
                                                                                - - **`bujo habit`**: Weekly sparkline visualization for all habits.
                                                                                  - - **`bujo habit map <name>`**: Monthly ASCII heatmap (GitHub-style).
                                                                                    - - **`bujo summary --horizon [day|week|quarter|year]`**: Triggers Gemini reflection with pattern analysis.
                                                                                     
                                                                                      - ## 5. Implementation Roadmap
                                                                                     
                                                                                      - 1. **Domain**: Implement Entry and Habit structs with recursive TreeParser (TDD). Ensure 100% test coverage.
                                                                                        2. 2. **Repository**: Implement SQLite storage with migrations for entries, day_context, habits, summaries, and audit trails.
                                                                                           3. 3. **CLI**: Implement cobra or urfave/cli for command handling with os.Stdin support and tree rendering.
                                                                                              4. 4. **Web Server**: Implement HTTP endpoints (future) using shared service layer via dependency injection.
                                                                                                 5. 5. **Services**: Integrate Gemini API with rolling summary logic and pattern detection.
                                                                                                    6. 6. **Config**: Environment-based configuration (DB_PATH, GEMINI_API_KEY, etc.) following 12-Factor principles.
                                                                                                      
                                                                                                       7. ## 6. Engineering Standards
                                                                                                      
                                                                                                       8. - **12-Factor Lite**: Config via environment variables; stateless app logic.
                                                                                                          - - **Dependency Injection**: All external dependencies (DB, AI) are injected via interfaces.
                                                                                                            - - **Testing**: TDD-first approach with 100% logic coverage.
                                                                                                              - - **Documentation**: Inline code comments and README for quick onboarding.
