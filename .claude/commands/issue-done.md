---
description: Mark an issue as done/fixed
allowed-tools: Read, Edit
---

# Mark Issue Done

Mark an issue as "done/fixed" in `docs/issues.txt`.

## Arguments

`$ARGUMENTS` should contain the issue number (line number).

## Instructions

1. Parse the issue number from `$ARGUMENTS`
2. Read `docs/issues.txt`
3. Find the line with that number
4. Change the marker from `.` to `x` (done)
5. Report the change

## Validation

- Issue number must be provided
- Issue must exist
- Issue must currently be open (`.` marker)
- If issue is already done (`x`) or closed (`~`), report that

## Output

```
Marked issue #[NUMBER] as done: [description]
```

Or if already done/closed:
```
Issue #[NUMBER] is already [done/closed] ([marker])
```
