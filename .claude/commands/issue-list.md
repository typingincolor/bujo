---
description: List issues from docs/issues.txt with optional filtering
allowed-tools: Read, Bash(grep:*), Bash(wc:*)
---

# List Issues

Read and display issues from `docs/issues.txt`.

**Issue markers:**
- `.` = open (pending)
- `x` = done/fixed
- `~` = won't fix/abandoned

## Arguments

The command accepts an optional filter argument: `$ARGUMENTS`

- No argument or `all`: Show all issues
- `open`: Show only open issues (`.`)
- `done`: Show only completed issues (`x`)
- `closed`: Show only won't-fix issues (`~`)

## Instructions

1. Read `docs/issues.txt`
2. Parse and display issues based on filter
3. Show summary counts

## Output Format

```
## Issues (filter: [FILTER])

### Open (N)
[line#]. [description]
...

### Done (N)
[line#]. [description]
...

### Won't Fix (N)
[line#]. [description]
...

---
Total: N issues (N open, N done, N won't fix)
```

If filtered, only show the relevant section.

For `open` filter, just show:
```
## Open Issues (N)

[line#]. [description]
...
```
