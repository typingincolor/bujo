package remarkable

import (
	"fmt"
	"os"
	"path/filepath"
)

const remarkableScreenWidth = 1404

func RenderPageToPNG(dir string, pageID string, rmData []byte) (string, error) {
	strokes, err := ParseRM(rmData)
	if err != nil {
		return "", fmt.Errorf("parse .rm failed: %w", err)
	}

	pngData, err := RenderStrokes(strokes)
	if err != nil {
		return "", fmt.Errorf("render failed: %w", err)
	}

	pngPath := filepath.Join(dir, pageID+".png")
	if err := os.WriteFile(pngPath, pngData, 0644); err != nil {
		return "", fmt.Errorf("write PNG failed: %w", err)
	}

	return pngPath, nil
}
