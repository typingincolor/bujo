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

func TestReconstructTextWithConfidence_DepthNormalization(t *testing.T) {
	results := []OCRResult{
		{Text: ". root", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 0.95},
		{Text: ". skipped deep", X: 200, Y: 200, Width: 200, Height: 30, Confidence: 0.90},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, ". root\n  . skipped deep", result.Text)
	assert.Equal(t, 0, result.LowConfidenceCount)
}

func TestReconstructTextWithConfidence_CountsLowConfidence(t *testing.T) {
	results := []OCRResult{
		{Text: ". clear", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 0.95},
		{Text: ". fuzzy", X: 50, Y: 200, Width: 200, Height: 30, Confidence: 0.5},
		{Text: ". also fuzzy", X: 50, Y: 300, Width: 200, Height: 30, Confidence: 0.7},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, 2, result.LowConfidenceCount)
}

func TestReconstructText_ConcatenatesUnprefixedLinesToAbove(t *testing.T) {
	results := []OCRResult{
		{Text: "Responsible for observability", X: 50, Y: 100, Width: 300, Height: 30},
		{Text: ". buy milk", X: 50, Y: 140, Width: 200, Height: 30},
		{Text: "streams platform", X: 50, Y: 180, Width: 250, Height: 30},
		{Text: "= US", X: 100, Y: 220, Width: 100, Height: 30},
	}

	text := ReconstructText(results)
	assert.Equal(t, "- Responsible for observability\n. buy milk streams platform = US", text)
}

func TestReconstructText_ConcatenatesAllUnprefixedLines(t *testing.T) {
	results := []OCRResult{
		{Text: "adoption is Key", X: 50, Y: 100, Width: 300, Height: 30},
		{Text: "over budget", X: 50, Y: 140, Width: 200, Height: 30},
		{Text: "x-ray results", X: 50, Y: 180, Width: 250, Height: 30},
	}

	text := ReconstructText(results)
	assert.Equal(t, "- adoption is Key over budget x-ray results", text)
}

func TestReconstructTextWithConfidence_ReportsLowConfidenceLines(t *testing.T) {
	results := []OCRResult{
		{Text: ". clear", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 0.95},
		{Text: ". fuzzy", X: 50, Y: 200, Width: 200, Height: 30, Confidence: 0.5},
		{Text: ". also clear", X: 50, Y: 300, Width: 200, Height: 30, Confidence: 0.9},
		{Text: ". also fuzzy", X: 50, Y: 400, Width: 200, Height: 30, Confidence: 0.7},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, []int{1, 3}, result.LowConfidenceLines)
	assert.Equal(t, 2, result.LowConfidenceCount)
}

func TestReconstructText_MergesFragmentsOnSameLine(t *testing.T) {
	results := []OCRResult{
		{Text: "- note", X: 50, Y: 200, Width: 120, Height: 30},
		{Text: "2", X: 180, Y: 205, Width: 30, Height: 30},
	}

	text := ReconstructText(results)
	assert.Equal(t, "- note 2", text)
}

func TestReconstructText_MergesMultipleFragmentsOnSameLine(t *testing.T) {
	results := []OCRResult{
		{Text: ". buy", X: 50, Y: 100, Width: 80, Height: 30},
		{Text: "milk", X: 140, Y: 103, Width: 70, Height: 30},
		{Text: "today", X: 220, Y: 98, Width: 80, Height: 30},
	}

	text := ReconstructText(results)
	assert.Equal(t, ". buy milk today", text)
}

func TestReconstructText_DoesNotMergeDistantLines(t *testing.T) {
	results := []OCRResult{
		{Text: "- first line", X: 50, Y: 100, Width: 200, Height: 30},
		{Text: "- second line", X: 50, Y: 200, Width: 200, Height: 30},
	}

	text := ReconstructText(results)
	assert.Equal(t, "- first line\n- second line", text)
}

func TestReconstructText_MergedLineUsesLowestConfidence(t *testing.T) {
	results := []OCRResult{
		{Text: "- note", X: 50, Y: 200, Width: 120, Height: 30, Confidence: 0.95},
		{Text: "2", X: 180, Y: 205, Width: 30, Height: 30, Confidence: 0.5},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, "- note 2", result.Text)
	assert.Equal(t, 1, result.LowConfidenceCount)
	assert.Equal(t, []int{0}, result.LowConfidenceLines)
}

func TestReconstructText_SelectsBestCandidateWithBujoPrefix(t *testing.T) {
	results := []OCRResult{
		{
			Text: "Fest task 1", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 0.85,
			Candidates: []OCRCandidate{
				{Text: "Fest task 1", Confidence: 0.85},
				{Text: ". Test task 1", Confidence: 0.80},
				{Text: "Best task 1", Confidence: 0.75},
			},
		},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, ". Test task 1", result.Text)
}

func TestReconstructText_KeepsTopCandidateWhenNoBujoPrefixFound(t *testing.T) {
	results := []OCRResult{
		{
			Text: "some text", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 0.85,
			Candidates: []OCRCandidate{
				{Text: "some text", Confidence: 0.85},
				{Text: "same text", Confidence: 0.80},
			},
		},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, "- some text", result.Text)
}

func TestReconstructText_UncertainWhenAlternativeIsValidWord(t *testing.T) {
	results := []OCRResult{
		{
			Text: "• Fest task 1", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 1.0,
			Candidates: []OCRCandidate{
				{Text: "• Fest task 1", Confidence: 1.0},
				{Text: "• Test task 1", Confidence: 1.0},
				{Text: "• test task 1", Confidence: 1.0},
			},
		},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, []int{0}, result.UncertainLines)
}

func TestReconstructText_UncertainWhenNeitherWordValid(t *testing.T) {
	results := []OCRResult{
		{
			Text: "- moh stuff", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 1.0,
			Candidates: []OCRCandidate{
				{Text: "- moh stuff", Confidence: 1.0},
				{Text: "- mol stuff", Confidence: 1.0},
			},
		},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, []int{0}, result.UncertainLines)
}

func TestReconstructText_DetectsUncertaintyWhenBothWordsValid(t *testing.T) {
	results := []OCRResult{
		{
			Text: "- test note", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 1.0,
			Candidates: []OCRCandidate{
				{Text: "- test note", Confidence: 1.0},
				{Text: "- best note", Confidence: 1.0},
			},
		},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, []int{0}, result.UncertainLines)
}

func TestReconstructText_NoUncertaintyWhenCandidatesAgree(t *testing.T) {
	results := []OCRResult{
		{
			Text: "- Test note", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 1.0,
			Candidates: []OCRCandidate{
				{Text: "- Test note", Confidence: 1.0},
				{Text: "-Test note", Confidence: 1.0},
				{Text: "Test note", Confidence: 1.0},
			},
		},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Empty(t, result.UncertainLines)
}

func TestReconstructText_NoUncertaintyWithoutCandidates(t *testing.T) {
	results := []OCRResult{
		{Text: ". task", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 0.9},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Empty(t, result.UncertainLines)
}

func TestReconstructText_UncertaintyIgnoresCaseDifferences(t *testing.T) {
	results := []OCRResult{
		{
			Text: "• Task 2", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 1.0,
			Candidates: []OCRCandidate{
				{Text: "• Task 2", Confidence: 1.0},
				{Text: "• task 2", Confidence: 1.0},
				{Text: "o Task 2", Confidence: 1.0},
			},
		},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Empty(t, result.UncertainLines)
}

func TestReconstructText_UncertaintyIgnoresTruncation(t *testing.T) {
	results := []OCRResult{
		{
			Text: "- Test note here", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 1.0,
			Candidates: []OCRCandidate{
				{Text: "- Test note here", Confidence: 1.0},
				{Text: "- Test note", Confidence: 1.0},
				{Text: "- Test", Confidence: 1.0},
			},
		},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Empty(t, result.UncertainLines)
}

func TestReconstructText_UncertaintyIgnoresSpacingVariants(t *testing.T) {
	results := []OCRResult{
		{
			Text: "o Test event", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 1.0,
			Candidates: []OCRCandidate{
				{Text: "o Test event", Confidence: 1.0},
				{Text: "o Testevent", Confidence: 1.0},
				{Text: "oTest event", Confidence: 1.0},
			},
		},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Empty(t, result.UncertainLines)
}

func TestReconstructText_UncertaintyIgnoresWordJoining(t *testing.T) {
	results := []OCRResult{
		{
			Text: "- note 3", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 1.0,
			Candidates: []OCRCandidate{
				{Text: "- note 3", Confidence: 1.0},
				{Text: "- note3", Confidence: 1.0},
				{Text: "-note 3", Confidence: 1.0},
			},
		},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Empty(t, result.UncertainLines)
}

func TestReconstructText_NoUncertaintyWhenPrimaryIsCommonButAlternativeIsGarbled(t *testing.T) {
	results := []OCRResult{
		{
			Text: "• task 2", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 1.0,
			Candidates: []OCRCandidate{
				{Text: "• task 2", Confidence: 1.0},
				{Text: "• Fask 2", Confidence: 1.0},
			},
		},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Empty(t, result.UncertainLines)
}

func TestReconstructText_UncertainWhenLineHasUnknownWords(t *testing.T) {
	results := []OCRResult{
		{Text: "- Benck is going to get us a slot", X: 50, Y: 100, Width: 300, Height: 30, Confidence: 1.0},
		{Text: "- can we hook into existing channels?", X: 50, Y: 200, Width: 300, Height: 30, Confidence: 1.0},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, []int{0}, result.UncertainLines)
}

func TestReconstructText_NotUncertainWhenAllWordsKnown(t *testing.T) {
	results := []OCRResult{
		{Text: "- can we hook into existing channels?", X: 50, Y: 100, Width: 300, Height: 30, Confidence: 1.0},
		{Text: "- meeting with the team", X: 50, Y: 200, Width: 300, Height: 30, Confidence: 1.0},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Empty(t, result.UncertainLines)
}

func TestReconstructText_ConcatenatesUnprefixedLineToAbove(t *testing.T) {
	results := []OCRResult{
		{Text: ". buy milk", X: 50, Y: 100, Width: 200, Height: 30},
		{Text: "and eggs", X: 50, Y: 140, Width: 200, Height: 30},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, ". buy milk and eggs", result.Text)
	assert.Equal(t, []int{0}, result.ConcatenatedLines)
}

func TestReconstructText_FirstUnprefixedLineGetsDashPrefix(t *testing.T) {
	results := []OCRResult{
		{Text: "just a note", X: 50, Y: 100, Width: 200, Height: 30},
	}

	text := ReconstructText(results)
	assert.Equal(t, "- just a note", text)
}

func TestReconstructText_ConcatenatedLineNotDuplicated(t *testing.T) {
	results := []OCRResult{
		{Text: ". task one", X: 50, Y: 100, Width: 200, Height: 30},
		{Text: "continued here", X: 50, Y: 140, Width: 200, Height: 30},
		{Text: ". task two", X: 50, Y: 200, Width: 200, Height: 30},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, ". task one continued here\n. task two", result.Text)
	assert.Equal(t, []int{0}, result.ConcatenatedLines)
}

func TestReconstructTextWithConfidence_DepthResetsOnRoot(t *testing.T) {
	results := []OCRResult{
		{Text: ". a", X: 50, Y: 100, Width: 200, Height: 30, Confidence: 0.9},
		{Text: ". b", X: 100, Y: 200, Width: 200, Height: 30, Confidence: 0.9},
		{Text: ". c", X: 50, Y: 300, Width: 200, Height: 30, Confidence: 0.9},
		{Text: ". d", X: 200, Y: 400, Width: 200, Height: 30, Confidence: 0.9},
	}

	result := ReconstructTextWithConfidence(results, 0.8)
	assert.Equal(t, ". a\n  . b\n. c\n  . d", result.Text)
}
