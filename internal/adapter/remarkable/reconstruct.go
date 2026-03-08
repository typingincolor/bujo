package remarkable

import (
	"math"
	"sort"
	"strings"
)

const (
	defaultIndentWidth      = 50.0
	defaultConfidenceThreshold = 0.8
)

type ReconstructResult struct {
	Text               string
	LowConfidenceCount int
}

func ReconstructText(results []OCRResult) string {
	return ReconstructTextWithConfidence(results, defaultConfidenceThreshold).Text
}

func ReconstructTextWithConfidence(results []OCRResult, threshold float32) ReconstructResult {
	if len(results) == 0 {
		return ReconstructResult{}
	}

	sorted := make([]OCRResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Y < sorted[j].Y
	})

	minX := math.MaxFloat64
	for _, r := range sorted {
		if float64(r.X) < minX {
			minX = float64(r.X)
		}
	}

	var lines []string
	var lowConfidenceCount int
	var maxDepth int

	for _, r := range sorted {
		depth := int(math.Round((r.X - minX) / defaultIndentWidth))
		if depth > maxDepth+1 {
			depth = maxDepth + 1
		}
		if depth == 0 {
			maxDepth = 0
		} else {
			maxDepth = depth
		}

		indent := strings.Repeat("  ", depth)
		lines = append(lines, indent+r.Text)

		if r.Confidence < threshold {
			lowConfidenceCount++
		}
	}

	return ReconstructResult{
		Text:               strings.Join(lines, "\n"),
		LowConfidenceCount: lowConfidenceCount,
	}
}
