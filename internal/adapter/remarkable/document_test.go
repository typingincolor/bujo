package remarkable

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestZIP(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := w.Create(name)
		require.NoError(t, err)
		_, err = f.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	return buf.Bytes()
}

func TestExtractTextFromZIP(t *testing.T) {
	zipData := createTestZIP(t, map[string]string{
		"doc-id.content":         `{"fileType": "notebook"}`,
		"doc-id/0.rm":            "binary-stroke-data",
		"doc-id/0-metadata.json": `{"layers": [{"name": "Layer 1"}]}`,
		"doc-id/0.txt":           ". Buy groceries\n- Remember to call dentist",
	})

	texts, err := ExtractTextFromZIP(zipData)
	require.NoError(t, err)
	require.Len(t, texts, 1)
	assert.Contains(t, texts[0], "Buy groceries")
}

func TestExtractTextFromZIPNoTextFiles(t *testing.T) {
	zipData := createTestZIP(t, map[string]string{
		"doc-id.content": `{"fileType": "notebook"}`,
		"doc-id/0.rm":    "binary-stroke-data",
	})

	texts, err := ExtractTextFromZIP(zipData)
	require.NoError(t, err)
	assert.Empty(t, texts)
}

func TestListZIPContents(t *testing.T) {
	zipData := createTestZIP(t, map[string]string{
		"doc-id.content": `{"fileType": "notebook"}`,
		"doc-id/0.rm":    "binary-data",
	})

	names, err := ListZIPContents(zipData)
	require.NoError(t, err)
	assert.Len(t, names, 2)
}
