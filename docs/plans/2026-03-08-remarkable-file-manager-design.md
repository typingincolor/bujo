# reMarkable File Manager View

## Problem

The reMarkable tab renders documents as a flat list of buttons with no folder navigation, no file type distinction, and no scrolling support.

## Design

Rework `RemarkableView.tsx` to render a VS Code explorer-style file browser with folder navigation, distinct file type icons, breadcrumb bar, and scrolling.

### Layout (top to bottom)

1. **Breadcrumb bar** — `My files > Folder > Subfolder`, each segment clickable to jump back. Root is "My files".
2. **File list** — scrollable container filling available space:
   - Each row: icon (folder/notebook/pdf/epub) + name + last modified date (right-aligned)
   - Folders sort first, then files alphabetically
   - Folders clickable to navigate into them
   - Notebooks and PDFs clickable to trigger import (current behavior)
   - ePubs and other types shown muted/disabled (not importable)
3. **Empty state** — "This folder is empty" when navigating into an empty folder

### Behavior

- Track `currentFolderId` (root = `""`)
- Filter documents by `Parent === currentFolderId` to show current folder contents
- Maintain breadcrumb stack of `{id, name}` pairs for navigation
- File list container gets `overflow-y-auto` bounded to viewport
- Icons: Lucide — `Folder`, `NotebookPen`, `FileText` (PDF), `BookOpen` (ePub), `File` (fallback)

### Sorting

- Folders first, then files
- Alphabetical by name within each group

### Importable types

- Notebooks: yes (click triggers import + OCR)
- PDFs: yes (click triggers import + OCR)
- ePubs: no (shown muted)
- Other: no (shown muted)

### What stays the same

- Loading, not-registered, importing, and review steps unchanged
- `handleSelectDocument` still triggers import for notebooks/PDFs
- OCRReviewPanel untouched

### Scope

- Only `RemarkableView.tsx` changes
- No backend changes
- No new component files
- Tags deferred to follow-up
