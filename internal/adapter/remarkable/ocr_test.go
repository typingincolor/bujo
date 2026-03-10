package remarkable

import (
	"os"
	"strings"
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

func TestParseOCRResults_WithCandidates(t *testing.T) {
	jsonData := `[
		{"text": "Fest task", "x": 50.0, "y": 100.0, "width": 200.0, "height": 30.0, "confidence": 0.85,
		 "candidates": [
			{"text": "Fest task", "confidence": 0.85},
			{"text": "Test task", "confidence": 0.80},
			{"text": "Best task", "confidence": 0.75}
		 ]}
	]`

	results, err := ParseOCRResults([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Fest task", results[0].Text)
	require.Len(t, results[0].Candidates, 3)
	assert.Equal(t, "Test task", results[0].Candidates[1].Text)
	assert.InDelta(t, 0.80, float64(results[0].Candidates[1].Confidence), 0.01)
}

func TestOCRCustomWords_NotEmpty(t *testing.T) {
	words := OCRCustomWords()
	assert.NotEmpty(t, words)
}

func TestOCRCustomWords_ContainsBujoTerms(t *testing.T) {
	words := OCRCustomWords()
	wordSet := make(map[string]bool)
	for _, w := range words {
		wordSet[w] = true
	}
	assert.True(t, wordSet["Engagement"], "should contain 'Engagement'")
	assert.True(t, wordSet["Architecture"], "should contain 'Architecture'")
	assert.True(t, wordSet["Derek"], "should contain 'Derek'")
}

func TestOCRCustomWords_SkipsComments(t *testing.T) {
	words := OCRCustomWords()
	for _, w := range words {
		assert.False(t, strings.HasPrefix(w, "#"), "should not contain comment: %s", w)
	}
}

func TestWriteOCRCustomWordsFile_CreatesFile(t *testing.T) {
	path, cleanup, err := writeOCRCustomWordsFile()
	require.NoError(t, err)
	defer cleanup()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "Engagement")
	assert.Contains(t, content, "Architecture")
}

func TestAppleVisionOCR_ImplementsOCRProvider(t *testing.T) {
	var _ OCRProvider = &AppleVisionOCR{}
}

func TestParseOCRResults_WithoutCandidatesBackwardCompat(t *testing.T) {
	jsonData := `[{"text": ". task", "x": 50.0, "y": 100.0, "width": 200.0, "height": 30.0, "confidence": 0.9}]`

	results, err := ParseOCRResults([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Nil(t, results[0].Candidates)
}
