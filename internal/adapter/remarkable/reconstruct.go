package remarkable

import (
	"math"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	defaultIndentWidth         = 50.0
	DefaultConfidenceThreshold = 0.8
	lineOverlapThreshold       = 0.5
)

type ReconstructResult struct {
	Text               string
	LowConfidenceCount int
	LowConfidenceLines []int
	UncertainLines     []int
}

type mergedLine struct {
	fragments  []OCRResult
	minX       float64
	confidence float32
}

func ReconstructText(results []OCRResult) string {
	return ReconstructTextWithConfidence(results, DefaultConfidenceThreshold).Text
}

func ReconstructTextWithConfidence(results []OCRResult, threshold float32) ReconstructResult {
	if len(results) == 0 {
		return ReconstructResult{}
	}

	merged := mergeLines(results)

	minX := math.MaxFloat64
	for _, m := range merged {
		if m.minX < minX {
			minX = m.minX
		}
	}

	var lines []string
	var lowConfidenceCount int
	lowConfidenceLines := []int{}
	uncertainLines := []int{}
	var maxDepth int

	for i, m := range merged {
		depth := int(math.Round((m.minX - minX) / defaultIndentWidth))
		if depth > maxDepth+1 {
			depth = maxDepth + 1
		}
		if depth == 0 {
			maxDepth = 0
		} else {
			maxDepth = depth
		}

		indent := strings.Repeat("  ", depth)
		var text string
		if len(m.fragments) == 1 {
			text = selectBestText(m.fragments[0])
		} else {
			text = joinFragments(m.fragments)
		}
		if !hasBujoPrefix(text) {
			text = "- " + text
		}
		lines = append(lines, indent+text)

		if m.confidence < threshold {
			lowConfidenceCount++
			lowConfidenceLines = append(lowConfidenceLines, i)
		}
		// Multi-fragment lines skip candidate uncertainty: fragments are separate OCR observations
		// joined by position, so per-fragment candidates don't map to the merged text.
		if len(m.fragments) == 1 && hasCandidateDisagreement(m.fragments[0]) {
			uncertainLines = append(uncertainLines, i)
		} else if hasUnknownWords(text) {
			uncertainLines = append(uncertainLines, i)
		}
	}

	return ReconstructResult{
		Text:               strings.Join(lines, "\n"),
		LowConfidenceCount: lowConfidenceCount,
		LowConfidenceLines: lowConfidenceLines,
		UncertainLines:     uncertainLines,
	}
}

func mergeLines(results []OCRResult) []mergedLine {
	sorted := make([]OCRResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Y < sorted[j].Y
	})

	var merged []mergedLine
	for _, r := range sorted {
		added := false
		for i := range merged {
			if overlapsVertically(merged[i].fragments[0], r) {
				merged[i].fragments = append(merged[i].fragments, r)
				if r.X < merged[i].minX {
					merged[i].minX = r.X
				}
				if r.Confidence < merged[i].confidence {
					merged[i].confidence = r.Confidence
				}
				added = true
				break
			}
		}
		if !added {
			merged = append(merged, mergedLine{
				fragments:  []OCRResult{r},
				minX:       r.X,
				confidence: r.Confidence,
			})
		}
	}

	for i := range merged {
		sort.Slice(merged[i].fragments, func(a, b int) bool {
			return merged[i].fragments[a].X < merged[i].fragments[b].X
		})
	}

	return merged
}

func overlapsVertically(a, b OCRResult) bool {
	aTop := a.Y
	aBottom := a.Y + a.Height
	bTop := b.Y
	bBottom := b.Y + b.Height

	overlapStart := math.Max(aTop, bTop)
	overlapEnd := math.Min(aBottom, bBottom)
	overlap := math.Max(0, overlapEnd-overlapStart)

	shorter := math.Min(a.Height, b.Height)
	if shorter <= 0 {
		return false
	}
	return overlap/shorter >= lineOverlapThreshold
}

func joinFragments(fragments []OCRResult) string {
	var parts []string
	for _, f := range fragments {
		parts = append(parts, f.Text)
	}
	return strings.Join(parts, " ")
}

var bujoSymbols = map[rune]bool{
	'.': true, '-': true, 'o': true, 'x': true,
	'>': true, '?': true, 'a': true,
	'•': true, '–': true, '○': true, '✓': true,
	'→': true, '★': true, '↳': true,
}

func hasBujoPrefix(text string) bool {
	r, size := utf8.DecodeRuneInString(text)
	return bujoSymbols[r] && len(text) > size && text[size] == ' '
}

func selectBestText(r OCRResult) string {
	if len(r.Candidates) == 0 {
		return r.Text
	}
	for _, c := range r.Candidates {
		if hasBujoPrefix(c.Text) {
			return c.Text
		}
	}
	return r.Text
}

func hasCandidateDisagreement(r OCRResult) bool {
	if len(r.Candidates) < 2 {
		return false
	}
	ref := contentWords(r.Candidates[0].Text)
	for _, c := range r.Candidates[1:] {
		words := contentWords(c.Text)
		if wordsMeaningfullyDisagree(ref, words) {
			return true
		}
	}
	return false
}

func contentWords(text string) []string {
	stripped := stripBujoPrefix(text)
	fields := strings.Fields(stripped)
	for i, f := range fields {
		fields[i] = strings.ToLower(f)
	}
	return fields
}

func wordsMeaningfullyDisagree(a, b []string) bool {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] == b[i] {
			continue
		}
		if strings.HasPrefix(a[i], b[i]) || strings.HasPrefix(b[i], a[i]) {
			continue
		}
		if !isCommonWord(a[i]) || isCommonWord(b[i]) {
			return true
		}
	}
	return false
}

func stripBujoPrefix(text string) string {
	r, size := utf8.DecodeRuneInString(text)
	if !bujoSymbols[r] {
		return text
	}
	if len(text) > size && text[size] == ' ' {
		return text[size+1:]
	}
	return text[size:]
}
