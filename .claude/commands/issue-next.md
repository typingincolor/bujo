---
description: Recommend the next issue to work on
allowed-tools: Read, Glob, Grep
---

# Find Next Issue

Analyze open issues and recommend which one to tackle next.

## File Format

Issues use explicit IDs: `#ID STATUS DESCRIPTION`
- Open issues are in `docs/issues.txt`
- Parse `#N . description` format from each non-comment line

## Instructions

### Step 1: Get Open Issues

1. Read `docs/issues.txt`
2. Filter to only open issues (`.` marker)
3. Parse the issue ID and description from each line

### Step 2: Analyze Each Issue

For each open issue, consider:

1. **Complexity**: Is it a small fix or large feature?
   - Small: bug fix, UI tweak, text change
   - Medium: new feature in existing component, refactor
   - Large: new system, major architectural change

2. **Dependencies**: Does it depend on other issues?
   - Check if description references other issues
   - Check if it builds on recently completed work

3. **Impact**: How many users/features does it affect?
   - High: core functionality, blocking other work
   - Medium: important feature, nice to have
   - Low: edge case, cosmetic

4. **Context**: Is there related recent work?
   - Check git log for recently touched files
   - Issues in the same area as recent commits are easier

### Step 3: Rank and Recommend

Score issues and recommend the top 1-3, prioritizing:
1. Small, high-impact issues (quick wins)
2. Issues related to recent work (context is fresh)
3. Issues that unblock other work

## Output Format

```
## Recommended Next Issue

### #[ID]: [description]

**Why this one:**
- [Reason 1]
- [Reason 2]

**Estimated scope:** [small/medium/large]

---

### Other candidates:

**#[ID]:** [description]
- [Brief reason for/against]

**#[ID]:** [description]
- [Brief reason for/against]

---

To start working on the recommended issue:
`/issue-fix [ID]`
```
