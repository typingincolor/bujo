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
	assert.Equal(t, "rmc", cmd.Args[0])
	assert.Contains(t, cmd.Args, "-o")
	assert.Contains(t, cmd.Args, "/tmp/page.svg")
	assert.Contains(t, cmd.Args, "/tmp/page.rm")
}

func TestBuildCairoSVGCommand(t *testing.T) {
	cmd := BuildCairoSVGCommand("/tmp/page.svg", "/tmp/page.png")
	assert.Equal(t, "python3", cmd.Args[0])
	assert.Contains(t, cmd.Args, "-c")

	script := cmd.Args[2]
	assert.Contains(t, script, "cairosvg")
	assert.Contains(t, script, "/tmp/page.svg")
	assert.Contains(t, script, "/tmp/page.png")
	assert.Contains(t, script, "output_width=1404")
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
