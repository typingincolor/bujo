package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetEditorCommand() string {
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	return "vi"
}

func BuildEditorCmd(editorCmd string, filePath string) *exec.Cmd {
	parts := strings.Fields(editorCmd)
	args := append(parts[1:], filePath)
	return exec.Command(parts[0], args...)
}

func ReadEditorResult(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func CaptureTempFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "bujo_capture.txt")
	}
	return filepath.Join(home, ".bujo", "capture_temp.txt")
}

func PrepareEditorFile(filePath string, content string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filePath, []byte(content), 0644)
}
