---
description: Add a new issue to docs/issues.txt
allowed-tools: Read, Edit
---

# Add Issue

Add a new issue to `docs/issues.txt`.

## Arguments

`$ARGUMENTS` contains the issue description.

## File Format

Issues use explicit IDs: `#ID STATUS DESCRIPTION`
- `#` prefix followed by numeric ID
- Status: `.` (open), `x` (done), `~` (won't fix)
- Closed issues are archived in `docs/issues-archive.txt`

## Instructions

1. Read both `docs/issues.txt` and `docs/issues-archive.txt`
2. Find the highest issue ID across both files (parse `#N` from each line)
3. Add a new line to `docs/issues.txt` with format: `#[NEXT_ID] . [description]`
4. Report the new issue number

## Validation

- Issue description must not be empty
- If `$ARGUMENTS` is empty, ask the user for the issue description

## Output

```
Added issue #[ID]: [description]
```
