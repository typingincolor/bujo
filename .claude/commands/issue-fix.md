---
description: Start fixing an issue with TDD workflow
allowed-tools: Read, Edit, Bash(git status), Bash(git checkout *), Bash(git pull), Bash(git branch *), Glob, Grep
---

# Fix Issue

Create a branch and start fixing an issue using strict TDD.

## Arguments

`$ARGUMENTS` should contain the issue number (e.g., `90` or `#90`).

## File Format

Issues use explicit IDs: `#ID STATUS DESCRIPTION`
- Open issues are in `docs/issues.txt`
- Closed issues are archived in `docs/issues-archive.txt`

## Instructions

### Step 1: Validate Issue

1. Parse the issue number from `$ARGUMENTS` (strip `#` if present)
2. Read `docs/issues.txt`
3. Find the line matching `#[NUMBER]`
4. Verify issue is open (`.` marker)

### Step 2: Create Branch

1. Ensure working directory is clean (`git status`)
2. Checkout main and pull latest
3. Create branch: `fix/issue-[NUMBER]-[slug]`
   - slug = first 3-4 words of description, kebab-case, max 30 chars

### Step 3: Understand the Issue

1. Read the issue description carefully
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
- Issue must exist and be open (`.` marker)
- Git working directory must be clean

## Output

```
Starting fix for issue #[NUMBER]: [description]

Branch: fix/issue-[NUMBER]-[slug]

## Analysis
[Summary of what needs to change]

## Plan
1. [First step]
2. [Second step]
...

Proceed with TDD implementation? (The fix will follow RED-GREEN-REFACTOR)
```
