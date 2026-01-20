---
description: Close a GitHub issue as won't fix
allowed-tools: Bash(gh issue close *), Bash(gh issue view *)
---

# Close Issue (Won't Fix)

Close a GitHub issue as "not planned" (won't fix).

## Arguments

`$ARGUMENTS` should contain the GitHub issue number (e.g., `359` or `#359`).

## Instructions

1. Parse the issue number from `$ARGUMENTS` (strip `#` if present)
2. View the issue to get details: `gh issue view [NUMBER]`
3. Close the issue as not planned: `gh issue close [NUMBER] --reason "not planned"`
4. Report the change

## Validation

- Issue number must be provided
- Issue must exist on GitHub
- If issue is already closed, report that

## Output

```
Closed issue #[NUMBER] as won't fix: [title]
```

Or if already closed:
```
Issue #[NUMBER] is already closed
```
