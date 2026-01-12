package tui

// renderMarkdown renders markdown content using the pre-initialized glamour renderer.
// If the renderer is not available or rendering fails, the original content is returned
// as a fallback, ensuring graceful degradation.
//
// Performance note: This method reuses the Model's markdownRenderer instance to avoid
// the overhead of creating a new renderer on every call. The renderer is created once
// during Model initialization with a reasonable default width.
func (m Model) renderMarkdown(content string) (string, error) {
	if content == "" {
		return "", nil
	}

	// If renderer isn't available, return original content as fallback
	if m.markdownRenderer == nil {
		return content, nil
	}

	// Render the markdown using the pre-created renderer
	// If this fails, caller will fall back to raw content
	rendered, err := m.markdownRenderer.Render(content)
	if err != nil {
		return content, err
	}

	return rendered, nil
}
