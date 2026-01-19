---
description: Close an issue as won't fix
allowed-tools: Read, Edit
---

# Close Issue (Won't Fix)

Mark an issue as "won't fix" in `docs/issues.txt`.

## Arguments

`$ARGUMENTS` should contain the issue number (line number).

## Instructions

1. Parse the issue number from `$ARGUMENTS`
2. Read `docs/issues.txt`
3. Find the line with that number
4. Change the marker from `.` to `~` (won't fix)
5. Report the change

## Validation

- Issue number must be provided
- Issue must exist
- Issue must currently be open (`.` marker)
- If issue is already closed (`x` or `~`), report that it's already closed

## Output

```
Closed issue #[NUMBER] as won't fix: [description]
```

Or if already closed:
```
Issue #[NUMBER] is already closed ([marker])
```
