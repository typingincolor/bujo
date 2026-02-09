package domain

import (
	"regexp"
	"sort"
	"strings"
)

var tagPattern = regexp.MustCompile(`#([a-zA-Z][a-zA-Z0-9-]*)`)
var mentionPattern = regexp.MustCompile(`@([a-zA-Z][a-zA-Z0-9-]*(?:\.[a-zA-Z][a-zA-Z0-9-]*)*)`)

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

func ExtractMentions(content string) []string {
	matches := mentionPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var mentions []string

	for _, match := range matches {
		mention := strings.ToLower(match[1])
		if !seen[mention] {
			seen[mention] = true
			mentions = append(mentions, mention)
		}
	}

	sort.Strings(mentions)
	return mentions
}
