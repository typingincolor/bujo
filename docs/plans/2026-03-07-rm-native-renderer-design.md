# Native Go .rm Renderer Design

## Goal

Eliminate all Python dependencies (rmc, cairosvg, Pillow) from the reMarkable import pipeline by implementing a native Go .rm v6 parser and PNG renderer.

## Context

The current pipeline shells out to Python tools:
1. `rmc` (Python/pip) parses .rm binary → SVG
2. `cairosvg` (Python/pip) converts SVG → PNG
3. `Pillow` (Python/pip) flattens alpha channel for OCR compatibility

This requires a Python venv with three packages installed. The native Go approach replaces all three with a single Go function.

## Architecture

```
.rm binary bytes (v6 format)
    |
    v
Go v6 parser (minimal extraction - stroke coordinates only)
    |
    v
fogleman/gg renderer (anti-aliased polylines on white background)
    |
    v
PNG (1404x1872, white background, RGB)
```

### Component 1: Parser (`rmparse.go`)

Minimal extraction from .rm v6 binary format:
- Validate v6 header (`reMarkable .lines file, version=6`)
- Walk binary blocks, extract only SceneLineItem blocks (stroke data)
- For each stroke: extract point array (X, Y coordinates)
- Ignore CRDT metadata, tombstones, text blocks, pressure, tilt
- Return `[]Stroke` where `Stroke` contains `[]Point{X, Y float32}`
- Return error on unsupported versions (no fallback)

The v6 format uses CRDT serialization with tagged blocks (TLV-like). We parse just enough structure to locate stroke coordinate data.

Reference implementations:
- `ricklupton/rmscene` (Python, canonical v6 parser)
- `rorycl/rm2pdf/rmparsev6` (Go, incomplete but shows data structures)

### Component 2: Renderer (`rmrender.go`)

Render extracted strokes to PNG using `fogleman/gg`:
- Canvas: 1404x1872 pixels (reMarkable screen dimensions)
- White background (RGB, no alpha - OCR compatible)
- Black strokes, 1px anti-aliased polylines
- Accept `[]Stroke`, return PNG as `[]byte` or write to file path

### Component 3: Integration (`render.go` refactor)

- Replace `RenderPageToPNG()` internals with native parser + renderer
- Remove `BuildRmcCommand()` and `BuildCairoSVGCommand()`
- Same function signature, no subprocess calls
- Remove `SavePageToFile()` (no longer need temp .rm files on disk)

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Rendering fidelity | Minimal (1px polylines) | PNGs are only used for OCR input, not display |
| Format parsing depth | Minimal extraction | Only need stroke coordinates, not CRDT metadata |
| Drawing library | fogleman/gg | Anti-aliased lines, single dependency, clean API |
| Unsupported versions | Fail with error | No fallback to Python - clean elimination |

## Error Handling

- Unsupported .rm version: return descriptive error, caller skips page
- Malformed binary data: return error with byte offset context
- No fallback to external tools

## Dependencies

- Add: `github.com/fogleman/gg`
- Remove: Python venv, rmc, cairosvg, Pillow

## Testing Strategy

- Parser: unit tests against real .rm v6 fixture files (from rmscene test suite)
- Renderer: unit tests verify valid PNG output with correct dimensions
- Integration: parse + render produces non-empty RGB PNG
