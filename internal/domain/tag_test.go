package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractMentions(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "no mentions",
			content:  "buy groceries",
			expected: nil,
		},
		{
			name:     "single first name mention",
			content:  "call @john about the project",
			expected: []string{"john"},
		},
		{
			name:     "mention with surname",
			content:  "meeting with @john.smith tomorrow",
			expected: []string{"john.smith"},
		},
		{
			name:     "multiple mentions",
			content:  "meet @alice and @bob for lunch",
			expected: []string{"alice", "bob"},
		},
		{
			name:     "mentions normalized to lowercase",
			content:  "email @John.Smith and @ALICE",
			expected: []string{"alice", "john.smith"},
		},
		{
			name:     "mention at start of content",
			content:  "@sarah review the PR",
			expected: []string{"sarah"},
		},
		{
			name:     "mention with hyphenated surname",
			content:  "call @mary.jones-smith about budget",
			expected: []string{"mary.jones-smith"},
		},
		{
			name:     "duplicate mentions deduplicated",
			content:  "@john said @john will handle it",
			expected: []string{"john"},
		},
		{
			name:     "mention must start with letter",
			content:  "email @123 is invalid",
			expected: nil,
		},
		{
			name:     "empty content",
			content:  "",
			expected: nil,
		},
		{
			name:     "at sign without name",
			content:  "issue @ is broken",
			expected: nil,
		},
		{
			name:     "mixed tags and mentions",
			content:  "call @john about #project",
			expected: []string{"john"},
		},
		{
			name:     "mention with numbers in name",
			content:  "ping @user42 about deploy",
			expected: []string{"user42"},
		},
		{
			name:     "surname must start with letter",
			content:  "meet @john.123",
			expected: []string{"john"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractMentions(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTags(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "no tags",
			content:  "buy groceries",
			expected: nil,
		},
		{
			name:     "single tag",
			content:  "buy groceries #shopping",
			expected: []string{"shopping"},
		},
		{
			name:     "multiple tags",
			content:  "buy groceries #shopping #errands",
			expected: []string{"errands", "shopping"},
		},
		{
			name:     "tag at start of content",
			content:  "#urgent fix the build",
			expected: []string{"urgent"},
		},
		{
			name:     "tag in middle of content",
			content:  "fix the #urgent build issue",
			expected: []string{"urgent"},
		},
		{
			name:     "tags normalized to lowercase",
			content:  "meeting #Work #PERSONAL",
			expected: []string{"personal", "work"},
		},
		{
			name:     "tag with hyphens",
			content:  "review #code-review #pull-request",
			expected: []string{"code-review", "pull-request"},
		},
		{
			name:     "tag with numbers",
			content:  "sprint #sprint2 planning",
			expected: []string{"sprint2"},
		},
		{
			name:     "tag must start with letter",
			content:  "issue #123 is broken",
			expected: nil,
		},
		{
			name:     "duplicate tags deduplicated",
			content:  "#shopping buy milk #shopping",
			expected: []string{"shopping"},
		},
		{
			name:     "tag cannot start with hyphen",
			content:  "test #-invalid tag",
			expected: nil,
		},
		{
			name:     "empty content",
			content:  "",
			expected: nil,
		},
		{
			name:     "hash without tag name",
			content:  "issue # is broken",
			expected: nil,
		},
		{
			name:     "mixed valid and invalid tags",
			content:  "#valid #123invalid #also-valid",
			expected: []string{"also-valid", "valid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTags(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}
