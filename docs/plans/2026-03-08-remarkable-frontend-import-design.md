# reMarkable Frontend Import Design

Date: 2026-03-08

## Overview

Integrate the reMarkable handwriting import pipeline into the Wails desktop app. Users select a reMarkable notebook, pages are downloaded and processed through OCR, and the recognized text is reviewed in a side-by-side editor before importing into the journal.

## Navigation & View

- New **reMarkable** sidebar entry, shown only on macOS when OCR binary is available
- Platform check via `GetPlatformCapabilities()` returning `{hasOCR: bool, platform: string}`
- Registration flow triggered from **Settings page**, not the reMarkable view
- Three-step flow within the view:
  1. **Document List** — selectable notebook list from reMarkable Cloud
  2. **Page Preview + Auto-OCR** — thumbnail grid with loading/processing states
  3. **Side-by-side Review** — PNG on left, CodeMirror editor on right, low-confidence spans highlighted

## Data Flow

1. `ListRemarkableDocuments()` fetches notebook list from reMarkable Cloud
2. User selects a document
3. `ImportRemarkablePages(docID)` runs the full pipeline:
   - Downloads all pages from reMarkable Cloud
   - Renders each page to PNG via `RenderPageToPNG`
   - Runs OCR on each PNG via the Swift `remarkable-ocr` binary
   - Returns `{pages: [{png: base64, ocrResults: [...]}]}`
4. Frontend `ReconstructText()` assembles OCR results into editable bullet journal text
5. User reviews in side-by-side view with confidence highlighting
6. `ImportEntries(text, date)` parses reviewed text via TreeParser, inserts into DB, emits `data:changed`

## Backend Methods (Wails)

| Method | Signature | Purpose |
|--------|-----------|---------|
| `GetPlatformCapabilities` | `() -> {hasOCR: bool, platform: string}` | Platform and OCR availability check |
| `ListRemarkableDocuments` | `() -> []Document` | Fetch notebook list from cloud |
| `ImportRemarkablePages` | `(docID string) -> ImportResult` | Download + render + OCR pipeline |
| `ImportEntries` | `(text string, date string) -> error` | Parse and insert into journal |

## OCR Binary Discovery

Single discovery mechanism for both deployment targets:

```go
execPath, _ := os.Executable()
ocrPath := filepath.Join(filepath.Dir(execPath), "remarkable-ocr")
```

| Target | Binary Location | How it gets there |
|--------|----------------|-------------------|
| Wails `.app` | `Contents/MacOS/remarkable-ocr` | Wails build bundles it |
| Homebrew CLI | `$(brew --prefix)/bin/remarkable-ocr` | Formula builds Swift + installs |

Both cases place `remarkable-ocr` next to the main `bujo` binary, so `os.Executable()` resolves correctly for both.

### Homebrew Formula

```ruby
def install
  system "go", "build", "-o", bin/"bujo", "./cmd/bujo"

  cd "tools/remarkable-ocr" do
    system "swift", "build", "-c", "release"
    bin.install ".build/release/remarkable-ocr"
  end
end
```

## Frontend Components

- **`RemarkableView`** — container with step state machine (list → preview → review)
- **`DocumentList`** — selectable notebook list with metadata (title, last modified)
- **`PagePreview`** — thumbnail grid with loading/processing indicators per page
- **`OCRReviewPanel`** — side-by-side layout:
  - Left: rendered PNG image (scrollable, zoomable)
  - Right: CodeMirror editor with reconstructed text
  - Low-confidence spans highlighted for user correction
- Date auto-detected from notebook metadata, displayed and editable by user

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Cloud auth expired | Redirect to Settings for re-registration |
| OCR fails on a page | Show page with "OCR failed" message, allow skip |
| No notebooks found | Empty state with guidance message |
| Network errors | Retry with backoff, user-visible error after 3 failures |
| OCR binary not found | `hasOCR: false`, reMarkable sidebar entry hidden |

## Platform Constraints

- macOS only (Apple Vision framework dependency for OCR)
- Non-macOS platforms: reMarkable sidebar entry not shown
- Graceful degradation: if OCR binary missing, feature hidden entirely
