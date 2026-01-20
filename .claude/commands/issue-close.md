---
description: Close an issue as won't fix
allowed-tools: Read, Edit
---

# Close Issue (Won't Fix)

Mark an issue as "won't fix" and move it to the archive.

## Arguments

`$ARGUMENTS` should contain the issue number (e.g., `90` or `#90`).

## File Format

Issues use explicit IDs: `#ID STATUS DESCRIPTION`
- Open issues are in `docs/issues.txt`
- Closed issues are archived in `docs/issues-archive.txt`

## Instructions

1. Parse the issue number from `$ARGUMENTS` (strip `#` if present)
2. Read `docs/issues.txt`
3. Find the line matching `#[NUMBER]`
4. Verify issue is open (`.` marker)
5. Remove the line from `docs/issues.txt`
6. Add the line to `docs/issues-archive.txt` with `~` marker: `#[NUMBER] ~ [description]`
7. Report the change

## Validation

- Issue number must be provided
- Issue must exist in `docs/issues.txt`
- Issue must currently be open (`.` marker)
- If issue is already in archive (`x` or `~`), report that it's already closed

## Output

```
Closed issue #[NUMBER] as won't fix: [description]
(Moved to issues-archive.txt)
```

Or if already closed:
```
Issue #[NUMBER] is already closed ([marker])
```
