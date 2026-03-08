# reMarkable Integration

This document describes the current reMarkable integration in bujo based on code and plans.

## Scope

- Register a device with reMarkable Cloud
- List cloud documents (including folder metadata)
- Import notebook pages through render + OCR pipeline
- Review/edit OCR text in desktop UI
- Import reviewed text into journal entries

## Where It Lives

- CLI harness: `cmd/bujo/cmd/remarkable.go`
- Cloud/API adapter: `internal/adapter/remarkable/client.go`
- Config storage: `internal/adapter/remarkable/config.go`
- Document model and zip handling: `internal/adapter/remarkable/document.go`
- Content file parsing: `internal/adapter/remarkable/content.go`
- `.rm` parser/renderer: `internal/adapter/remarkable/rmparse.go`, `rmrender.go`, `render.go`
- OCR runner: `internal/adapter/remarkable/ocr.go`
- Text reconstruction/normalization: `internal/adapter/remarkable/reconstruct.go`, `normalize.go`
- Wails backend methods: `internal/adapter/wails/app.go`
- Desktop UI: `frontend/src/components/bujo/RemarkableView.tsx`, `OCRReviewPanel.tsx`
- Frontend helpers: `frontend/src/components/bujo/remarkableUtils.ts`

## End-to-End Flow

1. User registers in Settings using one-time code from `my.remarkable.com/device/browser/connect`.
2. Device token is saved to `~/.config/bujo/remarkable.json`.
3. Desktop app calls `ListRemarkableDocuments()`.
4. User selects a document in the reMarkable view.
5. `ImportRemarkablePages(docID)`:
   - downloads ordered `.rm` pages via `.content`
   - renders each page to PNG (native Go renderer)
   - runs `remarkable-ocr` (Swift + Apple Vision)
   - reconstructs text and low-confidence count
6. User reviews/edits text in `OCRReviewPanel`.
7. `ImportEntries(text, date)` normalizes indentation and writes parsed entries to DB.

## Coordinate Systems

reMarkable `.rm` files use two different coordinate systems depending on how the page was created:

### Centered Coordinates (v6 native pages)

Pages created natively on reMarkable v6 firmware use centered coordinates where X ranges from approximately **-702 to +702**. The origin is at the horizontal center of the screen.

### Absolute Coordinates (migrated pages)

Pages migrated from older firmware versions use absolute coordinates where X ranges from approximately **0 to 1404**. The origin is at the top-left corner.

### Auto-Detection

The renderer (`rmrender.go`) auto-detects which coordinate system a page uses by examining the center of the X range across all strokes:

- If the center of X values is closer to 0 → **centered** coordinates → apply +702 pixel offset
- If the center of X values is closer to 702 → **absolute** coordinates → no offset applied

This happens per-page via `detectXOffset()`, so a single notebook can contain pages with different coordinate systems.

### Screen Dimensions

- Width: 1404 pixels
- Height: 1872 pixels

## Text Reconstruction

The `ReconstructText` function (`reconstruct.go`) converts OCR results into structured text:

- OCR results are sorted by Y position (top to bottom)
- Indentation depth is calculated from X position relative to the leftmost text
- Depth normalization prevents jumps greater than 1 level
- **Lines without a recognized bujo symbol prefix automatically get `- ` prepended** (note type), ensuring every line has a valid entry type for the parser

### Recognized Bujo Symbols

The following first characters are recognized as bujo entry type prefixes:

| Symbol | Type |
|--------|------|
| `.` `•` | Task |
| `-` `–` | Note |
| `o` `○` | Event |
| `x` `✓` | Done |
| `>` `→` | Migrated |
| `?` | Question |
| `★` | Answered |
| `a` `↳` | Answer |

Any line not starting with one of these symbols gets `- ` prepended during text reconstruction.

## Parser Behavior with OCR Input

The `TreeParser` (`internal/domain/parser.go`) handles OCR output gracefully:

- **Orphaned indented lines**: If a line is indented deeper than the current parent stack allows (e.g., indented text with no root parent), the depth is clamped to the maximum valid depth rather than producing an error.
- **Unrecognized symbols**: Lines without a valid bujo symbol prefix are treated as notes (`-`) with the full line content preserved.
- **Mixed indentation**: Common in OCR output where handwriting position varies; the parser auto-corrects rather than failing.

## Platform and Binary Requirements

- OCR is macOS-only (Apple Vision dependency).
- OCR binary expected: `remarkable-ocr` next to executable in production, or in `tools/remarkable-ocr/remarkable-ocr` in dev.
- Build OCR tool with `make ocr` (macOS).

## Data Notes

- Registration config path: `~/.config/bujo/remarkable.json`
- Current config stores `device_token`; `device_id` field exists but is not populated in current save path.
- Document list includes `Parent` and `FileType`, used by frontend folder navigation.

## Current UX Behavior

- Sidebar entry `reMarkable Import` is shown when platform capabilities report OCR support.
- Registration is handled in Settings.
- File manager supports folders/breadcrumbs and document sorting.
- Notebooks and PDFs are currently treated as importable in UI.
- Review panel shows page image, editable OCR text, and low-confidence count warning.
- Review panel includes a confidence gutter with amber dots beside lines where OCR confidence is below threshold.

## CLI Testing Guide

### Prerequisites

1. Build the binary: `go build -o bujo ./cmd/bujo`
2. Build the OCR tool: `make ocr` (macOS only)
3. Register with reMarkable Cloud (one-time): `./bujo remarkable register`

### List Available Documents

```bash
./bujo --db-path :memory: remarkable list
```

Output shows document name, type, modified timestamp, and document ID:

```
Bujo                                     notebook   1772974966631        4d8737b5-b17c-461f-9cb4-bc6c4ee88b62
Raghavan Swaminathan                     notebook   1772974996225        7488154c-cc78-4720-a509-e63c76ecfe04
```

### Import a Notebook

Use the document ID from `list`:

```bash
./bujo --db-path :memory: remarkable import <doc-id>
```

Example:

```bash
./bujo --db-path :memory: remarkable import 4d8737b5-b17c-461f-9cb4-bc6c4ee88b62
```

This will:
1. Download all pages from the reMarkable Cloud
2. Render each page to PNG (applying coordinate auto-detection)
3. Run OCR on each PNG using Apple Vision
4. Reconstruct text with indentation and bujo symbol prefixes
5. Parse entries and display the result

### Import to a Test Database

To persist imported entries for further testing:

```bash
./bujo --db-path ./test.db remarkable import <doc-id>
```

Then verify with:

```bash
./bujo --db-path ./test.db list --date today
```

### Run Tests

```bash
# All remarkable adapter tests
go test -v ./internal/adapter/remarkable/...

# Specific test areas
go test -v -run TestRenderStrokes ./internal/adapter/remarkable/...
go test -v -run TestReconstructText ./internal/adapter/remarkable/...
go test -v -run TestDetectXOffset ./internal/adapter/remarkable/...

# Parser tests (domain layer)
go test -v -run TestTreeParser ./internal/domain/...

# Full test suite
go test ./...
```

## Known Gaps / Drift From Plans

- `GetPlatformCapabilities()` currently gates by OS, not confirmed binary presence.
- PDF importability in UI may exceed backend page pipeline assumptions (`DownloadPages` is `.rm` page oriented).
- OCR accuracy is limited by handwriting quality; typos in output are expected and can be corrected in the review panel.

## Testing

Primary coverage is in:

- `internal/adapter/remarkable/*_test.go`
- `internal/adapter/wails/platform_test.go`
- frontend tests under `frontend/src/components/bujo/*.test.tsx`

When changing this feature, follow TDD and keep tests aligned with behavior.
