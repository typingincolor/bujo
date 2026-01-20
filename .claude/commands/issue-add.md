---
description: Add a new issue to GitHub (or stage in issues.txt for later sync)
allowed-tools: Bash(gh issue create *), Read, Edit
---

# Add Issue

Create a new GitHub issue.

## Arguments

`$ARGUMENTS` contains the issue description.

## Instructions

1. If `$ARGUMENTS` is empty, ask the user for the issue description
2. Create the GitHub issue using `gh issue create`
3. Report the new issue number and URL

## Usage

```bash
gh issue create --title "[title]" --body "[description]"
```

For simple issues, the title can be the full description. For complex issues, extract a short title and put details in the body.

## Validation

- Issue description must not be empty
- If `$ARGUMENTS` is empty, ask the user for the issue description

## Output

```
Created issue #[NUMBER]: [title]
[URL]
```
