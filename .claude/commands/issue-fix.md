---
description: Start fixing a GitHub issue with TDD workflow
allowed-tools: Bash(gh issue view *), Bash(git status), Bash(git checkout *), Bash(git pull), Bash(git branch *), Read, Glob, Grep
---

# Fix Issue

Create a branch and start fixing a GitHub issue using strict TDD.

## Arguments

`$ARGUMENTS` should contain the GitHub issue number (e.g., `359` or `#359`).

## Instructions

### Step 1: Validate Issue

1. Parse the issue number from `$ARGUMENTS` (strip `#` if present)
2. Fetch issue details from GitHub: `gh issue view [NUMBER]`
3. Verify issue is open

### Step 2: Create Branch

1. Ensure working directory is clean (`git status`)
2. Checkout main and pull latest
3. Create branch: `fix/issue-[NUMBER]-[slug]`
   - slug = first 3-4 words of title, kebab-case, max 30 chars

### Step 3: Understand the Issue

1. Read the issue title and body carefully
2. Search the codebase for relevant files using Glob and Grep
3. Identify what needs to change

### Step 4: Plan the Fix

Present a plan to the user:
- What files need to change
- What tests need to be written
- Estimated scope (small/medium/large)

### Step 5: Execute with Strict TDD

Follow the project's TDD requirements from CLAUDE.md:

1. **RED:** Write failing test first
2. **GREEN:** Write minimum code to pass
3. **REFACTOR:** Clean up if needed

Ask clarifying questions if the requirements are ambiguous.

## Validation

- Issue number must be provided
- Issue must exist and be open on GitHub
- Git working directory must be clean

## Output

```
Starting fix for issue #[NUMBER]: [title]

Branch: fix/issue-[NUMBER]-[slug]

## Analysis
[Summary of what needs to change]

## Plan
1. [First step]
2. [Second step]
...

Proceed with TDD implementation? (The fix will follow RED-GREEN-REFACTOR)
```
