# Native Go .rm Renderer Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace Python subprocess pipeline (rmc, cairosvg, Pillow) with a native Go .rm v6 parser and PNG renderer.

**Architecture:** Parse .rm v6 binary format to extract stroke coordinates, render anti-aliased polylines to PNG using fogleman/gg. Same `RenderPageToPNG()` signature, no subprocess calls.

**Tech Stack:** Go 1.23, fogleman/gg (2D graphics), encoding/binary (stdlib)

**Reference:** ricklupton/rmscene (Python canonical v6 parser) for binary format details

---

### Task 1: Add fogleman/gg dependency

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

**Step 1: Add the dependency**

Run: `go get github.com/fogleman/gg`

**Step 2: Verify it installed**

Run: `go mod tidy && grep fogleman go.mod`
Expected: Line containing `github.com/fogleman/gg`

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add fogleman/gg for native .rm rendering"
```

---

### Task 2: Parse .rm v6 header and block structure

**Files:**
- Create: `internal/adapter/remarkable/rmparse.go`
- Create: `internal/adapter/remarkable/rmparse_test.go`

This task implements the binary format reader: header validation, varuint decoding, and top-level block iteration. The .rm v6 format starts with a 43-byte header `"reMarkable .lines file, version=6          "`, then a sequence of blocks. Each block has: 4 bytes uint32 LE length, 1 byte unknown, 1 byte min_version, 1 byte current_version, 1 byte block_type, then content bytes.

**Step 1: Write failing test for header validation**

```go
package remarkable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHeader_Valid(t *testing.T) {
	header := []byte("reMarkable .lines file, version=6          ")
	r := newRMReader(header)
	err := r.parseHeader()
	require.NoError(t, err)
}

func TestParseHeader_InvalidMagic(t *testing.T) {
	header := []byte("not a remarkable file")
	r := newRMReader(header)
	err := r.parseHeader()
	assert.ErrorContains(t, err, "invalid .rm header")
}

func TestParseHeader_WrongVersion(t *testing.T) {
	header := []byte("reMarkable .lines file, version=5          ")
	r := newRMReader(header)
	err := r.parseHeader()
	assert.ErrorContains(t, err, "unsupported .rm version")
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestParseHeader ./internal/adapter/remarkable/...`
Expected: FAIL (newRMReader not defined)

**Step 3: Implement header parsing and varuint decoding**

```go
package remarkable

import (
	"encoding/binary"
	"fmt"
	"math"
)

const rmHeaderV6 = "reMarkable .lines file, version=6          "

const (
	blockTypeSceneLineItem = 0x05
)

type rmPoint struct {
	X float32
	Y float32
}

type rmStroke struct {
	Points []rmPoint
}

type rmReader struct {
	data []byte
	pos  int
}

func newRMReader(data []byte) *rmReader {
	return &rmReader{data: data, pos: 0}
}

func (r *rmReader) parseHeader() error {
	if len(r.data) < len(rmHeaderV6) {
		return fmt.Errorf("invalid .rm header: file too short")
	}
	header := string(r.data[:len(rmHeaderV6)])
	if header != rmHeaderV6 {
		if len(header) > 33 && header[:33] == "reMarkable .lines file, version=" {
			return fmt.Errorf("unsupported .rm version: %s", header[33:])
		}
		return fmt.Errorf("invalid .rm header")
	}
	r.pos = len(rmHeaderV6)
	return nil
}

func (r *rmReader) remaining() int {
	return len(r.data) - r.pos
}

func (r *rmReader) readUint8() (uint8, error) {
	if r.remaining() < 1 {
		return 0, fmt.Errorf("unexpected EOF at offset %d", r.pos)
	}
	v := r.data[r.pos]
	r.pos++
	return v, nil
}

func (r *rmReader) readUint16() (uint16, error) {
	if r.remaining() < 2 {
		return 0, fmt.Errorf("unexpected EOF at offset %d", r.pos)
	}
	v := binary.LittleEndian.Uint16(r.data[r.pos:])
	r.pos += 2
	return v, nil
}

func (r *rmReader) readUint32() (uint32, error) {
	if r.remaining() < 4 {
		return 0, fmt.Errorf("unexpected EOF at offset %d", r.pos)
	}
	v := binary.LittleEndian.Uint32(r.data[r.pos:])
	r.pos += 4
	return v, nil
}

func (r *rmReader) readFloat32() (float32, error) {
	bits, err := r.readUint32()
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(bits), nil
}

func (r *rmReader) readFloat64() (float64, error) {
	if r.remaining() < 8 {
		return 0, fmt.Errorf("unexpected EOF at offset %d", r.pos)
	}
	bits := binary.LittleEndian.Uint64(r.data[r.pos:])
	r.pos += 8
	return math.Float64frombits(bits), nil
}

func (r *rmReader) readVaruint() (uint64, error) {
	var result uint64
	var shift uint
	for {
		if r.remaining() < 1 {
			return 0, fmt.Errorf("unexpected EOF in varuint at offset %d", r.pos)
		}
		b := r.data[r.pos]
		r.pos++
		result |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			return result, nil
		}
		shift += 7
	}
}

func (r *rmReader) skip(n int) error {
	if r.remaining() < n {
		return fmt.Errorf("unexpected EOF: cannot skip %d bytes at offset %d", n, r.pos)
	}
	r.pos += n
	return nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test -v -run TestParseHeader ./internal/adapter/remarkable/...`
Expected: PASS

**Step 5: Write failing test for varuint decoding**

```go
func TestReadVaruint(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint64
	}{
		{"single byte", []byte{0x7F}, 127},
		{"two bytes", []byte{0x80, 0x01}, 128},
		{"example from spec", []byte{0x8C, 0x01}, 140},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRMReader(tt.data)
			v, err := r.readVaruint()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, v)
		})
	}
}
```

**Step 6: Run test to verify it passes** (already implemented above)

Run: `go test -v -run TestReadVaruint ./internal/adapter/remarkable/...`
Expected: PASS

**Step 7: Write failing test for block reading**

```go
func TestReadBlock(t *testing.T) {
	// Build a minimal block: 4 bytes length + 1 unknown + 1 min_ver + 1 cur_ver + 1 type + content
	content := []byte{0xAA, 0xBB}
	block := make([]byte, 0, 8+len(content))
	block = binary.LittleEndian.AppendUint32(block, uint32(len(content)+4)) // length includes the 4 header bytes after length
	block = append(block, 0x00)       // unknown
	block = append(block, 0x01)       // min_version
	block = append(block, 0x01)       // current_version
	block = append(block, 0x05)       // block_type = SceneLineItem
	block = append(block, content...) // content

	r := newRMReader(block)
	blockType, blockVersion, blockContent, err := r.readBlock()
	require.NoError(t, err)
	assert.Equal(t, uint8(0x05), blockType)
	assert.Equal(t, uint8(0x01), blockVersion)
	assert.Equal(t, content, blockContent)
}
```

**Step 8: Implement block reading**

Add to `rmparse.go`:

```go
func (r *rmReader) readBlock() (blockType uint8, version uint8, content []byte, err error) {
	length, err := r.readUint32()
	if err != nil {
		return 0, 0, nil, fmt.Errorf("reading block length: %w", err)
	}

	_, err = r.readUint8() // unknown byte
	if err != nil {
		return 0, 0, nil, err
	}

	_, err = r.readUint8() // min_version
	if err != nil {
		return 0, 0, nil, err
	}

	version, err = r.readUint8() // current_version
	if err != nil {
		return 0, 0, nil, err
	}

	blockType, err = r.readUint8()
	if err != nil {
		return 0, 0, nil, err
	}

	// length includes the 4 bytes we just read (unknown + min_ver + cur_ver + type)
	contentLen := int(length) - 4
	if contentLen < 0 || r.remaining() < contentLen {
		return 0, 0, nil, fmt.Errorf("invalid block length %d at offset %d", length, r.pos)
	}

	content = r.data[r.pos : r.pos+contentLen]
	r.pos += contentLen
	return blockType, version, content, nil
}
```

**Step 9: Run tests**

Run: `go test -v -run "TestParseHeader|TestReadVaruint|TestReadBlock" ./internal/adapter/remarkable/...`
Expected: PASS

**Step 10: Commit**

```bash
git add internal/adapter/remarkable/rmparse.go internal/adapter/remarkable/rmparse_test.go
git commit -m "feat(remarkable): add .rm v6 header and block parser"
```

---

### Task 3: Extract stroke points from SceneLineItem blocks

**Files:**
- Modify: `internal/adapter/remarkable/rmparse.go`
- Modify: `internal/adapter/remarkable/rmparse_test.go`

This task adds the tagged-value parser to navigate inside SceneLineItem blocks and extract point coordinates. The block content uses a tag system: each value starts with a varuint where lower 4 bits = tag type (0x4=uint32, 0x8=float64, 0xC=length-prefixed subblock, 0xF=CrdtId) and upper bits = field index.

For SceneLineItem (block type 0x05), the value subblock (field index 6, tag type 0xC) contains: 1 byte item_type (must be 0x03 for lines), then tagged fields including field 5 (tag type 0xC) which holds the raw point data.

Points are encoded as either 24 bytes (v1: 6 x float32) or 14 bytes (v2: 2 x float32 + 2 x uint16 + 2 x uint8). We only need the first two floats (X, Y) from each point.

**Step 1: Write failing test for tag parsing**

```go
func TestReadTagged(t *testing.T) {
	// Tag varuint = (index << 4) | tagType
	// index=1, tagType=4 (Byte4) => varuint = (1 << 4) | 4 = 0x14
	data := []byte{0x14}
	data = binary.LittleEndian.AppendUint32(data, 42)
	r := newRMReader(data)

	index, tagType, err := r.readTag()
	require.NoError(t, err)
	assert.Equal(t, uint64(1), index)
	assert.Equal(t, uint8(0x4), tagType)
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestReadTagged ./internal/adapter/remarkable/...`
Expected: FAIL (readTag not defined)

**Step 3: Implement tag reading and skipTaggedValue**

Add to `rmparse.go`:

```go
const (
	tagByte1   = 0x1
	tagByte4   = 0x4
	tagByte8   = 0x8
	tagLength4 = 0xC
	tagID      = 0xF
)

func (r *rmReader) readTag() (index uint64, tagType uint8, err error) {
	v, err := r.readVaruint()
	if err != nil {
		return 0, 0, err
	}
	return v >> 4, uint8(v & 0xF), nil
}

func (r *rmReader) skipCrdtID() error {
	if _, err := r.readUint8(); err != nil {
		return err
	}
	if _, err := r.readVaruint(); err != nil {
		return err
	}
	return nil
}

func (r *rmReader) skipTaggedValue(tagType uint8) error {
	switch tagType {
	case tagByte1:
		return r.skip(1)
	case tagByte4:
		return r.skip(4)
	case tagByte8:
		return r.skip(8)
	case tagLength4:
		length, err := r.readUint32()
		if err != nil {
			return err
		}
		return r.skip(int(length))
	case tagID:
		return r.skipCrdtID()
	default:
		return fmt.Errorf("unknown tag type 0x%x at offset %d", tagType, r.pos)
	}
}
```

**Step 4: Run test**

Run: `go test -v -run TestReadTagged ./internal/adapter/remarkable/...`
Expected: PASS

**Step 5: Write failing test for point extraction from a SceneLineItem block**

Build a synthetic SceneLineItem content blob with known points:

```go
func TestParseLineItemPoints(t *testing.T) {
	// We'll test the full ParseRM function with a complete .rm file
	// containing a single SceneLineItem block with 2 points
	strokes, err := buildTestFileAndParse(t, []rmPoint{
		{X: 100.0, Y: 200.0},
		{X: 150.0, Y: 250.0},
	})
	require.NoError(t, err)
	require.Len(t, strokes, 1)
	require.Len(t, strokes[0].Points, 2)
	assert.InDelta(t, 100.0, strokes[0].Points[0].X, 0.01)
	assert.InDelta(t, 200.0, strokes[0].Points[0].Y, 0.01)
	assert.InDelta(t, 150.0, strokes[0].Points[1].X, 0.01)
	assert.InDelta(t, 250.0, strokes[0].Points[1].Y, 0.01)
}
```

Note: `buildTestFileAndParse` is a test helper that constructs a valid .rm v6 binary with the given points. The implementer should build this helper by writing a valid v6 file: header + one SceneLineItem block with properly tagged fields. Use the format details from Task 2's description. The point data for v1 format is 6 float32s per point (X, Y, speed, direction, width, pressure) — set non-X/Y fields to 0.

**Step 6: Run test to verify it fails**

Run: `go test -v -run TestParseLineItemPoints ./internal/adapter/remarkable/...`
Expected: FAIL (buildTestFileAndParse and ParseRM not defined)

**Step 7: Implement SceneLineItem parsing and the top-level ParseRM function**

Add to `rmparse.go`:

```go
func (r *rmReader) parseLineItemContent(blockVersion uint8, content []byte) (rmStroke, error) {
	cr := newRMReader(content)
	var stroke rmStroke

	for cr.remaining() > 0 {
		index, tagType, err := cr.readTag()
		if err != nil {
			return stroke, err
		}

		// Field 6 (value subblock) contains the line data
		if index == 6 && tagType == tagLength4 {
			length, err := cr.readUint32()
			if err != nil {
				return stroke, err
			}
			subEnd := cr.pos + int(length)

			// First byte is item type (0x03 = line)
			itemType, err := cr.readUint8()
			if err != nil {
				return stroke, err
			}
			if itemType != 0x03 {
				cr.pos = subEnd
				continue
			}

			// Parse tagged fields within the line value
			for cr.pos < subEnd {
				fi, ft, err := cr.readTag()
				if err != nil {
					break
				}
				// Field 5 in the line value = points subblock
				if fi == 5 && ft == tagLength4 {
					pointsLen, err := cr.readUint32()
					if err != nil {
						return stroke, err
					}
					stroke.Points, err = cr.parsePoints(blockVersion, int(pointsLen))
					if err != nil {
						return stroke, err
					}
				} else {
					if err := cr.skipTaggedValue(ft); err != nil {
						return stroke, err
					}
				}
			}
			cr.pos = subEnd
		} else {
			if err := cr.skipTaggedValue(tagType); err != nil {
				return stroke, err
			}
		}
	}
	return stroke, nil
}

func (r *rmReader) parsePoints(blockVersion uint8, dataLen int) ([]rmPoint, error) {
	end := r.pos + dataLen
	var points []rmPoint

	if blockVersion >= 2 {
		// v2: 14 bytes per point (2 float32 + 2 uint16 + 2 uint8)
		pointSize := 14
		for r.pos+pointSize <= end {
			x, _ := r.readFloat32()
			y, _ := r.readFloat32()
			r.skip(2 + 2 + 1 + 1) // speed, width, direction, pressure
			points = append(points, rmPoint{X: x, Y: y})
		}
	} else {
		// v1: 24 bytes per point (6 float32)
		pointSize := 24
		for r.pos+pointSize <= end {
			x, _ := r.readFloat32()
			y, _ := r.readFloat32()
			r.skip(4 * 4) // speed, direction, width, pressure
			points = append(points, rmPoint{X: x, Y: y})
		}
	}
	r.pos = end
	return points, nil
}

func ParseRM(data []byte) ([]rmStroke, error) {
	r := newRMReader(data)
	if err := r.parseHeader(); err != nil {
		return nil, err
	}

	var strokes []rmStroke
	for r.remaining() > 0 {
		blockType, version, content, err := r.readBlock()
		if err != nil {
			return strokes, nil // end of file, return what we have
		}

		if blockType == blockTypeSceneLineItem {
			stroke, err := r.parseLineItemContent(version, content)
			if err != nil {
				continue // skip malformed strokes
			}
			if len(stroke.Points) > 0 {
				strokes = append(strokes, stroke)
			}
		}
	}
	return strokes, nil
}
```

**Step 8: Implement the test helper buildTestFileAndParse**

The implementer needs to construct a valid v6 binary for testing. The helper should:
1. Write the 43-byte header
2. Write a SceneLineItem block (type 0x05) with properly encoded content:
   - Tagged fields for parent_id, item_id, left_id, right_id (all CrdtId, can use dummy values)
   - Field 5 (deleted_length) as Byte4
   - Field 6 (value subblock) as Length4, containing:
     - Item type byte (0x03)
     - Tagged field 1 (tool) as Byte4
     - Tagged field 2 (color) as Byte4
     - Tagged field 3 (thickness) as Byte8
     - Tagged field 4 (starting_length) as Byte4
     - Tagged field 5 (points) as Length4 with v1 point data (24 bytes each)

This is the hardest part — getting the binary encoding right. Reference the format specification above. If encoding the full tagged structure proves too complex, an alternative approach is to capture a small real .rm file from the rmscene test suite and embed it as a test fixture.

**Step 9: Run tests**

Run: `go test -v -run "TestParseLineItemPoints|TestParseHeader|TestReadTagged" ./internal/adapter/remarkable/...`
Expected: PASS

**Step 10: Commit**

```bash
git add internal/adapter/remarkable/rmparse.go internal/adapter/remarkable/rmparse_test.go
git commit -m "feat(remarkable): parse stroke points from .rm v6 SceneLineItem blocks"
```

---

### Task 4: Render strokes to PNG

**Files:**
- Create: `internal/adapter/remarkable/rmrender.go`
- Create: `internal/adapter/remarkable/rmrender_test.go`

This task renders parsed strokes to a PNG image using fogleman/gg. The reMarkable screen is 1404x1872 pixels. We draw black anti-aliased polylines on a white background.

**Step 1: Write failing test for PNG rendering**

```go
package remarkable

import (
	"image/png"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderStrokes_Dimensions(t *testing.T) {
	strokes := []rmStroke{
		{Points: []rmPoint{{X: 100, Y: 100}, {X: 200, Y: 200}}},
	}

	data, err := RenderStrokes(strokes)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	bounds := img.Bounds()
	assert.Equal(t, remarkableScreenWidth, bounds.Max.X)
	assert.Equal(t, remarkableScreenHeight, bounds.Max.Y)
}

func TestRenderStrokes_EmptyStrokes(t *testing.T) {
	data, err := RenderStrokes(nil)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	bounds := img.Bounds()
	assert.Equal(t, remarkableScreenWidth, bounds.Max.X)
}

func TestRenderStrokes_WhiteBackground(t *testing.T) {
	strokes := []rmStroke{}

	data, err := RenderStrokes(strokes)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	// Check corner pixel is white
	r, g, b, a := img.At(0, 0).RGBA()
	assert.Equal(t, uint32(0xFFFF), r)
	assert.Equal(t, uint32(0xFFFF), g)
	assert.Equal(t, uint32(0xFFFF), b)
	assert.Equal(t, uint32(0xFFFF), a)
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestRenderStrokes ./internal/adapter/remarkable/...`
Expected: FAIL (RenderStrokes not defined)

**Step 3: Implement RenderStrokes**

```go
package remarkable

import (
	"bytes"
	"image/png"

	"github.com/fogleman/gg"
)

const remarkableScreenHeight = 1872

func RenderStrokes(strokes []rmStroke) ([]byte, error) {
	dc := gg.NewContext(remarkableScreenWidth, remarkableScreenHeight)

	// White background
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	// Black strokes
	dc.SetRGB(0, 0, 0)
	dc.SetLineWidth(1)

	for _, stroke := range strokes {
		if len(stroke.Points) < 2 {
			continue
		}
		dc.MoveTo(float64(stroke.Points[0].X), float64(stroke.Points[0].Y))
		for _, p := range stroke.Points[1:] {
			dc.LineTo(float64(p.X), float64(p.Y))
		}
		dc.Stroke()
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test -v -run TestRenderStrokes ./internal/adapter/remarkable/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adapter/remarkable/rmrender.go internal/adapter/remarkable/rmrender_test.go
git commit -m "feat(remarkable): render strokes to PNG with fogleman/gg"
```

---

### Task 5: Replace RenderPageToPNG with native implementation

**Files:**
- Modify: `internal/adapter/remarkable/render.go`
- Modify: `internal/adapter/remarkable/render_test.go`

This task replaces the subprocess-based `RenderPageToPNG` with the native parser + renderer. Remove `BuildRmcCommand`, `BuildCairoSVGCommand`, and `SavePageToFile`. The function signature stays the same: `RenderPageToPNG(dir string, pageID string, rmData []byte) (string, error)`.

**Step 1: Write failing test for the new RenderPageToPNG**

Replace the existing render_test.go contents:

```go
package remarkable

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderPageToPNG_WritesFile(t *testing.T) {
	dir := t.TempDir()
	// Minimal valid .rm v6: header only, no blocks = blank page
	rmData := []byte("reMarkable .lines file, version=6          ")

	path, err := RenderPageToPNG(dir, "test-page", rmData)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "test-page.png"), path)

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	img, err := png.Decode(f)
	require.NoError(t, err)
	bounds := img.Bounds()
	assert.Equal(t, remarkableScreenWidth, bounds.Max.X)
	assert.Equal(t, remarkableScreenHeight, bounds.Max.Y)
}

func TestRenderPageToPNG_InvalidRM(t *testing.T) {
	dir := t.TempDir()
	_, err := RenderPageToPNG(dir, "bad-page", []byte("not a valid rm file"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid .rm header")
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestRenderPageToPNG ./internal/adapter/remarkable/...`
Expected: FAIL (old RenderPageToPNG tries to run rmc subprocess)

**Step 3: Replace render.go implementation**

Replace `render.go` with:

```go
package remarkable

import (
	"fmt"
	"os"
	"path/filepath"
)

const remarkableScreenWidth = 1404

func RenderPageToPNG(dir string, pageID string, rmData []byte) (string, error) {
	strokes, err := ParseRM(rmData)
	if err != nil {
		return "", fmt.Errorf("parse .rm failed: %w", err)
	}

	pngData, err := RenderStrokes(strokes)
	if err != nil {
		return "", fmt.Errorf("render failed: %w", err)
	}

	pngPath := filepath.Join(dir, pageID+".png")
	if err := os.WriteFile(pngPath, pngData, 0644); err != nil {
		return "", fmt.Errorf("write PNG failed: %w", err)
	}

	return pngPath, nil
}
```

Note: `remarkableScreenWidth` is already defined in `render.go`. When moving `remarkableScreenHeight` to `rmrender.go`, ensure no duplicate constant definitions. The implementer should check which file owns each constant and resolve any conflicts.

**Step 4: Run tests to verify they pass**

Run: `go test -v -run TestRenderPageToPNG ./internal/adapter/remarkable/...`
Expected: PASS

**Step 5: Run all remarkable tests**

Run: `go test -v ./internal/adapter/remarkable/...`
Expected: All pass. The old `TestBuildRmcCommand`, `TestBuildCairoSVGCommand`, `TestSavePageToFile` tests should be deleted since those functions no longer exist.

**Step 6: Commit**

```bash
git add internal/adapter/remarkable/render.go internal/adapter/remarkable/render_test.go
git commit -m "feat(remarkable): replace Python subprocess pipeline with native Go renderer"
```

---

### Task 6: Clean up Python dependencies and old code

**Files:**
- Modify: `internal/adapter/remarkable/render.go` (verify no leftover imports)
- Verify: `cmd/bujo/cmd/remarkable.go` (no changes needed — uses same RenderPageToPNG signature)

**Step 1: Verify no Python references remain**

Run: `grep -r "rmc\|cairosvg\|python3\|Pillow\|PIL\|pip" internal/ cmd/ --include="*.go"`
Expected: No matches (or only in comments/docs)

**Step 2: Verify the build compiles**

Run: `go build ./cmd/bujo/...`
Expected: Success

**Step 3: Run all tests**

Run: `go test ./...`
Expected: All pass

**Step 4: Verify no unused imports in render.go**

The new render.go should NOT import `os/exec`. If it does, remove it.

Run: `go vet ./...`
Expected: Clean

**Step 5: Commit**

```bash
git add -A
git commit -m "chore: remove Python rendering dependencies"
```

---

### Task 7: Integration test with real .rm fixture

**Files:**
- Create: `internal/adapter/remarkable/testdata/` (directory for test fixtures)
- Modify: `internal/adapter/remarkable/rmparse_test.go`

Download a small .rm v6 test file from the rmscene project's test suite (or capture one from the user's reMarkable). Place it in `testdata/`. Write a test that parses it and verifies strokes are extracted.

**Step 1: Obtain a test fixture**

Option A: Download from rmscene test suite:
Run: `curl -L -o internal/adapter/remarkable/testdata/test_strokes.rm "https://github.com/ricklupton/rmscene/raw/main/tests/data/Lines_v2.rm"`

Option B: If the above URL doesn't work, use a minimal synthetic fixture created programmatically (from Task 3's test helper).

**Step 2: Write integration test**

```go
func TestParseRM_RealFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/test_strokes.rm")
	if os.IsNotExist(err) {
		t.Skip("test fixture not available")
	}
	require.NoError(t, err)

	strokes, err := ParseRM(data)
	require.NoError(t, err)
	assert.Greater(t, len(strokes), 0, "expected at least one stroke")

	for _, s := range strokes {
		assert.Greater(t, len(s.Points), 1, "each stroke should have multiple points")
	}
}

func TestParseAndRender_RealFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/test_strokes.rm")
	if os.IsNotExist(err) {
		t.Skip("test fixture not available")
	}
	require.NoError(t, err)

	strokes, err := ParseRM(data)
	require.NoError(t, err)

	pngData, err := RenderStrokes(strokes)
	require.NoError(t, err)
	assert.Greater(t, len(pngData), 1000, "PNG should have meaningful content")
}
```

**Step 3: Run tests**

Run: `go test -v -run "TestParseRM_RealFixture|TestParseAndRender_RealFixture" ./internal/adapter/remarkable/...`
Expected: PASS (or SKIP if fixture unavailable)

**Step 4: Commit**

```bash
git add internal/adapter/remarkable/testdata/ internal/adapter/remarkable/rmparse_test.go
git commit -m "test(remarkable): add integration test with real .rm v6 fixture"
```
