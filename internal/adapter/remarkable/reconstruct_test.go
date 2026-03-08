package remarkable

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestReconstructTextWithConfidence_DepthNormalization(t *testing.T) {
	results := []OCRResult{
		{Text: ". root", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 0.95},
		{Text: ". skipped deep", X: 200, Y: 200, Width: 200, Height: 30, Confidence: 0.90},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, ". root\n  . skipped deep", result.Text)
	assert.Equal(t, 0, result.LowConfidenceCount)
}

func TestReconstructTextWithConfidence_CountsLowConfidence(t *testing.T) {
	results := []OCRResult{
		{Text: ". clear", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 0.95},
		{Text: ". fuzzy", X: 50, Y: 200, Width: 200, Height: 30, Confidence: 0.5},
		{Text: ". also fuzzy", X: 50, Y: 300, Width: 200, Height: 30, Confidence: 0.7},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, 2, result.LowConfidenceCount)
}

func TestReconstructText_PrependsNoteSymbolToUnprefixedLines(t *testing.T) {
	results := []OCRResult{
		{Text: "Responsible for observability", X: 50, Y: 100, Width: 300, Height: 30},
		{Text: ". buy milk", X: 50, Y: 140, Width: 200, Height: 30},
		{Text: "streams platform", X: 50, Y: 180, Width: 250, Height: 30},
		{Text: "= US", X: 100, Y: 220, Width: 100, Height: 30},
	}

	text := ReconstructText(results)
	assert.Equal(t, "- Responsible for observability\n. buy milk\n- streams platform\n  - = US", text)
}

func TestReconstructText_WordsStartingWithSymbolChars(t *testing.T) {
	results := []OCRResult{
		{Text: "adoption is Key", X: 50, Y: 100, Width: 300, Height: 30},
		{Text: "over budget", X: 50, Y: 140, Width: 200, Height: 30},
		{Text: "x-ray results", X: 50, Y: 180, Width: 250, Height: 30},
	}

	text := ReconstructText(results)
	assert.Equal(t, "- adoption is Key\n- over budget\n- x-ray results", text)
}

func TestReconstructTextWithConfidence_ReportsLowConfidenceLines(t *testing.T) {
	results := []OCRResult{
		{Text: ". clear", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 0.95},
		{Text: ". fuzzy", X: 50, Y: 200, Width: 200, Height: 30, Confidence: 0.5},
		{Text: ". also clear", X: 50, Y: 300, Width: 200, Height: 30, Confidence: 0.9},
		{Text: ". also fuzzy", X: 50, Y: 400, Width: 200, Height: 30, Confidence: 0.7},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, []int{1, 3}, result.LowConfidenceLines)
	assert.Equal(t, 2, result.LowConfidenceCount)
}

func TestReconstructTextWithConfidence_DepthResetsOnRoot(t *testing.T) {
	results := []OCRResult{
		{Text: ". a", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 0.9},
		{Text: ". b", X: 100, Y: 200, Width: 200, Height: 30, Confidence: 0.9},
		{Text: ". c", X: 50, Y: 300, Width: 200, Height: 30, Confidence: 0.9},
		{Text: ". d", X: 200, Y: 400, Width: 200, Height: 30, Confidence: 0.9},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, ". a\n  . b\n. c\n  . d", result.Text)
}
