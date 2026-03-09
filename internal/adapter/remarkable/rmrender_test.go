package remarkable

import (
	"bytes"
	"image/png"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectXOffset_CenteredCoordinates(t *testing.T) {
	// v6 native pages have centered coordinates: X ~ -702 to +702
	strokes := []rmStroke{
		{Points: []rmPoint{{X: -600, Y: 100}, {X: 700, Y: 100}}},
	}
	offset := detectXOffset(strokes)
	assert.Equal(t, float64(remarkableScreenWidth)/2, offset)
}

func TestDetectXOffset_AbsoluteCoordinates(t *testing.T) {
	// Migrated pages have absolute coordinates: X ~ 0 to 1404
	strokes := []rmStroke{
		{Points: []rmPoint{{X: 50, Y: 100}, {X: 1300, Y: 100}}},
	}
	offset := detectXOffset(strokes)
	assert.Equal(t, float64(0), offset)
}

func TestDetectXOffset_EmptyStrokes(t *testing.T) {
	offset := detectXOffset(nil)
	assert.Equal(t, float64(remarkableScreenWidth)/2, offset,
		"default to centered when no data")
}

func TestRenderStrokes_CenteredCoordinates(t *testing.T) {
	// Centered coordinates: stroke at X=-200 to X=200 should render
	// at screen pixels (502 to 902) * renderScale (after adding 702 offset)
	strokes := []rmStroke{
		{Points: []rmPoint{{X: -200, Y: 100}, {X: 200, Y: 100}}},
	}

	data, err := RenderStrokes(strokes)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	halfW := remarkableScreenWidth / 2 * renderScale
	// Middle of stroke (X=0 → pixel 702*scale) should be dark
	r, g, b, _ := img.At(halfW, 100*renderScale).RGBA()
	assert.Less(t, r, uint32(0x8000), "center of stroke should be dark")
	assert.Less(t, g, uint32(0x8000))
	assert.Less(t, b, uint32(0x8000))

	// Far left (pixel 100) should be white — stroke starts at pixel 502*scale
	r2, g2, b2, _ := img.At(100, 100*renderScale).RGBA()
	assert.Equal(t, uint32(0xFFFF), r2)
	assert.Equal(t, uint32(0xFFFF), g2)
	assert.Equal(t, uint32(0xFFFF), b2)
}

func TestRenderStrokes_Dimensions(t *testing.T) {
	strokes := []rmStroke{
		{Points: []rmPoint{{X: -100, Y: 100}, {X: 200, Y: 200}}},
	}

	data, err := RenderStrokes(strokes)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	bounds := img.Bounds()
	assert.Equal(t, remarkableScreenWidth*renderScale, bounds.Max.X)
	assert.Equal(t, remarkableScreenHeight*renderScale, bounds.Max.Y)
}

func TestRenderStrokes_EmptyStrokes(t *testing.T) {
	data, err := RenderStrokes(nil)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	bounds := img.Bounds()
	assert.Equal(t, remarkableScreenWidth*renderScale, bounds.Max.X)
}

func TestRenderStrokes_WhiteBackground(t *testing.T) {
	strokes := []rmStroke{}

	data, err := RenderStrokes(strokes)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	r, g, b, a := img.At(0, 0).RGBA()
	assert.Equal(t, uint32(0xFFFF), r)
	assert.Equal(t, uint32(0xFFFF), g)
	assert.Equal(t, uint32(0xFFFF), b)
	assert.Equal(t, uint32(0xFFFF), a)
}

func TestDetectYOffset_NegativeCoordinates(t *testing.T) {
	strokes := []rmStroke{
		{Points: []rmPoint{{X: 100, Y: -210}, {X: 300, Y: 500}}},
	}
	offset := detectYOffset(strokes)
	assert.Equal(t, float64(210), offset)
}

func TestDetectYOffset_PositiveCoordinates(t *testing.T) {
	strokes := []rmStroke{
		{Points: []rmPoint{{X: 100, Y: 50}, {X: 300, Y: 500}}},
	}
	offset := detectYOffset(strokes)
	assert.Equal(t, float64(0), offset)
}

func TestDetectYOffset_EmptyStrokes(t *testing.T) {
	offset := detectYOffset(nil)
	assert.Equal(t, float64(0), offset)
}

func TestRenderStrokes_NegativeYCoordinates(t *testing.T) {
	// Quick Sheets notebook has Y range from -210 to 1529
	// Strokes with negative Y should still be visible in the rendered image
	// Use wide X range so detectXOffset returns 0 (absolute coordinates)
	strokes := []rmStroke{
		{Points: []rmPoint{{X: 100, Y: -100}, {X: 1300, Y: -100}}},
		{Points: []rmPoint{{X: 100, Y: 500}, {X: 1300, Y: 500}}},
	}

	data, err := RenderStrokes(strokes)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	// Y=-100 with yOffset=100 → renders at pixel Y=0, scaled
	// Check that there are dark pixels in the top portion
	foundDark := false
	for y := 0; y < 10*renderScale; y++ {
		for x := 100 * renderScale; x < 1300*renderScale; x++ {
			r, _, _, _ := img.At(x, y).RGBA()
			if r < 0x8000 {
				foundDark = true
				break
			}
		}
		if foundDark {
			break
		}
	}
	assert.True(t, foundDark, "stroke at negative Y should be visible in rendered image")
}

func TestComputeCanvasSize_StandardBounds(t *testing.T) {
	// Strokes within standard bounds should produce standard canvas
	strokes := []rmStroke{
		{Points: []rmPoint{{X: 100, Y: 100}, {X: 1300, Y: 1800}}},
	}
	w, h := computeCanvasSize(strokes, 0, 0)
	assert.Equal(t, remarkableScreenWidth, w)
	assert.Equal(t, remarkableScreenHeight, h)
}

func TestComputeCanvasSize_ExceedsWidth(t *testing.T) {
	// Strokes extending beyond right edge should expand canvas width
	strokes := []rmStroke{
		{Points: []rmPoint{{X: 100, Y: 100}, {X: 1800, Y: 100}}},
	}
	w, h := computeCanvasSize(strokes, 0, 0)
	assert.Equal(t, 1801, w, "canvas should expand to fit rightmost stroke point")
	assert.Equal(t, remarkableScreenHeight, h, "height should stay standard")
}

func TestComputeCanvasSize_ExceedsHeight(t *testing.T) {
	// Strokes extending below standard height should expand canvas
	strokes := []rmStroke{
		{Points: []rmPoint{{X: 100, Y: 100}, {X: 100, Y: 2200}}},
	}
	w, h := computeCanvasSize(strokes, 0, 0)
	assert.Equal(t, remarkableScreenWidth, w)
	assert.Equal(t, 2201, h, "canvas should expand to fit lowest stroke point")
}

func TestComputeCanvasSize_WithOffsets(t *testing.T) {
	// Offsets shift strokes — canvas must account for shifted positions
	strokes := []rmStroke{
		{Points: []rmPoint{{X: -200, Y: -100}, {X: 1200, Y: 1800}}},
	}
	// xOffset=702 (centered), yOffset=100 (negative Y)
	// maxX after offset: 1200 + 702 = 1902
	// maxY after offset: 1800 + 100 = 1900
	w, h := computeCanvasSize(strokes, 702, 100)
	assert.Equal(t, 1903, w, "canvas should fit shifted rightmost point")
	assert.Equal(t, 1901, h, "canvas should fit shifted lowest point")
}

func TestComputeCanvasSize_EmptyStrokes(t *testing.T) {
	w, h := computeCanvasSize(nil, 0, 0)
	assert.Equal(t, remarkableScreenWidth, w)
	assert.Equal(t, remarkableScreenHeight, h)
}

func TestRenderStrokes_ExpandsCanvasForWideContent(t *testing.T) {
	// Quick Sheets: text extends beyond 1404px width
	strokes := []rmStroke{
		{Points: []rmPoint{{X: 100, Y: 100}, {X: 1800, Y: 100}}},
	}

	data, err := RenderStrokes(strokes)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	bounds := img.Bounds()
	assert.Greater(t, bounds.Max.X, remarkableScreenWidth*renderScale, "canvas should be wider than standard")

	// Stroke should have dark pixels beyond standard width (not clipped)
	foundDark := false
	for x := remarkableScreenWidth * renderScale; x <= 1800*renderScale; x++ {
		r, _, _, _ := img.At(x, 100*renderScale).RGBA()
		if r < 0x8000 {
			foundDark = true
			break
		}
	}
	assert.True(t, foundDark, "stroke beyond standard width should be visible")
}

func TestRenderStrokes_AbsoluteCoordinates(t *testing.T) {
	// Migrated pages have absolute coordinates spanning ~0 to ~1404
	// Center of X range (~700) is close to 702 → no offset applied
	strokes := []rmStroke{
		{Points: []rmPoint{{X: 100, Y: 100}, {X: 1300, Y: 100}}},
	}

	data, err := RenderStrokes(strokes)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	// Middle of stroke (X=700, Y=100) should have dark pixels (no offset), scaled
	r, g, b, _ := img.At(700*renderScale, 100*renderScale).RGBA()
	assert.Less(t, r, uint32(0x8000), "stroke pixel should be dark")
	assert.Less(t, g, uint32(0x8000))
	assert.Less(t, b, uint32(0x8000))

	// Pixel 50 should be white — stroke starts at 100*scale, no offset
	r2, g2, b2, _ := img.At(50, 100*renderScale).RGBA()
	assert.Equal(t, uint32(0xFFFF), r2)
	assert.Equal(t, uint32(0xFFFF), g2)
	assert.Equal(t, uint32(0xFFFF), b2)
}
