---
description: Reopen a closed GitHub issue
allowed-tools: Bash(gh issue reopen *), Bash(gh issue view *)
---

# Reopen Issue

Reopen a GitHub issue that was previously closed.

## Arguments

`$ARGUMENTS` should contain the GitHub issue number (e.g., `359` or `#359`).

## Instructions

1. Parse the issue number from `$ARGUMENTS` (strip `#` if present)
2. View the issue to get details: `gh issue view [NUMBER]`
3. If already open, report that
4. Reopen the issue: `gh issue reopen [NUMBER]`
5. Report the change

## Validation

- Issue number must be provided
- Issue must exist on GitHub
- If issue is already open, report that

## Output

```
Reopened issue #[NUMBER]: [title]
```

Or if already open:
```
Issue #[NUMBER] is already open
```
