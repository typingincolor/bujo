package wails

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/typingincolor/bujo/internal/adapter/remarkable"
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
	caps := buildPlatformCapabilities()
	assert.Equal(t, runtime.GOOS, caps.Platform)
}

func TestPlatformCapabilities_HasOCR_OnDarwin(t *testing.T) {
	caps := buildPlatformCapabilities()
	if runtime.GOOS == "darwin" {
		assert.True(t, caps.HasOCR)
	} else {
		assert.False(t, caps.HasOCR)
	}
}

func TestListRemarkableDocuments_NoConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	app := &App{}
	_, err := app.ListRemarkableDocuments()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
}

func TestImportResult_JSONSerialization(t *testing.T) {
	result := ImportRemarkableResult{
		Pages: []ImportedPage{
			{
				PageID: "page-1",
				PNG:    "base64data",
				OCRResults: []remarkable.OCRResult{
					{Text: "hello", X: 10, Y: 20, Width: 100, Height: 30, Confidence: 0.95},
				},
				Text:               "hello",
				LowConfidenceCount: 0,
			},
		},
	}
	assert.Equal(t, 1, len(result.Pages))
	assert.Equal(t, "page-1", result.Pages[0].PageID)
	assert.Equal(t, float32(0.95), result.Pages[0].OCRResults[0].Confidence)
	assert.Equal(t, "hello", result.Pages[0].Text)
	assert.Equal(t, 0, result.Pages[0].LowConfidenceCount)
}

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

func TestRegisterRemarkableDevice_EmptyCode(t *testing.T) {
	app := &App{}
	err := app.RegisterRemarkableDevice("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code")
}
