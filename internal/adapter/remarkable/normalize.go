package remarkable

import (
	"strings"

	"github.com/typingincolor/bujo/internal/domain"
)

func NormalizeOCRIndentation(text string) string {
	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))
	maxDepth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		spaces := countLeadingSpaces(line)
		depth := spaces / 2

		if depth > maxDepth+1 {
			depth = maxDepth + 1
		}
		if depth > 0 {
			maxDepth = depth
		} else {
			maxDepth = 0
		}

		prefix := strings.Repeat("  ", depth)
		if !domain.ParseEntryType(trimmed).IsValid() {
			trimmed = "- " + trimmed
		}
		result = append(result, prefix+trimmed)
	}
	return strings.Join(result, "\n")
}

func countLeadingSpaces(line string) int {
	n := 0
	for _, ch := range line {
		switch ch {
		case ' ':
			n++
		case '\t':
			n += 2
		default:
			return n
		}
	}
	return n
}
