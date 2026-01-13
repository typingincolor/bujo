package domain

import "regexp"

var urlRegex = regexp.MustCompile(`https?://[^\s\)\]]+`)

func ExtractURLs(text string) []string {
	matches := urlRegex.FindAllString(text, -1)
	if matches == nil {
		return []string{}
	}
	return matches
}

func HasURL(text string) bool {
	return urlRegex.MatchString(text)
}
