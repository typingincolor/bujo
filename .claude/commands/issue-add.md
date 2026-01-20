---
description: Add a new issue to GitHub (or stage in issues.txt for later sync)
allowed-tools: Bash(gh issue create *), Read, Edit
---

# Add Issue

Create a new GitHub issue with automatic label detection.

## Arguments

`$ARGUMENTS` contains the issue description.

## Instructions

1. If `$ARGUMENTS` is empty, ask the user for the issue description
2. Analyze the description to determine appropriate labels (see Label Detection below)
3. Create the GitHub issue using `gh issue create` with detected labels
4. Report the new issue number, URL, and applied labels

## Label Detection

Analyze the issue description and apply labels based on these rules:

### Type Labels (pick one)
- `bug` - Words like: bug, broken, error, crash, fail, fix, wrong, doesn't work
- `enhancement` - Words like: add, feature, improve, support, new, implement, allow
- `documentation` - Words like: docs, readme, document, guide, explain
- `question` - Words like: how, why, question, help, unclear
- Default to `enhancement` if unclear

### Area Labels (pick all that apply)
- `frontend` - Words like: UI, component, React, TypeScript, page, button, modal, display, view
- `wails` - Words like: desktop, app, window, Wails, native, electron
- `AI` - Words like: AI, Gemini, summary, generate, LLM, prompt
- `adapter` - Words like: CLI, command, Cobra, adapter
- `go` - Default for backend/domain features not matching other areas

### Size Labels (optional, based on scope)
- `size/xs` - Trivial change, single line
- `size/s` - Small change, single file
- `size/m` - Medium change, few files
- `size/l` - Large feature, multiple components
- `size/xl` - Major feature, significant architecture

## Usage

```bash
gh issue create --title "[title]" --body "[description]" --label "enhancement" --label "frontend"
```

For simple issues, the title can be the full description. For complex issues, extract a short title and put details in the body.

## Validation

- Issue description must not be empty
- If `$ARGUMENTS` is empty, ask the user for the issue description
- At minimum, apply a type label (bug/enhancement/documentation/question)

## Output

```
Created issue #[NUMBER]: [title]
Labels: [label1], [label2], ...
[URL]
```
