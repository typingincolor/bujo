# reMarkable OCR Import Pipeline — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Download reMarkable notebook pages, render to PNG, OCR with Apple Vision preserving layout, reconstruct indentation, and parse with TreeParser.

**Architecture:** Four-stage pipeline — (1) Go downloads `.rm` page files from cloud API using page order from `.content`, (2) Python subprocess converts `.rm` → SVG → PNG via `rmc` + `cairosvg`, (3) Swift CLI tool runs Apple Vision OCR returning JSON with bounding boxes, (4) Go reconstructs indented text from bounding box x-coordinates and feeds to TreeParser.

**Tech Stack:** Go (pipeline orchestration), Python `rmc`+`cairosvg` (rendering), Swift Vision framework (OCR), existing TreeParser (parsing)

---

### Task 1: Parse page order from .content JSON

The `.content` JSON from the cloud API contains `cPages.pages[].id` which gives the correct page order. We need a function to extract this ordered list of page UUIDs.

**Files:**
- Create: `internal/adapter/remarkable/content.go`
- Test: `internal/adapter/remarkable/content_test.go`

**Step 1: Write the failing test**

```go
// content_test.go
package remarkable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePageOrder(t *testing.T) {
	contentJSON := `{
		"cPages": {
			"pages": [
				{"id": "page-uuid-1", "idx": {"timestamp": "1:2", "value": "ba"}},
				{"id": "page-uuid-2", "idx": {"timestamp": "1:3", "value": "bb"}}
			]
		},
		"fileType": "notebook",
		"pageCount": 2
	}`

	pages, err := ParsePageOrder([]byte(contentJSON))
	require.NoError(t, err)
	assert.Equal(t, []string{"page-uuid-1", "page-uuid-2"}, pages)
}

func TestParsePageOrderEmpty(t *testing.T) {
	contentJSON := `{"cPages": {"pages": []}, "fileType": "notebook"}`

	pages, err := ParsePageOrder([]byte(contentJSON))
	require.NoError(t, err)
	assert.Empty(t, pages)
}

func TestParsePageOrderInvalidJSON(t *testing.T) {
	_, err := ParsePageOrder([]byte("not json"))
	assert.Error(t, err)
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestParsePageOrder ./internal/adapter/remarkable/...`
Expected: FAIL — `ParsePageOrder` not defined

**Step 3: Write minimal implementation**

```go
// content.go
package remarkable

import "encoding/json"

type ContentFile struct {
	CPages struct {
		Pages []struct {
			ID string `json:"id"`
		} `json:"pages"`
	} `json:"cPages"`
	FileType string `json:"fileType"`
}

func ParsePageOrder(data []byte) ([]string, error) {
	var content ContentFile
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, err
	}

	pages := make([]string, 0, len(content.CPages.Pages))
	for _, p := range content.CPages.Pages {
		pages = append(pages, p.ID)
	}
	return pages, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestParsePageOrder ./internal/adapter/remarkable/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adapter/remarkable/content.go internal/adapter/remarkable/content_test.go
git commit -m "feat(remarkable): parse page order from .content JSON"
```

---

### Task 2: Download notebook pages from cloud API

Add a method to the Client that downloads all `.rm` page files for a document in correct page order. It fetches the `.content` sub-entry to get page order, then downloads each `{docID}/{pageID}.rm` sub-entry.

**Files:**
- Modify: `internal/adapter/remarkable/client.go`
- Test: `internal/adapter/remarkable/client_test.go`

**Step 1: Write the failing test**

```go
// Add to client_test.go
func TestDownloadPages(t *testing.T) {
	page1Content := []byte("rm-page-1-binary")
	page2Content := []byte("rm-page-2-binary")
	docID := "doc-uuid-1"

	contentJSON := `{"cPages":{"pages":[{"id":"page-a"},{"id":"page-b"}]},"fileType":"notebook"}`

	rootEntries := fmt.Sprintf("3\ndocHash:%s:%s:5:1024\n", "80000000", docID)
	docEntries := fmt.Sprintf("3\ncontentHash:0:%s.content:0:100\nmetaHash:0:%s.metadata:0:50\npageAHash:0:%s/page-a.rm:0:200\npageBHash:0:%s/page-b.rm:0:300\n",
		docID, docID, docID, docID)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token/json/2/user/new" {
			w.Write([]byte("user-token"))
			return
		}
		switch r.URL.Path {
		case "/sync/v4/root":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"hash": "rootHash", "generation": 1, "schemaVersion": 3,
			})
		case "/sync/v3/files/rootHash":
			w.Write([]byte(rootEntries))
		case "/sync/v3/files/docHash":
			w.Write([]byte(docEntries))
		case "/sync/v3/files/contentHash":
			w.Write([]byte(contentJSON))
		case "/sync/v3/files/pageAHash":
			w.Write(page1Content)
		case "/sync/v3/files/pageBHash":
			w.Write(page2Content)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetSyncHost(server.URL)

	pages, err := client.DownloadPages("fake-device-token", docID)
	require.NoError(t, err)
	require.Len(t, pages, 2)
	assert.Equal(t, "page-a", pages[0].PageID)
	assert.Equal(t, page1Content, pages[0].Data)
	assert.Equal(t, "page-b", pages[1].PageID)
	assert.Equal(t, page2Content, pages[1].Data)
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestDownloadPages ./internal/adapter/remarkable/...`
Expected: FAIL — `DownloadPages` not defined, `PageData` type not defined

**Step 3: Write minimal implementation**

Add to `document.go`:
```go
type PageData struct {
	PageID string
	Data   []byte
}
```

Add to `client.go`:
```go
func (c *Client) DownloadPages(deviceToken string, docID string) ([]PageData, error) {
	userToken, err := c.RefreshUserToken(deviceToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	req, err := http.NewRequest("GET", c.syncHost+"/sync/v4/root", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+userToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var root RootHashResponse
	if err := json.NewDecoder(resp.Body).Decode(&root); err != nil {
		return nil, err
	}

	rootEntries, err := c.GetEntries(userToken, root.Hash)
	if err != nil {
		return nil, err
	}

	for _, entry := range rootEntries {
		if entry.ID != docID {
			continue
		}

		subEntries, err := c.GetEntries(userToken, entry.Hash)
		if err != nil {
			return nil, fmt.Errorf("failed to get document entries: %w", err)
		}

		// Find .content to get page order
		var pageOrder []string
		for _, sub := range subEntries {
			if strings.HasSuffix(sub.ID, ".content") {
				data, err := c.GetFileContent(userToken, sub.Hash)
				if err != nil {
					return nil, fmt.Errorf("failed to get content file: %w", err)
				}
				pageOrder, err = ParsePageOrder(data)
				if err != nil {
					return nil, fmt.Errorf("failed to parse page order: %w", err)
				}
				break
			}
		}

		// Build hash lookup for .rm files
		rmHashes := make(map[string]string)
		for _, sub := range subEntries {
			if strings.HasSuffix(sub.ID, ".rm") {
				rmHashes[sub.ID] = sub.Hash
			}
		}

		// Download pages in order
		var pages []PageData
		for _, pageID := range pageOrder {
			rmKey := docID + "/" + pageID + ".rm"
			hash, ok := rmHashes[rmKey]
			if !ok {
				continue
			}
			data, err := c.GetFileContent(userToken, hash)
			if err != nil {
				return nil, fmt.Errorf("failed to download page %s: %w", pageID, err)
			}
			pages = append(pages, PageData{PageID: pageID, Data: data})
		}

		return pages, nil
	}

	return nil, fmt.Errorf("document %s not found", docID)
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestDownloadPages ./internal/adapter/remarkable/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adapter/remarkable/client.go internal/adapter/remarkable/client_test.go internal/adapter/remarkable/document.go
git commit -m "feat(remarkable): download notebook pages in correct order"
```

---

### Task 3: Render .rm pages to PNG via Python subprocess

Shell out to `rmc` to convert `.rm` → SVG, then `cairosvg` to convert SVG → PNG. This is a Go function that writes `.rm` data to a temp file, runs the Python tools, and returns the PNG path.

**Files:**
- Create: `internal/adapter/remarkable/render.go`
- Test: `internal/adapter/remarkable/render_test.go`

**Step 1: Write the failing test**

The render function calls external tools, so we test the command construction and file handling, not the actual rendering. We test `BuildRenderCommands` which returns the shell commands to execute, and `RenderPage` which is the integration function.

```go
// render_test.go
package remarkable

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRmcCommand(t *testing.T) {
	cmd := BuildRmcCommand("/tmp/page.rm", "/tmp/page.svg")
	assert.Equal(t, "rmc", cmd.Path)
	assert.Contains(t, cmd.Args, "-o")
	assert.Contains(t, cmd.Args, "/tmp/page.svg")
	assert.Contains(t, cmd.Args, "/tmp/page.rm")
}

func TestBuildCairoSVGCommand(t *testing.T) {
	cmd := BuildCairoSVGCommand("/tmp/page.svg", "/tmp/page.png")
	assert.Equal(t, "python3", cmd.Path)
	assert.Contains(t, cmd.Args, "cairosvg")
	assert.Contains(t, cmd.Args, "/tmp/page.svg")
	assert.Contains(t, cmd.Args, "/tmp/page.png")
}

func TestSavePageToFile(t *testing.T) {
	dir := t.TempDir()
	data := []byte("fake-rm-data")

	path, err := SavePageToFile(dir, "page-uuid-1", data)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "page-uuid-1.rm"), path)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, data, content)
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run "TestBuild|TestSavePage" ./internal/adapter/remarkable/...`
Expected: FAIL — functions not defined

**Step 3: Write minimal implementation**

```go
// render.go
package remarkable

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func SavePageToFile(dir string, pageID string, data []byte) (string, error) {
	path := filepath.Join(dir, pageID+".rm")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", path, err)
	}
	return path, nil
}

func BuildRmcCommand(rmPath string, svgPath string) *exec.Cmd {
	return exec.Command("rmc", "-o", svgPath, rmPath)
}

func BuildCairoSVGCommand(svgPath string, pngPath string) *exec.Cmd {
	return exec.Command("python3", "-c",
		fmt.Sprintf("import cairosvg; cairosvg.svg2png(url='%s', write_to='%s')", svgPath, pngPath))
}

func RenderPageToPNG(dir string, pageID string, rmData []byte) (string, error) {
	rmPath, err := SavePageToFile(dir, pageID, rmData)
	if err != nil {
		return "", err
	}

	svgPath := filepath.Join(dir, pageID+".svg")
	cmd := BuildRmcCommand(rmPath, svgPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("rmc failed: %w\n%s", err, out)
	}

	pngPath := filepath.Join(dir, pageID+".png")
	cmd = BuildCairoSVGCommand(svgPath, pngPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("cairosvg failed: %w\n%s", err, out)
	}

	return pngPath, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run "TestBuild|TestSavePage" ./internal/adapter/remarkable/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adapter/remarkable/render.go internal/adapter/remarkable/render_test.go
git commit -m "feat(remarkable): add .rm to PNG rendering via rmc and cairosvg"
```

---

### Task 4: Build Swift Vision OCR CLI tool

Create a small Swift CLI that takes a PNG path, runs `VNRecognizeTextRequest`, and outputs JSON with text and bounding boxes.

**Files:**
- Create: `tools/remarkable-ocr/main.swift`

**Step 1: Write the Swift OCR tool**

```swift
// tools/remarkable-ocr/main.swift
import Foundation
import Vision
import AppKit

struct OCRResult: Codable {
    let text: String
    let x: Double
    let y: Double
    let width: Double
    let height: Double
    let confidence: Float
}

guard CommandLine.arguments.count > 1 else {
    fputs("Usage: remarkable-ocr <image-path>\n", stderr)
    exit(1)
}

let imagePath = CommandLine.arguments[1]
guard let image = NSImage(contentsOfFile: imagePath),
      let cgImage = image.cgImage(forProposedRect: nil, context: nil, hints: nil) else {
    fputs("Error: could not load image at \(imagePath)\n", stderr)
    exit(1)
}

let request = VNRecognizeTextRequest()
request.recognitionLevel = .accurate
request.usesLanguageCorrection = true

let handler = VNImageRequestHandler(cgImage: cgImage, options: [:])
try handler.perform([request])

guard let observations = request.results else {
    print("[]")
    exit(0)
}

let imageHeight = Double(cgImage.height)
let imageWidth = Double(cgImage.width)

var results: [OCRResult] = []
for observation in observations {
    guard let candidate = observation.topCandidates(1).first else { continue }
    let box = observation.boundingBox
    results.append(OCRResult(
        text: candidate.string,
        x: box.origin.x * imageWidth,
        y: (1 - box.origin.y - box.height) * imageHeight,
        width: box.width * imageWidth,
        height: box.height * imageHeight,
        confidence: candidate.confidence
    ))
}

let encoder = JSONEncoder()
encoder.outputFormatting = .prettyPrinted
let data = try encoder.encode(results)
print(String(data: data, encoding: .utf8)!)
```

**Step 2: Compile and test manually**

```bash
mkdir -p tools/remarkable-ocr
# After writing the file:
swiftc -o tools/remarkable-ocr/remarkable-ocr tools/remarkable-ocr/main.swift -framework Vision -framework AppKit
# Verify it compiles
./tools/remarkable-ocr/remarkable-ocr
# Expected: "Usage: remarkable-ocr <image-path>" on stderr, exit 1
```

**Step 3: Commit**

```bash
git add tools/remarkable-ocr/main.swift
git commit -m "feat(remarkable): add Swift Vision OCR CLI tool"
```

---

### Task 5: Go wrapper for Vision OCR tool

Create a Go function that calls the Swift OCR tool and parses its JSON output.

**Files:**
- Create: `internal/adapter/remarkable/ocr.go`
- Test: `internal/adapter/remarkable/ocr_test.go`

**Step 1: Write the failing test**

```go
// ocr_test.go
package remarkable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOCRResults(t *testing.T) {
	jsonData := `[
		{"text": ". buy milk", "x": 50.0, "y": 100.0, "width": 200.0, "height": 30.0, "confidence": 0.95},
		{"text": "- meeting notes", "x": 50.0, "y": 140.0, "width": 250.0, "height": 30.0, "confidence": 0.92},
		{"text": ". sub task", "x": 100.0, "y": 180.0, "width": 180.0, "height": 30.0, "confidence": 0.88}
	]`

	results, err := ParseOCRResults([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, ". buy milk", results[0].Text)
	assert.InDelta(t, 50.0, results[0].X, 0.01)
	assert.InDelta(t, 100.0, results[0].Y, 0.01)
}

func TestParseOCRResultsEmpty(t *testing.T) {
	results, err := ParseOCRResults([]byte("[]"))
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestParseOCRResultsInvalidJSON(t *testing.T) {
	_, err := ParseOCRResults([]byte("not json"))
	assert.Error(t, err)
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestParseOCR ./internal/adapter/remarkable/...`
Expected: FAIL — `ParseOCRResults` not defined

**Step 3: Write minimal implementation**

```go
// ocr.go
package remarkable

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type OCRResult struct {
	Text       string  `json:"text"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
	Confidence float32 `json:"confidence"`
}

func ParseOCRResults(data []byte) ([]OCRResult, error) {
	var results []OCRResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func RunOCR(ocrToolPath string, pngPath string) ([]OCRResult, error) {
	cmd := exec.Command(ocrToolPath, pngPath)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("OCR failed: %w", err)
	}
	return ParseOCRResults(out)
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestParseOCR ./internal/adapter/remarkable/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adapter/remarkable/ocr.go internal/adapter/remarkable/ocr_test.go
git commit -m "feat(remarkable): add OCR result parsing and runner"
```

---

### Task 6: Reconstruct indented text from OCR bounding boxes

The core algorithm: take OCR results with bounding box positions, group into lines by y-coordinate, sort top-to-bottom, and determine indentation level from x-coordinate. The minimum x across all results is the left margin. Indentation depth = `(x - minX) / indentWidth`, where `indentWidth` is calibrated from the data.

**Files:**
- Create: `internal/adapter/remarkable/reconstruct.go`
- Test: `internal/adapter/remarkable/reconstruct_test.go`

**Step 1: Write the failing test**

```go
// reconstruct_test.go
package remarkable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReconstructText(t *testing.T) {
	results := []OCRResult{
		{Text: ". buy milk", X: 50, Y: 100, Width: 200, Height: 30},
		{Text: "- meeting notes", X: 50, Y: 140, Width: 250, Height: 30},
		{Text: ". sub task", X: 100, Y: 180, Width: 180, Height: 30},
		{Text: ". deep task", X: 150, Y: 220, Width: 180, Height: 30},
	}

	text := ReconstructText(results)
	assert.Equal(t, ". buy milk\n- meeting notes\n  . sub task\n    . deep task", text)
}

func TestReconstructTextSingleLine(t *testing.T) {
	results := []OCRResult{
		{Text: ". only task", X: 50, Y: 100, Width: 200, Height: 30},
	}

	text := ReconstructText(results)
	assert.Equal(t, ". only task", text)
}

func TestReconstructTextEmpty(t *testing.T) {
	text := ReconstructText(nil)
	assert.Equal(t, "", text)
}

func TestReconstructTextUnordered(t *testing.T) {
	results := []OCRResult{
		{Text: "- second line", X: 50, Y: 200, Width: 200, Height: 30},
		{Text: ". first line", X: 50, Y: 100, Width: 200, Height: 30},
	}

	text := ReconstructText(results)
	assert.Equal(t, ". first line\n- second line", text)
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestReconstructText ./internal/adapter/remarkable/...`
Expected: FAIL — `ReconstructText` not defined

**Step 3: Write minimal implementation**

```go
// reconstruct.go
package remarkable

import (
	"math"
	"sort"
	"strings"
)

const defaultIndentWidth = 50.0

func ReconstructText(results []OCRResult) string {
	if len(results) == 0 {
		return ""
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Y < results[j].Y
	})

	minX := math.MaxFloat64
	for _, r := range results {
		if r.X < minX {
			minX = r.X
		}
	}

	var lines []string
	for _, r := range results {
		depth := int(math.Round((r.X - minX) / defaultIndentWidth))
		indent := strings.Repeat("  ", depth)
		lines = append(lines, indent+r.Text)
	}

	return strings.Join(lines, "\n")
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestReconstructText ./internal/adapter/remarkable/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adapter/remarkable/reconstruct.go internal/adapter/remarkable/reconstruct_test.go
git commit -m "feat(remarkable): reconstruct indented text from OCR bounding boxes"
```

---

### Task 7: CLI command — `remarkable render`

Download notebook pages and render to PNGs in a directory. Prerequisite: `pip install rmc cairosvg`.

**Files:**
- Modify: `cmd/bujo/cmd/remarkable.go`

**Step 1: Add the render command**

```go
var remarkableRenderCmd = &cobra.Command{
	Use:   "render <doc-id>",
	Short: "Download notebook pages and render to PNG",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docID := args[0]
		outDir, _ := cmd.Flags().GetString("out-dir")

		configPath, err := remarkable.DefaultConfigPath()
		if err != nil {
			return err
		}
		cfg, err := remarkable.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("not registered — run 'bujo remarkable register <code>' first: %w", err)
		}

		client := remarkable.NewClient(remarkable.DefaultAuthHost)
		client.SetSyncHost(remarkable.DefaultSyncHost)

		fmt.Printf("Downloading pages for %s...\n", docID)
		pages, err := client.DownloadPages(cfg.DeviceToken, docID)
		if err != nil {
			return fmt.Errorf("failed to download pages: %w", err)
		}
		fmt.Printf("Downloaded %d pages\n", len(pages))

		if outDir == "" {
			outDir, err = os.MkdirTemp("", "remarkable-render-*")
			if err != nil {
				return err
			}
		} else {
			if err := os.MkdirAll(outDir, 0755); err != nil {
				return err
			}
		}

		for i, page := range pages {
			fmt.Printf("Rendering page %d/%d (%s)...\n", i+1, len(pages), page.PageID)
			pngPath, err := remarkable.RenderPageToPNG(outDir, page.PageID, page.Data)
			if err != nil {
				return fmt.Errorf("failed to render page %s: %w", page.PageID, err)
			}
			fmt.Printf("  → %s\n", pngPath)
		}

		fmt.Printf("\nPNGs saved to: %s\n", outDir)
		return nil
	},
}
```

Add `"os"` to imports. In `init()`:
```go
remarkableRenderCmd.Flags().String("out-dir", "", "Output directory for PNGs (default: temp dir)")
remarkableCmd.AddCommand(remarkableRenderCmd)
```

**Step 2: Build and verify compiles**

Run: `go build ./cmd/bujo/...`
Expected: No errors

**Step 3: Commit**

```bash
git add cmd/bujo/cmd/remarkable.go
git commit -m "feat(remarkable): add render command for .rm to PNG conversion"
```

---

### Task 8: CLI command — `remarkable ocr`

Run Apple Vision OCR on a PNG file or directory of PNGs. Outputs JSON with text and bounding boxes.

**Files:**
- Modify: `cmd/bujo/cmd/remarkable.go`

**Step 1: Add the ocr command**

```go
var remarkableOcrCmd = &cobra.Command{
	Use:   "ocr <png-path-or-dir>",
	Short: "Run Apple Vision OCR on PNG(s), output text with bounding boxes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

		ocrTool := filepath.Join(getToolsDir(), "remarkable-ocr", "remarkable-ocr")
		if _, err := os.Stat(ocrTool); os.IsNotExist(err) {
			return fmt.Errorf("OCR tool not found at %s — build with: swiftc -o %s tools/remarkable-ocr/main.swift -framework Vision -framework AppKit", ocrTool, ocrTool)
		}

		info, err := os.Stat(target)
		if err != nil {
			return fmt.Errorf("cannot access %s: %w", target, err)
		}

		var pngFiles []string
		if info.IsDir() {
			entries, err := os.ReadDir(target)
			if err != nil {
				return err
			}
			for _, e := range entries {
				if filepath.Ext(e.Name()) == ".png" {
					pngFiles = append(pngFiles, filepath.Join(target, e.Name()))
				}
			}
		} else {
			pngFiles = []string{target}
		}

		for i, png := range pngFiles {
			fmt.Fprintf(os.Stderr, "OCR page %d/%d: %s\n", i+1, len(pngFiles), filepath.Base(png))
			results, err := remarkable.RunOCR(ocrTool, png)
			if err != nil {
				return fmt.Errorf("OCR failed on %s: %w", png, err)
			}

			text := remarkable.ReconstructText(results)
			fmt.Printf("--- Page %d ---\n%s\n\n", i+1, text)
		}
		return nil
	},
}
```

Add helper function:
```go
func getToolsDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "tools"
	}
	return filepath.Join(filepath.Dir(exe), "..", "tools")
}
```

In `init()`:
```go
remarkableCmd.AddCommand(remarkableOcrCmd)
```

Add `"path/filepath"` to imports.

**Step 2: Build and verify compiles**

Run: `go build ./cmd/bujo/...`
Expected: No errors

**Step 3: Commit**

```bash
git add cmd/bujo/cmd/remarkable.go
git commit -m "feat(remarkable): add ocr command for Apple Vision text recognition"
```

---

### Task 9: CLI command — `remarkable import` (full pipeline)

Rewrite the existing `import` command to use the full pipeline: download pages → render → OCR → reconstruct → TreeParser.

**Files:**
- Modify: `cmd/bujo/cmd/remarkable.go`

**Step 1: Rewrite the import command**

```go
var remarkableImportCmd = &cobra.Command{
	Use:   "import <doc-id>",
	Short: "Download notebook, OCR pages, parse bujo entries, print to stdout",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docID := args[0]

		configPath, err := remarkable.DefaultConfigPath()
		if err != nil {
			return err
		}
		cfg, err := remarkable.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("not registered — run 'bujo remarkable register <code>' first: %w", err)
		}

		ocrTool := filepath.Join(getToolsDir(), "remarkable-ocr", "remarkable-ocr")
		if _, err := os.Stat(ocrTool); os.IsNotExist(err) {
			return fmt.Errorf("OCR tool not found — build with: swiftc -o %s tools/remarkable-ocr/main.swift -framework Vision -framework AppKit", ocrTool)
		}

		client := remarkable.NewClient(remarkable.DefaultAuthHost)
		client.SetSyncHost(remarkable.DefaultSyncHost)

		// Step 1: Download pages
		fmt.Fprintf(os.Stderr, "Downloading pages for %s...\n", docID)
		pages, err := client.DownloadPages(cfg.DeviceToken, docID)
		if err != nil {
			return fmt.Errorf("failed to download pages: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Downloaded %d pages\n", len(pages))

		tmpDir, err := os.MkdirTemp("", "remarkable-import-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)

		parser := domain.NewTreeParser()

		for i, page := range pages {
			// Step 2: Render to PNG
			fmt.Fprintf(os.Stderr, "Rendering page %d/%d...\n", i+1, len(pages))
			pngPath, err := remarkable.RenderPageToPNG(tmpDir, page.PageID, page.Data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: render failed for page %s: %v\n", page.PageID, err)
				continue
			}

			// Step 3: OCR
			fmt.Fprintf(os.Stderr, "OCR page %d/%d...\n", i+1, len(pages))
			results, err := remarkable.RunOCR(ocrTool, pngPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: OCR failed for page %s: %v\n", page.PageID, err)
				continue
			}

			// Step 4: Reconstruct indented text
			text := remarkable.ReconstructText(results)

			fmt.Printf("\n--- Page %d ---\n", i+1)
			fmt.Printf("Reconstructed text:\n%s\n", text)

			// Step 5: Parse with TreeParser
			entries, err := parser.Parse(text)
			if err != nil {
				fmt.Printf("Parse error: %v\n", err)
				continue
			}

			fmt.Printf("\nParsed %d entries:\n", len(entries))
			for _, e := range entries {
				indent := strings.Repeat("  ", e.Depth)
				fmt.Printf("%s%s %s", indent, e.Type, e.Content)
				if e.Priority != domain.PriorityNone {
					fmt.Printf(" [%s]", e.Priority)
				}
				if len(e.Tags) > 0 {
					fmt.Printf(" tags:%v", e.Tags)
				}
				fmt.Println()
			}
		}
		return nil
	},
}
```

**Step 2: Build and verify compiles**

Run: `go build ./cmd/bujo/...`
Expected: No errors

**Step 3: Commit**

```bash
git add cmd/bujo/cmd/remarkable.go
git commit -m "feat(remarkable): rewrite import to use render → OCR → parse pipeline"
```

---

### Task 10: Install dependencies and end-to-end test

Install Python dependencies, build Swift tool, and test the full pipeline.

**Step 1: Install Python dependencies**

```bash
pip install rmc cairosvg
```

**Step 2: Build Swift OCR tool**

```bash
swiftc -o tools/remarkable-ocr/remarkable-ocr tools/remarkable-ocr/main.swift -framework Vision -framework AppKit
```

**Step 3: Test render command**

```bash
./bujo --db-path :memory: remarkable render 4d8737b5-b17c-461f-9cb4-bc6c4ee88b62 --out-dir ./test-render
ls ./test-render/
```

**Step 4: Test OCR command**

```bash
./bujo --db-path :memory: remarkable ocr ./test-render/
```

**Step 5: Test full import**

```bash
./bujo --db-path :memory: remarkable import 4d8737b5-b17c-461f-9cb4-bc6c4ee88b62
```

**Step 6: Clean up**

```bash
rm -rf ./test-render
```
