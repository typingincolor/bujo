# Advanced Automation Features

This document describes the advanced automation features added to the bujo repository.

## 🤖 AI-Powered Workflows

These workflows require `ANTHROPIC_API_KEY` to be set in repository secrets.

### 1. AI Code Review (`ai-code-review.yml`)

**Triggers:** PR opened/updated with Go changes
**Purpose:** Automated code review focusing on project-specific patterns

**What it checks:**
- ✅ TDD Compliance - Were tests written first?
- ✅ Hexagonal Architecture - Proper layer separation?
- ✅ Event Sourcing - Following event sourcing pattern?
- ✅ Functional Patterns - Immutability, pure functions, early returns?
- ✅ Test Quality - Testing behavior not implementation?

**How it works:**
1. Fetches PR diff
2. Reads `.claude/CLAUDE.md` for project context
3. Sends to Claude API for analysis
4. Posts review as PR comment (updates existing comment on re-runs)

**Example output:**
```markdown
## 🤖 AI Code Review

✅ Great TDD compliance - tests added before implementation
⚠️ Consider using value receivers in domain layer for immutability
✅ Event sourcing pattern correctly implemented in repository

internal/domain/habit.go:45 - Consider making this a pure function
```

### 2. Auto-Generate ADRs (`auto-adr.yml`)

**Triggers:** PR merged with architectural changes
**Purpose:** Automatically document architectural decisions

**When it runs:**
- PR title contains "refactor"
- PR has `architecture` or `breaking` label
- PR modifies files in domain/service/repository layers

**What it does:**
1. Analyzes PR commits and changes
2. Generates Architecture Decision Record using AI
3. Numbers it sequentially (0001, 0002, etc.)
4. Commits to `docs/adr/` directory
5. Comments on PR

**ADR Format:**
```markdown
# 0004. Enum Pattern for View State

**Date:** 2026-01-13
**Status:** Accepted

## Context
[What problem motivated this decision]

## Decision
[What we decided to do]

## Consequences
- Positive: Type safety, extensibility
- Negative: More boilerplate
- Neutral: Standard Go pattern

## Related
- PR #159
```

### 3. Smart PR Descriptions (`smart-pr-description.yml`)

**Triggers:** PR opened with minimal/no description
**Purpose:** Auto-generate comprehensive PR descriptions

**What it generates:**
```markdown
## Summary
Implements quarterly view for habit tracker using enum pattern

## Changes
- Added HabitViewMode enum in internal/tui/model.go
- Updated view cycle logic to support 3 states
- Added tests for tri-state cycling

## Architecture
- **Layers affected:** Domain, TUI
- **Pattern:** Enum pattern for state machines

## Testing
- TestHabitViewModeToggle
- TestQuarterlyViewRendering

## Related Issues
Closes #151
```

**Fallback:** If `ANTHROPIC_API_KEY` not set, uses template-based description

## 📊 Analytics & Metrics

### 4. Batch Analytics (`batch-analytics.yml`)

**Triggers:** Weekly (Sunday) + manual dispatch
**Purpose:** Track batch workflow statistics

**What it analyzes:**
- Number of batches completed
- Average PRs per batch
- Batch duration (time from first to last PR)
- Issues closed per batch
- Most frequently changed files
- Commit type distribution

**Output:** Generates `docs/analytics/batch-report.md`

**Example report:**
```markdown
# Batch Analytics Report

**Total Batches:** 12
**Average PRs per Batch:** 3.4
**Average Duration:** 2.3 days

## Recent Batches
| Batch | PRs | Issues | Duration |
|-------|-----|--------|----------|
| H+I   | 6   | 12     | 1.2d     |
| G     | 4   | 3      | 3.5d     |

## Most Changed Files
| File | Changes |
|------|---------|
| internal/tui/model.go | 23 |
| internal/tui/view.go | 18 |
```

### 5. Performance Benchmarking (`benchmark.yml`)

**Triggers:** PR with Go changes
**Purpose:** Detect performance regressions

**What it does:**
1. Runs `go test -bench` on PR code
2. Runs same benchmarks on base branch
3. Uses `benchstat` to compare results
4. Posts comparison as PR comment
5. Warns on statistically significant slowdowns

**Example output:**
```markdown
## ⚡ Performance Benchmark Results

### Comparison vs Base Branch

name                old time/op  new time/op  delta
TreeParser-8         125ns ± 2%   120ns ± 1%  -4.00%  (p=0.000 n=10+10)
EntryValidation-8    89.3ns ± 1%  87.2ns ± 2%  -2.35%  (p=0.001 n=10+10)

✅ No significant performance regressions detected.
```

### 6. Test Coverage Trends (`coverage-check.yml`)

**Triggers:** PR with Go changes
**Purpose:** Track coverage changes and prevent regressions

**What it does:**
1. Runs tests with coverage on PR
2. Runs tests with coverage on base branch
3. Calculates diff
4. Posts per-package breakdown
5. **Fails CI if coverage drops >1%**

**Example output:**
```markdown
## 📊 Test Coverage Report

| Metric | Value |
|--------|-------|
| **Base Coverage** | 87.3% |
| **PR Coverage** | 88.1% |
| **Difference** | +0.8% |
| **Status** | increased |

### Top Packages by Coverage
| Package | Coverage |
|---------|----------|
| internal/domain | 98.5% |
| internal/service | 85.2% |
| internal/tui | 76.4% |

### ✅ Great Job!
Coverage improved! This PR adds valuable test coverage.
```

**Enforcement:**
- Coverage drop <1%: Warning only
- Coverage drop ≥1%: **CI fails** ❌

## Setup Requirements

### Required Secrets

Add to repository secrets (Settings → Secrets → Actions):

```bash
ANTHROPIC_API_KEY=sk-ant-...  # For AI features (1, 2, 3)
```

**Note:** If `ANTHROPIC_API_KEY` is not set:
- AI Code Review: Skips review
- Auto-ADR: Skips ADR generation
- Smart PR Description: Uses template fallback

### Optional Configuration

#### Disable specific workflows

To disable a workflow, add this to the workflow file:
```yaml
on:
  workflow_dispatch:  # Manual only
  # Remove automatic triggers
```

#### Adjust thresholds

**Coverage threshold** (coverage-check.yml):
```yaml
# Change -1.0 to desired threshold
if (( $(echo "$DIFF < -1.0" | bc -l) )); then
```

**Benchmark regression detection** (benchmark.yml):
- Controlled by `benchstat` statistical significance
- No manual threshold needed

## Workflow Summary

| Workflow | Trigger | Requires API Key | Posts Comment | Fails CI |
|----------|---------|------------------|---------------|----------|
| AI Code Review | PR opened/updated | ✅ | ✅ | ❌ |
| Auto-ADR | PR merged (arch changes) | ✅ | ✅ | ❌ |
| Smart PR Description | PR opened (empty body) | Optional | ❌ | ❌ |
| Batch Analytics | Weekly + manual | ❌ | ❌ | ❌ |
| Benchmarking | PR with Go changes | ❌ | ✅ | ❌ |
| Coverage Check | PR with Go changes | ❌ | ✅ | ✅ (if >1% drop) |

## Benefits

### For Solo Development
- **AI Code Review** catches TDD violations you might miss
- **Auto-ADR** documents decisions without manual work
- **Coverage Check** enforces test discipline
- **Batch Analytics** shows workflow patterns over time

### For Team Development
- **Smart PR Description** ensures consistent PR format
- **Benchmarking** prevents performance regressions
- **Coverage trends** maintain code quality baseline
- **ADRs** provide historical context for decisions

## Examples

### Example 1: Opening a PR

```bash
# You create PR with minimal description
gh pr create --title "feat: add quarterly view"

# Smart PR Description triggers:
→ Analyzes commits: "add enum", "update tests", "refactor view"
→ Generates comprehensive description with sections
→ Updates PR automatically

# AI Code Review triggers:
→ Analyzes diff
→ Checks: "✅ Tests added before implementation"
→ Comments: "Great TDD compliance!"

# Coverage Check triggers:
→ Base: 87.3%
→ PR: 88.1%
→ Comments: "+0.8% coverage increase ✅"

# Benchmarking triggers:
→ Runs benchmarks
→ No regressions found
→ Posts results
```

### Example 2: Merging Architectural PR

```bash
# You merge PR with "refactor: switch to enum pattern"
gh pr merge 159

# Auto-ADR triggers:
→ Detects domain/service layer changes
→ Generates ADR from commits
→ Commits to docs/adr/0004-enum-pattern.md
→ Comments on PR: "📝 ADR created"
```

### Example 3: Weekly Analytics

```bash
# Every Sunday at midnight (or manual trigger)
gh workflow run batch-analytics.yml

# Batch Analytics runs:
→ Analyzes last 100 merged PRs
→ Finds batches A through I
→ Calculates statistics
→ Generates docs/analytics/batch-report.md
→ Commits report
```

## Troubleshooting

### AI Code Review not posting
- Check `ANTHROPIC_API_KEY` is set in secrets
- Verify PR has `.go` file changes
- Check workflow run logs for API errors

### Coverage check failing unexpectedly
- Base branch may have had coverage decrease
- Rebuild base branch coverage: merge main, rerun
- Check if tests are flaky (coverage variance)

### Benchmarks showing false regressions
- Benchmarks sensitive to system load
- Re-run workflow to confirm
- Check if changes expected to impact performance

### ADR not generated
- PR must be merged (not just closed)
- Must have architectural changes (domain/service/repository)
- Must have `refactor` in title OR `architecture`/`breaking` label

## Future Enhancements

Potential additions:
- **Mutation testing tracking** - Trend mutation scores
- **Security scanning alerts** - CVE notifications
- **Dependency update PRs** - Auto-merge low-risk deps
- **Release notes AI** - Better categorization
- **Code complexity metrics** - Track cyclomatic complexity
