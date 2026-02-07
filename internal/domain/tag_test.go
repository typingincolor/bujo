package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
