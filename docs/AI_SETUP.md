# AI / Insights

bujo uses a local SQLite-based insights system rather than live AI API calls.

## How It Works

Insights are pre-computed analysis data stored in a separate database (`~/.bujo/claude-insights.db`). When this file is present, bujo loads read-only insights and surfaces them in the TUI and desktop app.

## Architecture

- Domain types: `internal/domain/insights.go`
- Repository: `internal/repository/sqlite/insights_repository.go`
- TUI integration: `internal/tui/` (messages, model, views)
- Wails integration: `internal/adapter/wails/app.go`

## Previous AI Approaches

Earlier documentation referenced Gemini API integration, Ollama local models, and `bujo summary` commands. These were removed during the insights redesign. The `summaries` table was dropped in migration 000031.
