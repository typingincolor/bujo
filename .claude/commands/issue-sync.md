---
description: Sync issues.txt entries to GitHub issues
allowed-tools: Read, Edit, Bash(gh issue create *)
---

# Sync Issues

Create GitHub issues from entries staged in `docs/issues.txt`.

## File Format

Issues in `docs/issues.txt` use format: `#ID . DESCRIPTION`
- These are local staging IDs, not GitHub issue numbers
- After sync, entries are removed from issues.txt

## Instructions

1. Read `docs/issues.txt`
2. Parse all open issues (`.` marker)
3. For each issue:
   - Create GitHub issue: `gh issue create --title "[description]"`
   - Report the mapping: local #ID -> GitHub #NUMBER
4. Clear `docs/issues.txt` (keep header comments only)
5. Report summary

## Output Format

```
## Synced Issues

Local #ID -> GitHub #NUMBER: [title]
Local #ID -> GitHub #NUMBER: [title]
...

---
Created N GitHub issues
Cleared docs/issues.txt
```

If issues.txt is empty:
```
No staged issues to sync.
```
