---
description: List issues from docs/issues.txt with optional filtering
allowed-tools: Read
---

# List Issues

Read and display issues from `docs/issues.txt`.

**Issue markers:**
- `.` = open (pending)
- `x` = done/fixed
- `~` = won't fix/abandoned

## Arguments

The command accepts an optional filter argument: `$ARGUMENTS`

- No argument or `open`: Show only open issues (`.`) **(default)**
- `all`: Show all issues
- `done`: Show only completed issues (`x`)
- `closed`: Show only won't-fix issues (`~`)

## Instructions

1. Read `docs/issues.txt`
2. Parse and display issues based on filter
3. Show summary counts

## Output Format

Use the actual markers from the file (`.`, `x`, `~`) in the output.

**Default (no argument or `open`):**
```
## Open Issues (N)

. [line#] [description]
...
```

**For `all` filter:**
```
## Issues (filter: all)

### Open (N)
. [line#] [description]
...

### Done (N)
x [line#] [description]
...

### Won't Fix (N)
~ [line#] [description]
...

---
Total: N issues (N open, N done, N won't fix)
```
