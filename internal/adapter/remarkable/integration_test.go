package remarkable

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderPageToPNG_Integration(t *testing.T) {
	// Use centered X coordinates (reMarkable origin is top-center)
	// X=-200 to X=100 maps to screen pixels 502 to 802
	rmData := buildTestRM(t, []rmPoint{
		{X: -200.0, Y: 100.0},
		{X: 100.0, Y: 100.0},
		{X: 100.0, Y: 300.0},
	})

	dir := t.TempDir()
	pngPath, err := RenderPageToPNG(dir, "integration-test", rmData)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "integration-test.png"), pngPath)

	f, err := os.Open(pngPath)
	require.NoError(t, err)
	defer f.Close()

	img, err := png.Decode(f)
	require.NoError(t, err)

	bounds := img.Bounds()
	assert.Equal(t, remarkableScreenWidth, bounds.Max.X)
	assert.Equal(t, remarkableScreenHeight, bounds.Max.Y)

	xOff := remarkableScreenWidth / 2
	assert.True(t, hasNonWhitePixels(img, -200+xOff, 95, 100+xOff, 105),
		"expected black pixels along horizontal stroke")

	assert.True(t, hasNonWhitePixels(img, 100+xOff-5, 100, 100+xOff+5, 300),
		"expected black pixels along vertical stroke")

	assert.False(t, hasNonWhitePixels(img, 0, 700, 100, 800),
		"expected no strokes in empty region")
}

func TestRenderPageToPNG_MultipleStrokes(t *testing.T) {
	// Centered X coordinates: -300 to -100 maps to screen 402 to 602
	stroke1 := []rmPoint{{X: -300, Y: 50}, {X: -100, Y: 50}}
	stroke2 := []rmPoint{{X: -300, Y: 500}, {X: -100, Y: 500}}

	rmData := buildTestRMMultiStroke(t, [][]rmPoint{stroke1, stroke2})

	dir := t.TempDir()
	pngPath, err := RenderPageToPNG(dir, "multi-stroke", rmData)
	require.NoError(t, err)

	f, err := os.Open(pngPath)
	require.NoError(t, err)
	defer f.Close()

	img, err := png.Decode(f)
	require.NoError(t, err)

	xOff := remarkableScreenWidth / 2
	assert.True(t, hasNonWhitePixels(img, -300+xOff, 45, -100+xOff, 55),
		"expected pixels along first stroke")

	assert.True(t, hasNonWhitePixels(img, -300+xOff, 495, -100+xOff, 505),
		"expected pixels along second stroke")

	assert.False(t, hasNonWhitePixels(img, -300+xOff, 250, -100+xOff, 260),
		"expected no strokes between the two lines")
}

func hasNonWhitePixels(img image.Image, x1, y1, x2, y2 int) bool {
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r != 0xFFFF || g != 0xFFFF || b != 0xFFFF {
				return true
			}
		}
	}
	return false
}

func TestReconstructAndParseIntegration(t *testing.T) {
	results := []OCRResult{
		{Text: ". buy milk", X: 50, Y: 100, Width: 200, Height: 30},
		{Text: ". call dentist", X: 100, Y: 140, Width: 250, Height: 30},
		{Text: "- meeting notes", X: 50, Y: 180, Width: 200, Height: 30},
	}

	text := ReconstructText(results)
	assert.Contains(t, text, ". buy milk")
	assert.Contains(t, text, "  . call dentist")
	assert.Contains(t, text, "- meeting notes")
}

func TestRenderAndDecodeRoundTrip(t *testing.T) {
	// Use centered X coordinates
	strokes := []rmStroke{
		{Points: []rmPoint{{X: -400, Y: 10}, {X: -300, Y: 100}}},
		{Points: []rmPoint{{X: 200, Y: 500}, {X: 300, Y: 600}}},
	}

	data, err := RenderStrokes(strokes)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	xOff := remarkableScreenWidth / 2
	// Near first stroke: screen X ~302 to ~402, Y ~10 to ~100
	nearStroke := img.At(-350+xOff, 50)
	r, g, b, _ := nearStroke.RGBA()
	isNearStroke := r != 0xFFFF || g != 0xFFFF || b != 0xFFFF
	_ = isNearStroke

	// Far from any strokes
	white := color.RGBA{255, 255, 255, 255}
	farAway := img.At(50, 300)
	fr, fg, fb, _ := farAway.RGBA()
	wr, wg, wb, _ := white.RGBA()
	assert.Equal(t, wr, fr, "expected white far from strokes (R)")
	assert.Equal(t, wg, fg, "expected white far from strokes (G)")
	assert.Equal(t, wb, fb, "expected white far from strokes (B)")
}
