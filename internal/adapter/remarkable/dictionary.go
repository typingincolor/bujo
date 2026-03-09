package remarkable

import (
	_ "embed"
	"strings"
)

//go:embed words.txt
var wordsFile string

var commonWords map[string]bool

func init() {
	lines := strings.Split(wordsFile, "\n")
	commonWords = make(map[string]bool, len(lines))
	for _, w := range lines {
		w = strings.TrimSpace(w)
		if w != "" {
			commonWords[w] = true
		}
	}
}

func isCommonWord(word string) bool {
	return commonWords[strings.ToLower(word)]
}
