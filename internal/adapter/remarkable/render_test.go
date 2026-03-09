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
	rmData := []byte("reMarkable .lines file, version=6          ")

	path, err := RenderPageToPNG(dir, "test-page", rmData)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "test-page.png"), path)

	f, err := os.Open(path)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	img, err := png.Decode(f)
	require.NoError(t, err)
	bounds := img.Bounds()
	assert.Equal(t, remarkableScreenWidth*renderScale, bounds.Max.X)
	assert.Equal(t, remarkableScreenHeight*renderScale, bounds.Max.Y)
}

func TestRenderPageToPNG_InvalidRM(t *testing.T) {
	dir := t.TempDir()
	_, err := RenderPageToPNG(dir, "bad-page", []byte("not a valid rm file"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid .rm header")
}
