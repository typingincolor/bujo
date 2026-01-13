# Issue Automation

This repository includes automated issue enrichment to help with quick issue creation during development.

## How It Works

### 1. Issue Templates

When creating a new issue, you can choose:

- **Quick Issue** - For rapid captures during the day
  - Just write a brief note
  - Bot automatically adds context from codebase
  - Great for "future me" to pick up later

- **Feature Request** - For more detailed planning
  - Structured fields guide you through details
  - Optional architecture layer selection

- **Bug Report** - For tracking issues
  - Steps to reproduce
  - Expected vs actual behavior

### 2. Automated Enrichment

When you create an issue, the **Issue Enrichment** workflow automatically:

#### 📁 Finds Relevant Files
Searches codebase for related files based on keywords in your issue:
- "habit" → finds `habit.go`, `habit_service.go`, etc.
- "tui" → finds files in `internal/tui/`
- "summary" → finds AI/summary related code

#### 🏗️ Suggests Implementation Layer
Analyzes your issue text and suggests where to start:
- Domain layer for new types/models
- Service layer for business logic
- TUI layer for interface changes
- CLI layer for new commands
- Repository layer for database changes

#### ✅ Adds TDD Checklist
Includes standard TDD workflow:
- Write failing test (RED)
- Implement minimum code (GREEN)
- Refactor
- Run full test suite

#### 💡 Provides Quick Start
Adds helpful commands:
- How to search for related code
- Which test package to run
- Relevant grep commands

### 3. Example Workflow

```
You (during the day):
  Create issue: "Add undo feature"

Bot (within 30 seconds):
  💬 Comments with:
  - Files that might need changes
  - Suggested layer (probably service + tui)
  - TDD checklist
  - Command to search for related code

You (later):
  Review bot's analysis
  Click into suggested files
  Follow TDD checklist
  Implement feature
```

## Benefits

1. **Quick Capture** - Don't break flow during development
2. **Context Recovery** - Bot does initial research for you
3. **Consistency** - Every issue gets same analysis treatment
4. **Learning** - Reinforces architecture patterns

## Customization

Edit `.github/workflows/issue-enrichment.yml` to:
- Add more keyword detection
- Change analysis format
- Adjust file search patterns
- Add custom checklists

## Disabling

To skip enrichment for a specific issue:
- Add `skip-enrichment` label when creating issue
- Or delete the bot's comment afterward
