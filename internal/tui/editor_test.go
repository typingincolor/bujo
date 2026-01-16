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

func TestBuildEditorCmd_EmptyStringDefaultsToVi(t *testing.T) {
	cmd := BuildEditorCmd("", "/tmp/test.txt")

	if cmd.Path == "" {
		t.Error("expected Path to be set")
	}
	args := cmd.Args
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	if args[0] != "vi" {
		t.Errorf("expected 'vi' as command, got %q", args[0])
	}
	if args[1] != "/tmp/test.txt" {
		t.Errorf("expected file path as arg, got %q", args[1])
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

func TestCleanupCaptureTempFile_RemovesFile(t *testing.T) {
	tmpDir := t.TempDir()
	tempFile := filepath.Join(tmpDir, "capture.txt")

	if err := os.WriteFile(tempFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := CleanupCaptureTempFile(tempFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Error("temp file should be deleted after cleanup")
	}
}

func TestCleanupCaptureTempFile_NoErrorWhenFileDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	tempFile := filepath.Join(tmpDir, "nonexistent.txt")

	err := CleanupCaptureTempFile(tempFile)
	if err != nil {
		t.Errorf("expected no error for nonexistent file, got %v", err)
	}
}

func TestEditorFinishedMsg_WithContent_ReturnsSaveCommand(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24

	msg := editorFinishedMsg{
		content: ". Task to save\n- A note",
		err:     nil,
	}

	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("expected a command to be returned for non-empty content")
	}
}

func TestEditorFinishedMsg_WithEmptyContent_ReturnsNoCommand(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24

	msg := editorFinishedMsg{
		content: "",
		err:     nil,
	}

	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("expected no command for empty content")
	}
}

func TestEditorFinishedMsg_WithWhitespaceOnly_ReturnsNoCommand(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24

	msg := editorFinishedMsg{
		content: "   \n\t\n  ",
		err:     nil,
	}

	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("expected no command for whitespace-only content")
	}
}

func TestEditorFinishedMsg_WithError_SetsModelError(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24

	testErr := os.ErrNotExist
	msg := editorFinishedMsg{
		content: "",
		err:     testErr,
	}

	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	if cmd != nil {
		t.Error("expected no command when error occurred")
	}
	if m.err != testErr {
		t.Errorf("expected model error to be set, got %v", m.err)
	}
}

func TestEditorFinishedMsg_DeletesDraftFile(t *testing.T) {
	tmpDir := t.TempDir()
	draftPath := filepath.Join(tmpDir, "draft.txt")

	if err := os.WriteFile(draftPath, []byte("draft content"), 0644); err != nil {
		t.Fatal(err)
	}

	model := New(nil)
	model.width = 80
	model.height = 24
	model.draftPath = draftPath

	msg := editorFinishedMsg{
		content: ". Task",
		err:     nil,
	}

	model.Update(msg)

	if _, err := os.Stat(draftPath); !os.IsNotExist(err) {
		t.Error("draft file should be deleted after editor finishes")
	}
}

func TestEditorFinishedMsg_DeletesDraftOnError(t *testing.T) {
	tmpDir := t.TempDir()
	draftPath := filepath.Join(tmpDir, "draft.txt")

	if err := os.WriteFile(draftPath, []byte("draft content"), 0644); err != nil {
		t.Fatal(err)
	}

	model := New(nil)
	model.width = 80
	model.height = 24
	model.draftPath = draftPath

	msg := editorFinishedMsg{
		content: "",
		err:     os.ErrNotExist,
	}

	model.Update(msg)

	if _, err := os.Stat(draftPath); !os.IsNotExist(err) {
		t.Error("draft file should be deleted even when editor returns error")
	}
}

func TestEditorFinishedMsg_DeletesDraftOnEmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	draftPath := filepath.Join(tmpDir, "draft.txt")

	if err := os.WriteFile(draftPath, []byte("draft content"), 0644); err != nil {
		t.Fatal(err)
	}

	model := New(nil)
	model.width = 80
	model.height = 24
	model.draftPath = draftPath

	msg := editorFinishedMsg{
		content: "",
		err:     nil,
	}

	model.Update(msg)

	if _, err := os.Stat(draftPath); !os.IsNotExist(err) {
		t.Error("draft file should be deleted on empty content")
	}
}

func TestCaptureTempFilePath_FallbackWhenNoHome(t *testing.T) {
	path := CaptureTempFilePath()

	if path == "" {
		t.Error("should always return a path")
	}

	if !filepath.IsAbs(path) {
		t.Error("should return absolute path")
	}

	filename := filepath.Base(path)
	if filename != "capture_temp.txt" && filename != "bujo_capture.txt" {
		t.Errorf("unexpected filename: %s", filename)
	}
}

func TestPrepareEditorFile_CreatesNestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "a", "b", "c", "capture.txt")

	err := PrepareEditorFile(nestedPath, "test content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(nestedPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("expected 'test content', got %q", string(content))
	}
}

func TestBuildEditorCmd_WhitespaceOnlyDefaultsToVi(t *testing.T) {
	cmd := BuildEditorCmd("   ", "/tmp/test.txt")

	if cmd.Path == "" {
		t.Error("expected Path to be set")
	}
	args := cmd.Args
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	if args[0] != "vi" {
		t.Errorf("expected 'vi' as command, got %q", args[0])
	}
}

func TestBuildEditorCmd_MultipleArgs(t *testing.T) {
	cmd := BuildEditorCmd("nvim --noplugin -u NONE", "/tmp/test.txt")

	args := cmd.Args
	if len(args) != 5 {
		t.Fatalf("expected 5 args, got %d: %v", len(args), args)
	}
	if args[0] != "nvim" {
		t.Errorf("expected 'nvim', got %q", args[0])
	}
	if args[1] != "--noplugin" {
		t.Errorf("expected '--noplugin', got %q", args[1])
	}
	if args[2] != "-u" {
		t.Errorf("expected '-u', got %q", args[2])
	}
	if args[4] != "/tmp/test.txt" {
		t.Errorf("expected file path as last arg, got %q", args[4])
	}
}

func TestReadEditorResult_ReturnsMultilineContent(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	content := ". Task one\n- Note one\n- Note two\no Event"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ReadEditorResult(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != content {
		t.Errorf("expected %q, got %q", content, result)
	}
}

func TestGetEditorCommand_HandlesComplexVISUAL(t *testing.T) {
	t.Setenv("VISUAL", "nvim --noplugin")
	t.Setenv("EDITOR", "vim")

	editor := GetEditorCommand()

	if editor != "nvim --noplugin" {
		t.Errorf("expected 'nvim --noplugin', got %q", editor)
	}
}

func TestEditorFinishedMsg_TrimsWhitespace(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24

	msg := editorFinishedMsg{
		content: "  . Task with whitespace  \n\n",
		err:     nil,
	}

	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("expected command for content with leading/trailing whitespace")
	}
}

func TestEditorFinishedMsg_HandlesNewlinesOnly(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24

	msg := editorFinishedMsg{
		content: "\n\n\n",
		err:     nil,
	}

	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("expected no command for newlines-only content")
	}
}

func TestPrepareEditorFile_OverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	tempFile := filepath.Join(tmpDir, "capture.txt")

	if err := os.WriteFile(tempFile, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	newContent := "new draft content"
	err := PrepareEditorFile(tempFile, newContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != newContent {
		t.Errorf("expected %q, got %q", newContent, string(content))
	}
}

func TestCleanupCaptureTempFile_RemovesFileInNestedDir(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "a", "b")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}
	tempFile := filepath.Join(nestedDir, "capture.txt")

	if err := os.WriteFile(tempFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	err := CleanupCaptureTempFile(tempFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Error("temp file should be deleted")
	}
}
