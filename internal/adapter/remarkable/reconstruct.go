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

	sort.Slice(results, func(i, j int) bool {
		return results[i].Y < results[j].Y
	})

	minX := math.MaxFloat64
	for _, r := range results {
		if r.X < minX {
			minX = r.X
		}
	}

	var lines []string
	for _, r := range results {
		depth := int(math.Round((r.X - minX) / defaultIndentWidth))
		indent := strings.Repeat("  ", depth)
		lines = append(lines, indent+r.Text)
	}

	return strings.Join(lines, "\n")
}
