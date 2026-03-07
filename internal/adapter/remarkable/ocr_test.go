package remarkable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOCRResults(t *testing.T) {
	jsonData := `[
		{"text": ". buy milk", "x": 50.0, "y": 100.0, "width": 200.0, "height": 30.0, "confidence": 0.95},
		{"text": "- meeting notes", "x": 50.0, "y": 140.0, "width": 250.0, "height": 30.0, "confidence": 0.92},
		{"text": ". sub task", "x": 100.0, "y": 180.0, "width": 180.0, "height": 30.0, "confidence": 0.88}
	]`

	results, err := ParseOCRResults([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, ". buy milk", results[0].Text)
	assert.InDelta(t, 50.0, results[0].X, 0.01)
	assert.InDelta(t, 100.0, results[0].Y, 0.01)
}

func TestParseOCRResultsEmpty(t *testing.T) {
	results, err := ParseOCRResults([]byte("[]"))
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestParseOCRResultsInvalidJSON(t *testing.T) {
	_, err := ParseOCRResults([]byte("not json"))
	assert.Error(t, err)
}
