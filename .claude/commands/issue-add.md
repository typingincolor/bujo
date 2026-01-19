---
description: Add a new issue to docs/issues.txt
allowed-tools: Read, Edit
---

# Add Issue

Add a new issue to `docs/issues.txt`.

## Arguments

`$ARGUMENTS` contains the issue description.

## Instructions

1. Read `docs/issues.txt` to find the last line
2. Add a new line with format: `. [description]`
3. The `.` marker indicates an open issue
4. Report the new issue number (line number)

## Validation

- Issue description must not be empty
- If `$ARGUMENTS` is empty, ask the user for the issue description

## Output

```
Added issue #[LINE_NUMBER]: [description]
```
