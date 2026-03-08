package remarkable

import (
	"bytes"
	"image/png"
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

	r, g, b, a := img.At(0, 0).RGBA()
	assert.Equal(t, uint32(0xFFFF), r)
	assert.Equal(t, uint32(0xFFFF), g)
	assert.Equal(t, uint32(0xFFFF), b)
	assert.Equal(t, uint32(0xFFFF), a)
}

func TestRenderStrokes_CenteredCoordinates(t *testing.T) {
	// reMarkable coordinates have X centered at 0 (range ~ -702 to +702)
	// A horizontal stroke at X=-100 to X=100, Y=100 should render
	// at screen pixels X=602 to X=802, Y=100
	strokes := []rmStroke{
		{Points: []rmPoint{{X: -100, Y: 100}, {X: 100, Y: 100}}},
	}

	data, err := RenderStrokes(strokes)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	// Center of stroke should have dark pixels (screen X=702, Y=100)
	r, g, b, _ := img.At(remarkableScreenWidth/2, 100).RGBA()
	assert.Less(t, r, uint32(0x8000), "center pixel should be dark")
	assert.Less(t, g, uint32(0x8000))
	assert.Less(t, b, uint32(0x8000))

	// Far left (X=0) should be white — the stroke doesn't reach there
	r2, g2, b2, _ := img.At(0, 100).RGBA()
	assert.Equal(t, uint32(0xFFFF), r2)
	assert.Equal(t, uint32(0xFFFF), g2)
	assert.Equal(t, uint32(0xFFFF), b2)
}
