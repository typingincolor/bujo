---
description: Reopen a closed or done issue
allowed-tools: Read, Edit
---

# Reopen Issue

Reopen an issue that was previously marked as done or closed.

## Arguments

`$ARGUMENTS` should contain the issue number (e.g., `90` or `#90`).

## File Format

Issues use explicit IDs: `#ID STATUS DESCRIPTION`
- Open issues are in `docs/issues.txt`
- Closed issues are archived in `docs/issues-archive.txt`

## Instructions

1. Parse the issue number from `$ARGUMENTS` (strip `#` if present)
2. First check `docs/issues.txt` - if found with `.` marker, it's already open
3. Read `docs/issues-archive.txt`
4. Find the line matching `#[NUMBER]`
5. Verify issue is closed (`x` or `~` marker)
6. Remove the line from `docs/issues-archive.txt`
7. Add the line to `docs/issues.txt` with `.` marker: `#[NUMBER] . [description]`
8. Report the change

## Validation

- Issue number must be provided
- Issue must exist in one of the files
- Issue must currently be done (`x`) or closed (`~`) in the archive
- If issue is already open (`.`), report that

## Output

```
Reopened issue #[NUMBER]: [description]
(Moved to issues.txt)
```

Or if already open:
```
Issue #[NUMBER] is already open
```
