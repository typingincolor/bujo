package tui

import (
	"github.com/charmbracelet/glamour"
)

func renderMarkdown(content string) (string, error) {
	if content == "" {
		return "", nil
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return content, err
	}

	rendered, err := r.Render(content)
	if err != nil {
		return content, err
	}

	return rendered, nil
}
