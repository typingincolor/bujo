package domain

import (
	"regexp"
	"sort"
	"strings"
)

var tagPattern = regexp.MustCompile(`#([a-zA-Z][a-zA-Z0-9-]*)`)

func ExtractTags(content string) []string {
	matches := tagPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var tags []string

	for _, match := range matches {
		tag := strings.ToLower(match[1])
		if !seen[tag] {
			seen[tag] = true
			tags = append(tags, tag)
		}
	}

	sort.Strings(tags)
	return tags
}
