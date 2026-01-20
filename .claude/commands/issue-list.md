---
description: List issues from docs/issues.txt with optional filtering
allowed-tools: Read
---

# List Issues

Read and display issues from `docs/issues.txt` (and optionally `docs/issues-archive.txt`).

**Issue markers:**
- `.` = open (pending)
- `x` = done/fixed
- `~` = won't fix/abandoned

**File format:** `#ID STATUS DESCRIPTION`

## Arguments

The command accepts an optional filter argument: `$ARGUMENTS`

- No argument or `open`: Show only open issues (`.`) from `docs/issues.txt` **(default)**
- `all`: Show all issues from both files
- `done`: Show only completed issues (`x`) from archive
- `closed`: Show only won't-fix issues (`~`) from archive

## Instructions

1. Read `docs/issues.txt` (and `docs/issues-archive.txt` if filter requires)
2. Parse `#ID STATUS DESCRIPTION` format from each non-comment line
3. Display issues based on filter
4. Show summary counts

## Output Format

Use the actual markers from the file (`.`, `x`, `~`) in the output.

**Default (no argument or `open`):**
```
## Open Issues (N)

#ID . [description]
...
```

**For `all` filter:**
```
## Issues (filter: all)

### Open (N)
#ID . [description]
...

### Done (N)
#ID x [description]
...

### Won't Fix (N)
#ID ~ [description]
...

---
Total: N issues (N open, N done, N won't fix)
```
