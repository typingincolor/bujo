package remarkable

import (
	_ "embed"
	"regexp"
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

var wordPattern = regexp.MustCompile(`[a-zA-Z]+`)

func hasUnknownWords(text string) bool {
	stripped := stripBujoPrefix(text)
	tokens := strings.Fields(stripped)
	for _, token := range tokens {
		if strings.HasPrefix(token, "@") {
			continue
		}
		words := wordPattern.FindAllString(token, -1)
		for _, w := range words {
			if len(w) <= 2 {
				continue
			}
			if !isCommonWord(w) {
				return true
			}
		}
	}
	return false
}
