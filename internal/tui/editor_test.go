package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetEditorCommand_UsesVISUAL(t *testing.T) {
	t.Setenv("VISUAL", "code --wait")
	t.Setenv("EDITOR", "nano")

	editor := GetEditorCommand()

	if editor != "code --wait" {
		t.Errorf("expected 'code --wait', got %q", editor)
	}
}

func TestGetEditorCommand_UsesEDITOR_WhenVISUALNotSet(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "nano")

	editor := GetEditorCommand()

	if editor != "nano" {
		t.Errorf("expected 'nano', got %q", editor)
	}
}

func TestGetEditorCommand_DefaultsToVi(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	editor := GetEditorCommand()

	if editor != "vi" {
		t.Errorf("expected 'vi', got %q", editor)
	}
}

func TestBuildEditorCmd_SimpleEditor(t *testing.T) {
	cmd := BuildEditorCmd("vi", "/tmp/test.txt")

	if cmd.Path == "" {
		t.Error("expected Path to be set")
	}
	args := cmd.Args
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	if args[1] != "/tmp/test.txt" {
		t.Errorf("expected file path as arg, got %q", args[1])
	}
}

func TestBuildEditorCmd_EditorWithArgs(t *testing.T) {
	cmd := BuildEditorCmd("code --wait", "/tmp/test.txt")

	if cmd.Path == "" {
		t.Error("expected Path to be set")
	}
	args := cmd.Args
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d: %v", len(args), args)
	}
	if args[1] != "--wait" {
		t.Errorf("expected '--wait' as second arg, got %q", args[1])
	}
	if args[2] != "/tmp/test.txt" {
		t.Errorf("expected file path as last arg, got %q", args[2])
	}
}

func TestOpenEditorAndGetContent_ReturnsFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	content := ". Test task\n- A note"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ReadEditorResult(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != content {
		t.Errorf("expected %q, got %q", content, result)
	}
}

func TestReadEditorResult_ReturnsEmptyForNonExistentFile(t *testing.T) {
	result, err := ReadEditorResult("/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestCaptureTempFilePath_ReturnsValidPath(t *testing.T) {
	path := CaptureTempFilePath()

	if path == "" {
		t.Error("expected non-empty path")
	}
	if !filepath.IsAbs(path) {
		t.Error("expected absolute path")
	}
}

func TestCaptureTempFilePath_IncludesDirectory(t *testing.T) {
	path := CaptureTempFilePath()

	dir := filepath.Dir(path)
	if dir == "" {
		t.Error("expected directory component in path")
	}
}

func TestEditorFinishedMsg_ParsesContent(t *testing.T) {
	msg := editorFinishedMsg{
		content: ". Task one\n- Note here",
		err:     nil,
	}

	if msg.content != ". Task one\n- Note here" {
		t.Errorf("expected content to be preserved, got %q", msg.content)
	}
	if msg.err != nil {
		t.Errorf("expected no error, got %v", msg.err)
	}
}

func TestPrepareEditorFile_CreatesFileWithDraft(t *testing.T) {
	tmpDir := t.TempDir()
	tempFile := filepath.Join(tmpDir, "capture.txt")
	draftContent := ". Existing draft content"

	err := PrepareEditorFile(tempFile, draftContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}
	if string(content) != draftContent {
		t.Errorf("expected %q, got %q", draftContent, string(content))
	}
}

func TestPrepareEditorFile_CreatesEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	tempFile := filepath.Join(tmpDir, "capture.txt")

	err := PrepareEditorFile(tempFile, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}
	if string(content) != "" {
		t.Errorf("expected empty file, got %q", string(content))
	}
}
