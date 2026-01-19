---
description: Reopen a closed or done issue
allowed-tools: Read, Edit
---

# Reopen Issue

Reopen an issue that was previously marked as done or closed in `docs/issues.txt`.

## Arguments

`$ARGUMENTS` should contain the issue number (line number).

## Instructions

1. Parse the issue number from `$ARGUMENTS`
2. Read `docs/issues.txt`
3. Find the line with that number
4. Change the marker from `x` (done) or `~` (won't fix) back to `.` (open)
5. Report the change

## Validation

- Issue number must be provided
- Issue must exist
- Issue must currently be done (`x`) or closed (`~`)
- If issue is already open (`.`), report that

## Output

```
Reopened issue #[NUMBER]: [description]
```

Or if already open:
```
Issue #[NUMBER] is already open
```
