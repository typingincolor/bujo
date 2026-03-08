# reMarkable Import via Apple Vision OCR

## Pipeline

```
Download (.rm pages) → Render (SVG→PNG) → OCR (Apple Vision) → Reconstruct (indentation from x-coords) → Parse (TreeParser)
```

## CLI Commands

| Command | Input | Output |
|---------|-------|--------|
| `bujo remarkable list` | — | List documents (already working) |
| `bujo remarkable render <id> [--out-dir ./out]` | Document ID | Downloads `.rm` pages, converts to PNGs via `rmc` + `cairosvg`, saves to directory |
| `bujo remarkable ocr <png-or-dir>` | PNG file or directory of PNGs | Runs Apple Vision OCR, outputs JSON with text + bounding boxes to stdout |
| `bujo remarkable import <id>` | Document ID | Full pipeline: render → OCR → reconstruct indentation → TreeParser → print entries to stdout |

## Components

### 1. Page downloader (Go, remarkable adapter)

Uses existing cloud client to get document sub-entries. Reads `.content` for page order from `cPages.pages[].id`. Downloads each `{uuid}/{page_uuid}.rm` file by hash. Saves to temp directory.

### 2. Renderer (Python subprocess)

Shells out to `rmc` to convert `.rm` → SVG. Shells out to `cairosvg` (Python) to convert SVG → PNG. Prerequisites: `pip install rmc cairosvg`.

### 3. Vision OCR tool (Swift CLI, ~50 lines)

Small Swift executable: `remarkable-ocr`. Takes PNG path, runs `VNRecognizeTextRequest`. Outputs JSON array of `{text, x, y, width, height}` per recognized text block. Built with `swiftc`, no Xcode project needed.

### 4. Indentation reconstructor (Go, remarkable adapter)

Takes Vision OCR JSON output. Groups text blocks into lines by y-coordinate proximity. Sorts lines top-to-bottom. Determines indentation level from x-coordinate relative to left margin. Produces indented plain text suitable for TreeParser.

## Dependencies

- `rmc` + `cairosvg` (pip install)
- Swift compiler (ships with Xcode Command Line Tools, already on macOS)
- No Gemini API needed

## Scope

- No database writes (test harness only)
- No folder hierarchy traversal (flat document list)
- No handling of imported PDFs (notebooks only for now)
