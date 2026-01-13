package tui

func (m Model) renderMarkdown(content string) (string, error) {
	if content == "" {
		return "", nil
	}

	if m.markdownRenderer == nil {
		return content, nil
	}

	rendered, err := m.markdownRenderer.Render(content)
	if err != nil {
		return content, err
	}

	return rendered, nil
}
