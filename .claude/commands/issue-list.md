---
description: List GitHub issues with optional filtering
allowed-tools: Bash(gh issue list *), Read
---

# List Issues

List GitHub issues with optional filtering.

## Arguments

The command accepts an optional filter argument: `$ARGUMENTS`

- No argument or `open`: Show only open issues **(default)**
- `all`: Show all issues including closed
- `closed`: Show only closed issues

## Instructions

### List Open Issues (default)

```bash
gh issue list --state open
```

### List All Issues

```bash
gh issue list --state all --limit 50
```

### List Closed Issues

```bash
gh issue list --state closed --limit 50
```

## Staging Area (Optional)

If `docs/issues.txt` has entries, mention them as "staged for sync":

```
## Staged in issues.txt (not yet in GitHub)
#ID . [description]
...

Run `/issue-sync` to create GitHub issues from staged entries.
```

## Output Format

```
## Open Issues (N)

#NUMBER  TITLE
...

---
Total: N open issues
```
