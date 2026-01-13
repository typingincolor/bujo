package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractURLs(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{
			name: "single http URL",
			text: "Check out http://example.com for details",
			want: []string{"http://example.com"},
		},
		{
			name: "single https URL",
			text: "Visit https://github.com/user/repo",
			want: []string{"https://github.com/user/repo"},
		},
		{
			name: "multiple URLs",
			text: "See https://example.com and http://test.com",
			want: []string{"https://example.com", "http://test.com"},
		},
		{
			name: "URL with path and query",
			text: "Link: https://example.com/path?query=value&foo=bar",
			want: []string{"https://example.com/path?query=value&foo=bar"},
		},
		{
			name: "URL with fragment",
			text: "Go to https://docs.example.com#section",
			want: []string{"https://docs.example.com#section"},
		},
		{
			name: "no URLs",
			text: "This is just plain text",
			want: []string{},
		},
		{
			name: "empty string",
			text: "",
			want: []string{},
		},
		{
			name: "URL at start of text",
			text: "https://example.com is the site",
			want: []string{"https://example.com"},
		},
		{
			name: "URL at end of text",
			text: "Visit the site at https://example.com",
			want: []string{"https://example.com"},
		},
		{
			name: "URL in parentheses",
			text: "See the docs (https://docs.example.com) for more",
			want: []string{"https://docs.example.com"},
		},
		{
			name: "URL with port",
			text: "Local server at http://localhost:8080/api",
			want: []string{"http://localhost:8080/api"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractURLs(tt.text)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHasURL(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{
			name: "has URL",
			text: "Check https://example.com",
			want: true,
		},
		{
			name: "no URL",
			text: "Just plain text",
			want: false,
		},
		{
			name: "empty string",
			text: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasURL(tt.text)
			assert.Equal(t, tt.want, got)
		})
	}
}
