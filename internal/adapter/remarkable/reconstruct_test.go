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
