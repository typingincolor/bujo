package tui

import (
	"strings"
	"testing"
)

func TestHighlightURLs_SingleURL(t *testing.T) {
	content := "check https://example.com now"
	result := highlightURLs(content)

	if result == content {
		t.Error("URL should be highlighted (should contain ANSI codes)")
	}
	if !strings.Contains(result, "check ") {
		t.Error("text before URL should be preserved")
	}
	if !strings.Contains(result, " now") {
		t.Error("text after URL should be preserved")
	}
}

func TestHighlightURLs_NoURL_NoChange(t *testing.T) {
	content := "just plain text"
	result := highlightURLs(content)

	if result != content {
		t.Errorf("no URL should return unchanged content, got '%s'", result)
	}
}

func TestHighlightURLs_MultipleURLs(t *testing.T) {
	content := "see https://one.com and https://two.com"
	result := highlightURLs(content)

	if result == content {
		t.Error("URLs should be highlighted")
	}
	if !strings.Contains(result, " and ") {
		t.Error("text between URLs should be preserved")
	}
	if len(result) <= len(content) {
		t.Error("result should be longer due to ANSI codes")
	}
}

func TestHighlightURLs_URLAtStart(t *testing.T) {
	content := "https://example.com is great"
	result := highlightURLs(content)

	if result == content {
		t.Error("URL at start should be highlighted")
	}
	if !strings.Contains(result, " is great") {
		t.Error("text after URL should be preserved")
	}
}

func TestHighlightURLs_URLAtEnd(t *testing.T) {
	content := "visit https://example.com"
	result := highlightURLs(content)

	if result == content {
		t.Error("URL at end should be highlighted")
	}
	if !strings.Contains(result, "visit ") {
		t.Error("text before URL should be preserved")
	}
}

func TestHighlightURLs_EmptyString(t *testing.T) {
	result := highlightURLs("")

	if result != "" {
		t.Errorf("empty string should return empty, got '%s'", result)
	}
}

func TestHighlightURLs_HTTPUrl(t *testing.T) {
	content := "old site http://example.com here"
	result := highlightURLs(content)

	if result == content {
		t.Error("http URL should also be highlighted")
	}
}
