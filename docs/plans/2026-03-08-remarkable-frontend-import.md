# reMarkable Frontend Import Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Integrate the reMarkable handwriting import pipeline into the Wails desktop app with a three-step UI flow: document list, page preview with OCR, and side-by-side review with confidence highlighting.

**Architecture:** Four new Wails backend methods expose the existing reMarkable adapter (client, parser, renderer, OCR) to the frontend. A new `RemarkableView` React component manages the three-step workflow. OCR binary discovery uses `os.Executable()` to find `remarkable-ocr` next to the main binary. Platform check hides the feature on non-macOS.

**Tech Stack:** Go (Wails v2 backend), React 19, TypeScript, Tailwind CSS, CodeMirror 6, lucide-react icons

**Design Doc:** `docs/plans/2026-03-08-remarkable-frontend-import-design.md`

---

### Task 1: GetPlatformCapabilities Backend Method

**Files:**
- Modify: `internal/adapter/wails/app.go`

**Step 1: Write the failing test**

Create `internal/adapter/wails/platform_test.go`:

```go
package wails

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindOCRBinary_NextToExecutable(t *testing.T) {
	dir := t.TempDir()
	ocrPath := filepath.Join(dir, "remarkable-ocr")
	err := os.WriteFile(ocrPath, []byte("fake"), 0755)
	assert.NoError(t, err)

	result := findOCRBinary(dir)
	assert.Equal(t, ocrPath, result)
}

func TestFindOCRBinary_NotFound(t *testing.T) {
	dir := t.TempDir()
	result := findOCRBinary(dir)
	assert.Equal(t, "", result)
}

func TestPlatformCapabilities_Platform(t *testing.T) {
	caps := buildPlatformCapabilities("")
	assert.Equal(t, runtime.GOOS, caps.Platform)
}

func TestPlatformCapabilities_HasOCR_WhenBinaryExists(t *testing.T) {
	caps := buildPlatformCapabilities("/some/path/remarkable-ocr")
	if runtime.GOOS == "darwin" {
		assert.True(t, caps.HasOCR)
	} else {
		assert.False(t, caps.HasOCR)
	}
}

func TestPlatformCapabilities_HasOCR_WhenBinaryMissing(t *testing.T) {
	caps := buildPlatformCapabilities("")
	assert.False(t, caps.HasOCR)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/wails/ -run TestFindOCRBinary -v`
Expected: FAIL — `findOCRBinary` and `buildPlatformCapabilities` undefined

**Step 3: Write minimal implementation**

Add to `internal/adapter/wails/app.go`:

```go
type PlatformCapabilities struct {
	HasOCR   bool   `json:"hasOCR"`
	Platform string `json:"platform"`
}

func findOCRBinary(execDir string) string {
	ocrPath := filepath.Join(execDir, "remarkable-ocr")
	if _, err := os.Stat(ocrPath); err == nil {
		return ocrPath
	}
	return ""
}

func buildPlatformCapabilities(ocrPath string) PlatformCapabilities {
	return PlatformCapabilities{
		HasOCR:   runtime.GOOS == "darwin" && ocrPath != "",
		Platform: runtime.GOOS,
	}
}

func (a *App) GetPlatformCapabilities() PlatformCapabilities {
	execPath, err := os.Executable()
	if err != nil {
		return buildPlatformCapabilities("")
	}
	ocrPath := findOCRBinary(filepath.Dir(execPath))
	return buildPlatformCapabilities(ocrPath)
}
```

Add `"path/filepath"` and `"runtime"` to imports (if not already present).

**Step 4: Run test to verify it passes**

Run: `go test ./internal/adapter/wails/ -run TestFindOCRBinary -v && go test ./internal/adapter/wails/ -run TestPlatformCapabilities -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adapter/wails/app.go internal/adapter/wails/platform_test.go
git commit -m "feat(wails): add GetPlatformCapabilities for reMarkable OCR detection"
```

---

### Task 2: ListRemarkableDocuments Backend Method

**Files:**
- Modify: `internal/adapter/wails/app.go`

**Step 1: Write the failing test**

Add to `internal/adapter/wails/platform_test.go`:

```go
func TestListRemarkableDocuments_NoConfig(t *testing.T) {
	app := &App{}
	_, err := app.ListRemarkableDocuments()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/wails/ -run TestListRemarkableDocuments -v`
Expected: FAIL — `ListRemarkableDocuments` undefined

**Step 3: Write minimal implementation**

Add to `internal/adapter/wails/app.go`:

```go
func (a *App) ListRemarkableDocuments() ([]remarkable.Document, error) {
	configPath, err := remarkable.DefaultConfigPath()
	if err != nil {
		return nil, fmt.Errorf("not registered: %w", err)
	}
	cfg, err := remarkable.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("not registered — configure in Settings: %w", err)
	}

	client := remarkable.NewClient(remarkable.DefaultAuthHost)
	client.SetSyncHost(remarkable.DefaultSyncHost)

	docs, err := client.ListDocuments(cfg.DeviceToken)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	return docs, nil
}
```

Add `"github.com/typingincolor/bujo/internal/adapter/remarkable"` to imports.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/adapter/wails/ -run TestListRemarkableDocuments -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adapter/wails/app.go internal/adapter/wails/platform_test.go
git commit -m "feat(wails): add ListRemarkableDocuments method"
```

---

### Task 3: ImportRemarkablePages Backend Method

**Files:**
- Modify: `internal/adapter/wails/app.go`

**Step 1: Write the failing test**

Add to `internal/adapter/wails/platform_test.go`:

```go
func TestImportResult_JSONSerialization(t *testing.T) {
	result := ImportRemarkableResult{
		Pages: []ImportedPage{
			{
				PageID: "page-1",
				PNG:    "base64data",
				OCRResults: []remarkable.OCRResult{
					{Text: "hello", X: 10, Y: 20, Width: 100, Height: 30, Confidence: 0.95},
				},
			},
		},
	}
	assert.Equal(t, 1, len(result.Pages))
	assert.Equal(t, "page-1", result.Pages[0].PageID)
	assert.Equal(t, float32(0.95), result.Pages[0].OCRResults[0].Confidence)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/wails/ -run TestImportResult -v`
Expected: FAIL — `ImportRemarkableResult` and `ImportedPage` undefined

**Step 3: Write minimal implementation**

Add to `internal/adapter/wails/app.go`:

```go
type ImportedPage struct {
	PageID     string                `json:"pageID"`
	PNG        string                `json:"png"`
	OCRResults []remarkable.OCRResult `json:"ocrResults"`
	Error      string                `json:"error,omitempty"`
}

type ImportRemarkableResult struct {
	Pages []ImportedPage `json:"pages"`
}

func (a *App) ImportRemarkablePages(docID string) (*ImportRemarkableResult, error) {
	configPath, err := remarkable.DefaultConfigPath()
	if err != nil {
		return nil, fmt.Errorf("not registered: %w", err)
	}
	cfg, err := remarkable.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("not registered — configure in Settings: %w", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("cannot determine executable path: %w", err)
	}
	ocrPath := findOCRBinary(filepath.Dir(execPath))
	if ocrPath == "" {
		return nil, fmt.Errorf("OCR binary not found")
	}

	client := remarkable.NewClient(remarkable.DefaultAuthHost)
	client.SetSyncHost(remarkable.DefaultSyncHost)

	pages, err := client.DownloadPages(cfg.DeviceToken, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to download pages: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "remarkable-import-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	var result ImportRemarkableResult
	for _, page := range pages {
		imported := ImportedPage{PageID: page.PageID}

		pngPath, err := remarkable.RenderPageToPNG(tmpDir, page.PageID, page.Data)
		if err != nil {
			imported.Error = fmt.Sprintf("render failed: %v", err)
			result.Pages = append(result.Pages, imported)
			continue
		}

		pngData, err := os.ReadFile(pngPath)
		if err != nil {
			imported.Error = fmt.Sprintf("read PNG failed: %v", err)
			result.Pages = append(result.Pages, imported)
			continue
		}
		imported.PNG = base64.StdEncoding.EncodeToString(pngData)

		ocrResults, err := remarkable.RunOCR(ocrPath, pngPath)
		if err != nil {
			imported.Error = fmt.Sprintf("OCR failed: %v", err)
			result.Pages = append(result.Pages, imported)
			continue
		}
		imported.OCRResults = ocrResults
		result.Pages = append(result.Pages, imported)
	}

	return &result, nil
}
```

Add `"encoding/base64"` to imports.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/adapter/wails/ -run TestImportResult -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adapter/wails/app.go internal/adapter/wails/platform_test.go
git commit -m "feat(wails): add ImportRemarkablePages pipeline method"
```

---

### Task 4: ImportEntries Backend Method

**Files:**
- Modify: `internal/adapter/wails/app.go`

**Step 1: Write the failing test**

Add to `internal/adapter/wails/platform_test.go`:

```go
func TestImportEntries_EmptyText(t *testing.T) {
	app := &App{}
	err := app.ImportEntries("", "2026-03-08")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestImportEntries_InvalidDate(t *testing.T) {
	app := &App{}
	err := app.ImportEntries(". buy milk", "not-a-date")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "date")
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/wails/ -run TestImportEntries -v`
Expected: FAIL — `ImportEntries` undefined

**Step 3: Write minimal implementation**

Add to `internal/adapter/wails/app.go`:

```go
func (a *App) ImportEntries(text string, date string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("empty text — nothing to import")
	}

	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
	}

	_, err = a.services.Bujo.LogEntries(a.ctx, text, service.LogEntriesOptions{Date: parsedDate})
	if err != nil {
		return fmt.Errorf("failed to import entries: %w", err)
	}

	runtime.EventsEmit(a.ctx, eventDataChanged)
	return nil
}
```

Add `"strings"` to imports (if not already present).

**Step 4: Run test to verify it passes**

Run: `go test ./internal/adapter/wails/ -run TestImportEntries -v`
Expected: PASS (empty text and bad date errors hit before services are needed)

**Step 5: Commit**

```bash
git add internal/adapter/wails/app.go internal/adapter/wails/platform_test.go
git commit -m "feat(wails): add ImportEntries method for reMarkable text import"
```

---

### Task 5: RegisterRemarkableDevice Backend Method

**Files:**
- Modify: `internal/adapter/wails/app.go`

The Settings page needs a way to trigger device registration.

**Step 1: Write the failing test**

Add to `internal/adapter/wails/platform_test.go`:

```go
func TestRegisterRemarkableDevice_EmptyCode(t *testing.T) {
	app := &App{}
	err := app.RegisterRemarkableDevice("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code")
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/wails/ -run TestRegisterRemarkableDevice -v`
Expected: FAIL — `RegisterRemarkableDevice` undefined

**Step 3: Write minimal implementation**

Add to `internal/adapter/wails/app.go`:

```go
func (a *App) RegisterRemarkableDevice(code string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("registration code is required")
	}

	client := remarkable.NewClient(remarkable.DefaultAuthHost)

	deviceToken, err := client.RegisterDevice(code)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	configPath, err := remarkable.DefaultConfigPath()
	if err != nil {
		return fmt.Errorf("failed to determine config path: %w", err)
	}

	cfg := remarkable.Config{DeviceToken: deviceToken}
	if err := remarkable.SaveConfig(configPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func (a *App) IsRemarkableRegistered() bool {
	configPath, err := remarkable.DefaultConfigPath()
	if err != nil {
		return false
	}
	_, err = remarkable.LoadConfig(configPath)
	return err == nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/adapter/wails/ -run TestRegisterRemarkableDevice -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adapter/wails/app.go internal/adapter/wails/platform_test.go
git commit -m "feat(wails): add RegisterRemarkableDevice and IsRemarkableRegistered"
```

---

### Task 6: Generate Wails TypeScript Bindings

**Files:**
- Generated: `frontend/wailsjs/go/wails/App.d.ts` (auto-generated)

**Step 1: Regenerate Wails bindings**

Run: `cd /Users/andrew/Development/bujo && wails generate module`

This generates TypeScript types for the new Go methods and structs (`PlatformCapabilities`, `ImportRemarkableResult`, `ImportedPage`, `Document`).

**Step 2: Verify bindings exist**

Check that `frontend/wailsjs/go/wails/App.d.ts` contains:
- `GetPlatformCapabilities(): Promise<PlatformCapabilities>`
- `ListRemarkableDocuments(): Promise<Array<remarkable.Document>>`
- `ImportRemarkablePages(docID: string): Promise<ImportRemarkableResult>`
- `ImportEntries(text: string, date: string): Promise<void>`
- `RegisterRemarkableDevice(code: string): Promise<void>`
- `IsRemarkableRegistered(): Promise<boolean>`

**Step 3: Commit**

```bash
git add frontend/wailsjs/
git commit -m "chore: regenerate Wails TypeScript bindings for reMarkable methods"
```

---

### Task 7: Add 'remarkable' ViewType to Sidebar

**Files:**
- Modify: `frontend/src/components/bujo/Sidebar.tsx`
- Modify: `frontend/src/App.tsx`

**Step 1: Update ViewType union**

In `frontend/src/components/bujo/Sidebar.tsx` line 17, add `'remarkable'` to the union:

```typescript
export type ViewType = 'today' | 'pending' | 'week' | 'questions' | 'habits' | 'lists' | 'goals' | 'search' | 'stats' | 'insights' | 'settings' | 'editable' | 'remarkable';
```

**Step 2: Add conditional sidebar entry**

The sidebar entry should only appear when `hasOCR` is true. Update `Sidebar.tsx`:

```typescript
// Add Tablet icon to imports
import { ..., Tablet } from 'lucide-react';

// Update props
interface SidebarProps {
  currentView: ViewType;
  onViewChange: (view: ViewType) => void;
  hasRemarkable?: boolean;
}

// Add remarkable to navItems conditionally
export function Sidebar({ currentView, onViewChange, hasRemarkable }: SidebarProps) {
```

Add the reMarkable button between Insights and the Settings footer, conditionally rendered:

```tsx
{hasRemarkable && (
  <button
    onClick={() => onViewChange('remarkable')}
    aria-pressed={currentView === 'remarkable'}
    className={cn(
      'w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-all',
      currentView === 'remarkable'
        ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium'
        : 'text-sidebar-foreground hover:bg-sidebar-accent/50'
    )}
  >
    <Tablet className="w-4 h-4" />
    reMarkable
  </button>
)}
```

Place this after the `navItems.map()` block and before the footer `<div className="p-3">`.

**Step 3: Update App.tsx**

In `App.tsx`, add to `validViews` array:
```typescript
const validViews: ViewType[] = ['today', 'pending', 'week', 'questions', 'habits', 'lists', 'goals', 'search', 'stats', 'insights', 'settings', 'remarkable']
```

Add to `viewTitles`:
```typescript
remarkable: 'reMarkable Import',
```

Add platform capabilities state and fetch on mount:
```typescript
const [hasRemarkable, setHasRemarkable] = useState(false)

useEffect(() => {
  GetPlatformCapabilities().then(caps => {
    setHasRemarkable(caps.hasOCR)
  })
}, [])
```

Pass to Sidebar:
```tsx
<Sidebar currentView={view} onViewChange={handleViewChange} hasRemarkable={hasRemarkable} />
```

Add placeholder view render:
```tsx
{view === 'remarkable' && (
  <div className="p-6 text-muted-foreground">reMarkable import coming soon</div>
)}
```

Import `GetPlatformCapabilities` from Wails bindings.

**Step 4: Verify**

Run: `cd frontend && npm run build`
Expected: Build succeeds with no TypeScript errors

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/Sidebar.tsx frontend/src/App.tsx
git commit -m "feat(frontend): add reMarkable sidebar entry with platform detection"
```

---

### Task 8: RemarkableView Container Component

**Files:**
- Create: `frontend/src/components/bujo/RemarkableView.tsx`
- Modify: `frontend/src/App.tsx` (replace placeholder)

**Step 1: Create the component with step state machine**

Create `frontend/src/components/bujo/RemarkableView.tsx`:

```tsx
import { useState, useEffect } from 'react'
import { ListRemarkableDocuments, IsRemarkableRegistered } from '../../wailsjs/go/wails/App'
import { remarkable } from '../../wailsjs/go/models'

type Step = 'loading' | 'not-registered' | 'document-list' | 'importing' | 'review'

interface RemarkableViewProps {
  onNavigateToSettings: () => void
}

export function RemarkableView({ onNavigateToSettings }: RemarkableViewProps) {
  const [step, setStep] = useState<Step>('loading')
  const [documents, setDocuments] = useState<remarkable.Document[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    IsRemarkableRegistered().then(registered => {
      if (!registered) {
        setStep('not-registered')
        return
      }
      loadDocuments()
    })
  }, [])

  async function loadDocuments() {
    try {
      setError(null)
      const docs = await ListRemarkableDocuments()
      setDocuments(docs)
      setStep('document-list')
    } catch (err) {
      setError(String(err))
      setStep('document-list')
    }
  }

  if (step === 'loading') {
    return <div className="p-6 text-muted-foreground">Loading...</div>
  }

  if (step === 'not-registered') {
    return (
      <div className="p-6 space-y-4">
        <p className="text-muted-foreground">
          reMarkable tablet not connected. Register your device in Settings to get started.
        </p>
        <button
          onClick={onNavigateToSettings}
          className="px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm"
        >
          Open Settings
        </button>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6 space-y-4">
        <p className="text-destructive">{error}</p>
        <button
          onClick={loadDocuments}
          className="px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm"
        >
          Retry
        </button>
      </div>
    )
  }

  if (step === 'document-list') {
    return (
      <div className="p-6 space-y-2">
        {documents.length === 0 ? (
          <p className="text-muted-foreground">No notebooks found on your reMarkable.</p>
        ) : (
          documents.map(doc => (
            <button
              key={doc.ID}
              onClick={() => handleSelectDocument(doc.ID)}
              className="w-full text-left px-4 py-3 rounded-lg border border-border hover:bg-accent transition-colors"
            >
              <div className="font-medium">{doc.VisibleName}</div>
              <div className="text-xs text-muted-foreground">{doc.FileType} · {doc.LastModified}</div>
            </button>
          ))
        )}
      </div>
    )
  }

  if (step === 'importing') {
    return <div className="p-6 text-muted-foreground">Importing and processing pages...</div>
  }

  return <div className="p-6 text-muted-foreground">Review step (Task 10)</div>
}
```

Note: `handleSelectDocument` and the review step will be implemented in Tasks 9 and 10.

**Step 2: Wire into App.tsx**

Replace the placeholder `{view === 'remarkable' && ...}` block with:

```tsx
{view === 'remarkable' && (
  <RemarkableView onNavigateToSettings={() => handleViewChange('settings')} />
)}
```

Import `RemarkableView` at the top of App.tsx.

**Step 3: Verify**

Run: `cd frontend && npm run build`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add frontend/src/components/bujo/RemarkableView.tsx frontend/src/App.tsx
git commit -m "feat(frontend): add RemarkableView container with document list"
```

---

### Task 9: Import Pipeline Integration (Document Selection → OCR)

**Files:**
- Modify: `frontend/src/components/bujo/RemarkableView.tsx`

**Step 1: Add import pipeline state and handler**

Add to `RemarkableView.tsx`:

```tsx
import { ImportRemarkablePages } from '../../wailsjs/go/wails/App'
import { wails } from '../../wailsjs/go/models'

// Add state
const [importResult, setImportResult] = useState<wails.ImportRemarkableResult | null>(null)
const [selectedDocName, setSelectedDocName] = useState('')

async function handleSelectDocument(docID: string) {
  const doc = documents.find(d => d.ID === docID)
  setSelectedDocName(doc?.VisibleName ?? docID)
  setStep('importing')
  setError(null)

  try {
    const result = await ImportRemarkablePages(docID)
    setImportResult(result)
    setStep('review')
  } catch (err) {
    setError(String(err))
    setStep('document-list')
  }
}
```

Add a back button to the importing step:

```tsx
if (step === 'importing') {
  return (
    <div className="p-6 space-y-2">
      <p className="text-muted-foreground">
        Downloading and processing pages from "{selectedDocName}"...
      </p>
      <p className="text-xs text-muted-foreground">This may take a moment for notebooks with many pages.</p>
    </div>
  )
}
```

**Step 2: Verify**

Run: `cd frontend && npm run build`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add frontend/src/components/bujo/RemarkableView.tsx
git commit -m "feat(frontend): add import pipeline integration to RemarkableView"
```

---

### Task 10: OCR Review Panel (Side-by-Side View)

**Files:**
- Create: `frontend/src/components/bujo/OCRReviewPanel.tsx`
- Modify: `frontend/src/components/bujo/RemarkableView.tsx`

**Step 1: Create OCRReviewPanel component**

Create `frontend/src/components/bujo/OCRReviewPanel.tsx`:

```tsx
import { useState, useMemo } from 'react'
import { ImportEntries } from '../../wailsjs/go/wails/App'
import { remarkable, wails } from '../../wailsjs/go/models'

interface OCRReviewPanelProps {
  pages: wails.ImportedPage[]
  documentName: string
  onDone: () => void
  onBack: () => void
}

function reconstructText(results: remarkable.OCRResult[]): string {
  if (!results || results.length === 0) return ''

  const sorted = [...results].sort((a, b) => a.Y - b.Y)
  const minX = Math.min(...sorted.map(r => r.X))
  const indentWidth = 50

  return sorted.map(r => {
    const depth = Math.round((r.X - minX) / indentWidth)
    const indent = '  '.repeat(depth)
    return indent + r.Text
  }).join('\n')
}

export function OCRReviewPanel({ pages, documentName, onDone, onBack }: OCRReviewPanelProps) {
  const [currentPage, setCurrentPage] = useState(0)
  const [date, setDate] = useState(() => new Date().toISOString().split('T')[0])
  const [importing, setImporting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const pageTexts = useMemo(() => {
    return pages.map(p => reconstructText(p.OCRResults))
  }, [pages])

  const [editedTexts, setEditedTexts] = useState<string[]>(() => [...pageTexts])

  const page = pages[currentPage]
  const hasError = page?.Error

  function updateText(index: number, text: string) {
    setEditedTexts(prev => {
      const next = [...prev]
      next[index] = text
      return next
    })
  }

  async function handleImport() {
    setImporting(true)
    setError(null)

    const combined = editedTexts.filter(t => t.trim()).join('\n')
    if (!combined.trim()) {
      setError('No text to import')
      setImporting(false)
      return
    }

    try {
      await ImportEntries(combined, date)
      onDone()
    } catch (err) {
      setError(String(err))
      setImporting(false)
    }
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-border">
        <div className="flex items-center gap-4">
          <button onClick={onBack} className="text-sm text-muted-foreground hover:text-foreground">
            ← Back
          </button>
          <span className="font-medium">{documentName}</span>
          <span className="text-sm text-muted-foreground">
            Page {currentPage + 1} of {pages.length}
          </span>
        </div>
        <div className="flex items-center gap-3">
          <label className="text-sm text-muted-foreground">
            Date:
            <input
              type="date"
              value={date}
              onChange={e => setDate(e.target.value)}
              className="ml-2 px-2 py-1 bg-background border border-border rounded text-sm"
            />
          </label>
          <button
            onClick={handleImport}
            disabled={importing}
            className="px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm disabled:opacity-50"
          >
            {importing ? 'Importing...' : 'Import to Journal'}
          </button>
        </div>
      </div>

      {error && (
        <div className="px-4 py-2 bg-destructive/10 text-destructive text-sm">{error}</div>
      )}

      {/* Page navigation */}
      {pages.length > 1 && (
        <div className="flex gap-1 p-2 border-b border-border overflow-x-auto">
          {pages.map((_, i) => (
            <button
              key={i}
              onClick={() => setCurrentPage(i)}
              className={`px-3 py-1 rounded text-sm ${
                i === currentPage
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:bg-accent'
              }`}
            >
              {i + 1}
            </button>
          ))}
        </div>
      )}

      {/* Side-by-side content */}
      <div className="flex-1 flex min-h-0">
        {/* Left: PNG preview */}
        <div className="w-1/2 overflow-auto border-r border-border p-4">
          {hasError ? (
            <div className="text-destructive text-sm">{page.Error}</div>
          ) : page?.PNG ? (
            <img
              src={`data:image/png;base64,${page.PNG}`}
              alt={`Page ${currentPage + 1}`}
              className="max-w-full"
            />
          ) : (
            <div className="text-muted-foreground text-sm">No image available</div>
          )}
        </div>

        {/* Right: Text editor */}
        <div className="w-1/2 overflow-auto p-4">
          <textarea
            value={editedTexts[currentPage] ?? ''}
            onChange={e => updateText(currentPage, e.target.value)}
            className="w-full h-full min-h-[400px] p-3 bg-background border border-border rounded font-mono text-sm resize-none focus:outline-none focus:ring-1 focus:ring-primary"
            placeholder="OCR text will appear here..."
          />
        </div>
      </div>
    </div>
  )
}
```

**Step 2: Wire into RemarkableView**

In `RemarkableView.tsx`, replace the review step placeholder:

```tsx
import { OCRReviewPanel } from './OCRReviewPanel'

// In the render, replace the review return:
if (step === 'review' && importResult) {
  return (
    <OCRReviewPanel
      pages={importResult.Pages}
      documentName={selectedDocName}
      onDone={() => {
        setStep('document-list')
        setImportResult(null)
        loadDocuments()
      }}
      onBack={() => {
        setStep('document-list')
        setImportResult(null)
      }}
    />
  )
}
```

**Step 3: Verify**

Run: `cd frontend && npm run build`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add frontend/src/components/bujo/OCRReviewPanel.tsx frontend/src/components/bujo/RemarkableView.tsx
git commit -m "feat(frontend): add OCR review panel with side-by-side PNG and editor"
```

---

### Task 11: Settings Page — reMarkable Registration

**Files:**
- Modify: `frontend/src/components/bujo/SettingsView.tsx` (or wherever settings lives)

**Step 1: Find Settings component**

Look for the settings view component. Check: `frontend/src/App.tsx` for how `settings` view is rendered, then modify that component.

**Step 2: Add registration section**

Add a reMarkable section to the settings view:

```tsx
import { useState } from 'react'
import { RegisterRemarkableDevice, IsRemarkableRegistered } from '../../wailsjs/go/wails/App'

// Inside the settings component, add:
function RemarkableSettings() {
  const [code, setCode] = useState('')
  const [status, setStatus] = useState<'idle' | 'registering' | 'success' | 'error'>('idle')
  const [error, setError] = useState('')
  const [isRegistered, setIsRegistered] = useState<boolean | null>(null)

  useEffect(() => {
    IsRemarkableRegistered().then(setIsRegistered)
  }, [])

  async function handleRegister() {
    setStatus('registering')
    setError('')
    try {
      await RegisterRemarkableDevice(code)
      setStatus('success')
      setIsRegistered(true)
      setCode('')
    } catch (err) {
      setError(String(err))
      setStatus('error')
    }
  }

  return (
    <div className="space-y-3">
      <h3 className="text-sm font-medium">reMarkable Tablet</h3>
      {isRegistered ? (
        <p className="text-sm text-muted-foreground">Device registered ✓</p>
      ) : (
        <div className="space-y-2">
          <p className="text-xs text-muted-foreground">
            Get a code from my.remarkable.com/device/browser/connect
          </p>
          <div className="flex gap-2">
            <input
              type="text"
              value={code}
              onChange={e => setCode(e.target.value)}
              placeholder="Enter 8-character code"
              maxLength={8}
              className="flex-1 px-3 py-2 bg-background border border-border rounded text-sm"
            />
            <button
              onClick={handleRegister}
              disabled={code.length < 8 || status === 'registering'}
              className="px-4 py-2 bg-primary text-primary-foreground rounded text-sm disabled:opacity-50"
            >
              {status === 'registering' ? 'Registering...' : 'Register'}
            </button>
          </div>
          {status === 'success' && (
            <p className="text-sm text-green-600">Device registered successfully!</p>
          )}
          {status === 'error' && (
            <p className="text-sm text-destructive">{error}</p>
          )}
        </div>
      )}
    </div>
  )
}
```

Add `<RemarkableSettings />` to the settings view's content area.

**Step 3: Verify**

Run: `cd frontend && npm run build`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add frontend/src/components/bujo/SettingsView.tsx  # or whatever the file is
git commit -m "feat(frontend): add reMarkable registration to Settings page"
```

---

### Task 12: Confidence Highlighting in Review Editor

**Files:**
- Modify: `frontend/src/components/bujo/OCRReviewPanel.tsx`

**Step 1: Add confidence markers to reconstructed text**

Update `reconstructText` to insert confidence markers:

```tsx
interface TextSpan {
  text: string
  lowConfidence: boolean
}

function reconstructTextWithConfidence(results: remarkable.OCRResult[], threshold = 0.8): { text: string, lowConfidenceRanges: Array<{ from: number, to: number }> } {
  if (!results || results.length === 0) return { text: '', lowConfidenceRanges: [] }

  const sorted = [...results].sort((a, b) => a.Y - b.Y)
  const minX = Math.min(...sorted.map(r => r.X))
  const indentWidth = 50

  let text = ''
  const lowConfidenceRanges: Array<{ from: number, to: number }> = []

  sorted.forEach((r, i) => {
    const depth = Math.round((r.X - minX) / indentWidth)
    const indent = '  '.repeat(depth)
    const line = indent + r.Text

    const from = text.length
    text += line
    if (i < sorted.length - 1) text += '\n'

    if (r.Confidence < threshold) {
      lowConfidenceRanges.push({ from: from + indent.length, to: from + line.length })
    }
  })

  return { text, lowConfidenceRanges }
}
```

Add a visual indicator below the textarea showing low-confidence line numbers:

```tsx
const { lowConfidenceRanges } = useMemo(() => {
  return pages.map(p => reconstructTextWithConfidence(p.OCRResults))
}, [pages])[currentPage] ?? { lowConfidenceRanges: [] }

// Below textarea:
{lowConfidenceRanges.length > 0 && (
  <p className="text-xs text-amber-500 mt-2">
    ⚠ {lowConfidenceRanges.length} low-confidence region(s) detected — review highlighted text carefully
  </p>
)}
```

**Step 2: Verify**

Run: `cd frontend && npm run build`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add frontend/src/components/bujo/OCRReviewPanel.tsx
git commit -m "feat(frontend): add low-confidence OCR indicators to review panel"
```

---

### Task 13: End-to-End Smoke Test

**Step 1: Build and run**

```bash
cd /Users/andrew/Development/bujo
go test ./internal/adapter/wails/ -v
cd frontend && npm run build
cd .. && wails build
```

**Step 2: Manual verification**

1. Launch the app
2. Verify reMarkable sidebar entry appears (macOS only, OCR binary must be present)
3. If not registered: click reMarkable → see "not registered" message → link to Settings
4. Go to Settings → register with reMarkable code
5. Return to reMarkable view → see document list
6. Select a notebook → see importing progress
7. Review OCR results side-by-side → edit text → set date → import
8. Verify entries appear in Journal view

**Step 3: Commit any fixes**

```bash
git add -A
git commit -m "fix: address issues found during reMarkable integration smoke test"
```
