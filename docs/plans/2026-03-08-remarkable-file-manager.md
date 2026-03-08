# reMarkable File Manager Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the flat document list in RemarkableView with a VS Code explorer-style file browser supporting folder navigation, file type icons, breadcrumbs, and scrolling.

**Architecture:** Frontend-only change to `RemarkableView.tsx`. The backend already returns `Parent` and `FileType` fields on each Document. We filter documents client-side by `Parent === currentFolderId` and maintain a breadcrumb stack for navigation. Folders have `FileType: ""` (empty string).

**Tech Stack:** React, TypeScript, TailwindCSS, Lucide React icons

---

### Task 1: Add folder navigation state and helpers

**Files:**
- Modify: `frontend/src/components/bujo/RemarkableView.tsx`

**Step 1: Add state and types for folder navigation**

Add these after the existing state declarations (line ~17):

```tsx
const [currentFolderId, setCurrentFolderId] = useState('')
const [breadcrumbs, setBreadcrumbs] = useState<{ id: string; name: string }[]>([])
```

**Step 2: Add navigation helper functions**

Add these after `handleSelectDocument` (line ~55):

```tsx
function isFolder(doc: remarkable.Document): boolean {
  return doc.FileType === ''
}

function isImportable(doc: remarkable.Document): boolean {
  return doc.FileType === 'notebook' || doc.FileType === 'pdf'
}

function handleNavigateToFolder(folderId: string, folderName: string) {
  setCurrentFolderId(folderId)
  setBreadcrumbs(prev => [...prev, { id: folderId, name: folderName }])
}

function handleBreadcrumbClick(index: number) {
  if (index === -1) {
    setCurrentFolderId('')
    setBreadcrumbs([])
  } else {
    const target = breadcrumbs[index]
    setCurrentFolderId(target.id)
    setBreadcrumbs(prev => prev.slice(0, index + 1))
  }
}

function currentDocuments(): remarkable.Document[] {
  const filtered = documents.filter(doc => doc.Parent === currentFolderId)
  const folders = filtered.filter(isFolder).sort((a, b) => a.VisibleName.localeCompare(b.VisibleName))
  const files = filtered.filter(d => !isFolder(d)).sort((a, b) => a.VisibleName.localeCompare(b.VisibleName))
  return [...folders, ...files]
}
```

**Step 3: Update handleSelectDocument to handle folders vs files**

Replace the existing `handleSelectDocument` function:

```tsx
async function handleSelectDocument(doc: remarkable.Document) {
  if (isFolder(doc)) {
    handleNavigateToFolder(doc.ID, doc.VisibleName)
    return
  }
  if (!isImportable(doc)) {
    return
  }
  setSelectedDocName(doc.VisibleName)
  setStep('importing')
  setError(null)

  try {
    const result = await ImportRemarkablePages(doc.ID)
    setImportResult(result)
    setStep('review')
  } catch (err) {
    setError(String(err))
    setStep('document-list')
  }
}
```

**Step 4: Verify the app still builds**

Run: `cd frontend && npm run build`
Expected: Build succeeds (unused state/functions are fine in dev)

**Step 5: Commit**

```
feat(remarkable): add folder navigation state and helpers
```

---

### Task 2: Replace flat list with file manager UI

**Files:**
- Modify: `frontend/src/components/bujo/RemarkableView.tsx`

**Step 1: Add Lucide icon imports**

Replace the existing imports at the top of the file:

```tsx
import { useState, useEffect, useCallback } from 'react'
import { Folder, NotebookPen, FileText, BookOpen, File, ChevronRight } from 'lucide-react'
import { ListRemarkableDocuments, IsRemarkableRegistered, ImportRemarkablePages } from '../../wailsjs/go/wails/App'
import { remarkable, wails } from '../../wailsjs/go/models'
import { OCRReviewPanel } from './OCRReviewPanel'
```

**Step 2: Add icon helper function**

Add after the navigation helpers:

```tsx
function fileIcon(doc: remarkable.Document) {
  if (isFolder(doc)) return <Folder className="w-4 h-4 text-blue-400" />
  switch (doc.FileType) {
    case 'notebook': return <NotebookPen className="w-4 h-4 text-amber-400" />
    case 'pdf': return <FileText className="w-4 h-4 text-red-400" />
    case 'epub': return <BookOpen className="w-4 h-4 text-green-400" />
    default: return <File className="w-4 h-4 text-muted-foreground" />
  }
}
```

**Step 3: Add date formatting helper**

```tsx
function formatDate(timestamp: string): string {
  const ms = parseInt(timestamp, 10)
  if (isNaN(ms)) return timestamp
  return new Date(ms).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}
```

**Step 4: Replace the document-list return block**

Replace the final return block (lines 120-137) with:

```tsx
return (
  <div className="flex flex-col h-full min-h-0">
    {/* Breadcrumb bar */}
    <div className="flex items-center gap-1 px-4 py-2 text-sm text-muted-foreground border-b border-border flex-shrink-0">
      <button
        onClick={() => handleBreadcrumbClick(-1)}
        className="hover:text-foreground transition-colors"
      >
        My files
      </button>
      {breadcrumbs.map((crumb, i) => (
        <span key={crumb.id} className="flex items-center gap-1">
          <ChevronRight className="w-3 h-3" />
          <button
            onClick={() => handleBreadcrumbClick(i)}
            className="hover:text-foreground transition-colors"
          >
            {crumb.name}
          </button>
        </span>
      ))}
    </div>

    {/* File list */}
    <div className="flex-1 min-h-0 overflow-y-auto">
      {currentDocuments().length === 0 ? (
        <p className="p-6 text-muted-foreground">This folder is empty.</p>
      ) : (
        <div className="divide-y divide-border">
          {currentDocuments().map(doc => {
            const clickable = isFolder(doc) || isImportable(doc)
            return (
              <button
                key={doc.ID}
                onClick={() => handleSelectDocument(doc)}
                disabled={!clickable}
                className={`w-full text-left px-4 py-2 flex items-center gap-3 transition-colors ${
                  clickable
                    ? 'hover:bg-accent cursor-pointer'
                    : 'opacity-50 cursor-default'
                }`}
              >
                {fileIcon(doc)}
                <span className={`flex-1 truncate text-sm ${clickable ? '' : 'text-muted-foreground'}`}>
                  {doc.VisibleName}
                </span>
                <span className="text-xs text-muted-foreground flex-shrink-0">
                  {formatDate(doc.LastModified)}
                </span>
              </button>
            )
          })}
        </div>
      )}
    </div>
  </div>
)
```

**Step 5: Verify build**

Run: `cd frontend && npm run build`
Expected: Build succeeds

**Step 6: Commit**

```
feat(remarkable): replace flat list with file manager UI
```

---

### Task 3: Wrap RemarkableView in proper height container in App.tsx

**Files:**
- Modify: `frontend/src/App.tsx`

**Step 1: Wrap RemarkableView in a full-height container**

Find (around line 841):
```tsx
{view === 'remarkable' && (
  <RemarkableView onNavigateToSettings={() => handleViewChange('settings')} />
)}
```

Replace with:
```tsx
{view === 'remarkable' && (
  <div className="h-full flex flex-col">
    <RemarkableView onNavigateToSettings={() => handleViewChange('settings')} />
  </div>
)}
```

**Step 2: Verify build**

Run: `cd frontend && npm run build`
Expected: Build succeeds

**Step 3: Commit**

```
feat(remarkable): wrap view in full-height container for scrolling
```

---

### Task 4: Manual verification and final commit

**Step 1: Run linting**

Run: `cd frontend && npm run lint`
Expected: No errors

**Step 2: Run frontend tests**

Run: `cd frontend && npm test -- --run`
Expected: All tests pass

**Step 3: Visual verification**

Launch the app and verify:
- Breadcrumb bar shows "My files" at root
- Folders display with blue folder icon, sorted first
- Notebooks show amber pen icon, PDFs show red file icon
- Clicking a folder navigates into it, breadcrumb updates
- Clicking breadcrumb segments navigates back correctly
- Non-importable files (epub, other) appear muted and are not clickable
- Long file lists scroll within the container
- Clicking a notebook/PDF still triggers import flow
- Empty folders show "This folder is empty."

**Step 4: Squash into single feature commit if desired**

```
feat(remarkable): file manager view with folder navigation and file type icons

Replace flat document list with VS Code explorer-style file browser.
Adds breadcrumb navigation, folder drill-down, file type icons
(notebook, PDF, ePub), sorting (folders first), and scrollable
file list.
```
