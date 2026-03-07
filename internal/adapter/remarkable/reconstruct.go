package remarkable

import (
	"math"
	"sort"
	"strings"
)

const defaultIndentWidth = 50.0

func ReconstructText(results []OCRResult) string {
	if len(results) == 0 {
		return ""
	}

	sorted := make([]OCRResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Y < sorted[j].Y
	})

	minX := math.MaxFloat64
	for _, r := range sorted {
		if r.X < minX {
			minX = r.X
		}
	}

	var lines []string
	for _, r := range sorted {
		depth := int(math.Round((r.X - minX) / defaultIndentWidth))
		indent := strings.Repeat("  ", depth)
		lines = append(lines, indent+r.Text)
	}

	return strings.Join(lines, "\n")
}
