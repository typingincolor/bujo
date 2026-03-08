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
	// at screen pixels 502 to 902 (after adding 702 offset)
	strokes := []rmStroke{
		{Points: []rmPoint{{X: -200, Y: 100}, {X: 200, Y: 100}}},
	}

	data, err := RenderStrokes(strokes)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	halfW := remarkableScreenWidth / 2 // 702
	// Middle of stroke (X=0 → pixel 702) should be dark
	r, g, b, _ := img.At(halfW, 100).RGBA()
	assert.Less(t, r, uint32(0x8000), "center of stroke should be dark")
	assert.Less(t, g, uint32(0x8000))
	assert.Less(t, b, uint32(0x8000))

	// Far left (pixel 100) should be white — stroke starts at pixel 502
	r2, g2, b2, _ := img.At(100, 100).RGBA()
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

	r, g, b, a := img.At(0, 0).RGBA()
	assert.Equal(t, uint32(0xFFFF), r)
	assert.Equal(t, uint32(0xFFFF), g)
	assert.Equal(t, uint32(0xFFFF), b)
	assert.Equal(t, uint32(0xFFFF), a)
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

	// Middle of stroke (X=700, Y=100) should have dark pixels (no offset)
	r, g, b, _ := img.At(700, 100).RGBA()
	assert.Less(t, r, uint32(0x8000), "stroke pixel should be dark")
	assert.Less(t, g, uint32(0x8000))
	assert.Less(t, b, uint32(0x8000))

	// Pixel 50 should be white — stroke starts at 100, no offset
	r2, g2, b2, _ := img.At(50, 100).RGBA()
	assert.Equal(t, uint32(0xFFFF), r2)
	assert.Equal(t, uint32(0xFFFF), g2)
	assert.Equal(t, uint32(0xFFFF), b2)
}
