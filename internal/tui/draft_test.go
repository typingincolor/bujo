package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/service"
)

func TestDraftPath_ReturnsCorrectLocation(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	expected := filepath.Join(home, ".bujo", "capture_draft.txt")
	actual := DraftPath()

	if actual != expected {
		t.Errorf("expected draft path %s, got %s", expected, actual)
	}
}

func TestLoadDraft_ReturnsEmptyWhenNoFile(t *testing.T) {
	// Use a temp directory to ensure no draft exists
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "nonexistent_draft.txt")

	content, exists := LoadDraft(path)

	if exists {
		t.Error("expected exists to be false for nonexistent file")
	}
	if content != "" {
		t.Errorf("expected empty content, got %q", content)
	}
}

func TestLoadDraft_ReturnsContentWhenFileExists(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "draft.txt")

	// Create draft file
	draftContent := ". Task one\n- Note here"
	if err := os.WriteFile(path, []byte(draftContent), 0644); err != nil {
		t.Fatal(err)
	}

	content, exists := LoadDraft(path)

	if !exists {
		t.Error("expected exists to be true for existing file")
	}
	if content != draftContent {
		t.Errorf("expected content %q, got %q", draftContent, content)
	}
}

func TestSaveDraft_CreatesFileWithContent(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "draft.txt")

	content := ". My task"
	err := SaveDraft(path, content)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created with correct content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read draft file: %v", err)
	}
	if string(data) != content {
		t.Errorf("expected file content %q, got %q", content, string(data))
	}
}

func TestSaveDraft_CreatesDirectoryIfNeeded(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "subdir", "draft.txt")

	content := ". My task"
	err := SaveDraft(path, content)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("draft file was not created")
	}
}

func TestSaveDraft_OverwritesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "draft.txt")

	// Create initial draft
	if err := os.WriteFile(path, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	newContent := ". New task"
	err := SaveDraft(path, newContent)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read draft file: %v", err)
	}
	if string(data) != newContent {
		t.Errorf("expected file content %q, got %q", newContent, string(data))
	}
}

func TestSaveDraft_DoesNotSaveEmptyContent(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "draft.txt")

	err := SaveDraft(path, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should not exist for empty content
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("draft file should not be created for empty content")
	}
}

func TestDeleteDraft_RemovesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "draft.txt")

	// Create draft file
	if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	err := DeleteDraft(path)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("draft file should be deleted")
	}
}

func TestDeleteDraft_NoErrorWhenFileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "nonexistent.txt")

	err := DeleteDraft(path)

	if err != nil {
		t.Errorf("expected no error for nonexistent file, got %v", err)
	}
}

// Integration tests for capture mode draft handling

func TestCaptureMode_LoadsDraftOnEnter(t *testing.T) {
	tempDir := t.TempDir()
	draftPath := filepath.Join(tempDir, "draft.txt")
	draftContent := ". Saved task from before"

	// Create draft file
	if err := os.WriteFile(draftPath, []byte(draftContent), 0644); err != nil {
		t.Fatal(err)
	}

	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.draftPath = draftPath

	// Press 'c' to enter capture mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.captureMode.active {
		t.Error("should be in capture mode")
	}
	if !m.captureMode.draftExists {
		t.Error("draftExists should be true when draft file exists")
	}
	if m.captureMode.draftContent != draftContent {
		t.Errorf("expected draft content %q, got %q", draftContent, m.captureMode.draftContent)
	}
}

func TestCaptureMode_NoDraftWhenFileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	draftPath := filepath.Join(tempDir, "nonexistent.txt")

	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.draftPath = draftPath

	// Press 'c' to enter capture mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.captureMode.active {
		t.Error("should be in capture mode")
	}
	if m.captureMode.draftExists {
		t.Error("draftExists should be false when no draft file")
	}
}

func TestCaptureMode_SavesDraftOnTyping(t *testing.T) {
	tempDir := t.TempDir()
	draftPath := filepath.Join(tempDir, "draft.txt")

	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.draftPath = draftPath
	model.captureMode = captureState{active: true, content: ". Task"}

	// Type a character
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Verify draft was saved
	content, exists := LoadDraft(draftPath)
	if !exists {
		t.Error("draft file should exist after typing")
	}
	if content != m.captureMode.content {
		t.Errorf("draft content %q should match capture content %q", content, m.captureMode.content)
	}
}

func TestCaptureMode_DeletesDraftOnSave(t *testing.T) {
	tempDir := t.TempDir()
	draftPath := filepath.Join(tempDir, "draft.txt")

	// Create draft file
	if err := os.WriteFile(draftPath, []byte(". Draft content"), 0644); err != nil {
		t.Fatal(err)
	}

	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.draftPath = draftPath
	model.captureMode = captureState{active: true, content: ". Task to save"}

	// Press Ctrl+X to save
	msg := tea.KeyMsg{Type: tea.KeyCtrlX}
	model.Update(msg)

	// Verify draft was deleted
	if _, exists := LoadDraft(draftPath); exists {
		t.Error("draft file should be deleted after save")
	}
}

func TestCaptureMode_DeletesDraftOnCancel(t *testing.T) {
	tempDir := t.TempDir()
	draftPath := filepath.Join(tempDir, "draft.txt")

	// Create draft file
	if err := os.WriteFile(draftPath, []byte(". Draft content"), 0644); err != nil {
		t.Fatal(err)
	}

	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.draftPath = draftPath
	model.captureMode = captureState{active: true, content: ". Task", confirmCancel: true}

	// Press 'y' to confirm cancel
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	model.Update(msg)

	// Verify draft was deleted
	if _, exists := LoadDraft(draftPath); exists {
		t.Error("draft file should be deleted after cancel")
	}
}

func TestCaptureMode_DeletesDraftOnEmptyExit(t *testing.T) {
	tempDir := t.TempDir()
	draftPath := filepath.Join(tempDir, "draft.txt")

	// Create draft file (maybe leftover from previous session)
	if err := os.WriteFile(draftPath, []byte(". Old draft"), 0644); err != nil {
		t.Fatal(err)
	}

	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.draftPath = draftPath
	model.captureMode = captureState{active: true, content: ""}

	// Press Ctrl+X with empty content
	msg := tea.KeyMsg{Type: tea.KeyCtrlX}
	model.Update(msg)

	// Verify draft was deleted
	if _, exists := LoadDraft(draftPath); exists {
		t.Error("draft file should be deleted on empty exit")
	}
}

func TestCaptureMode_EndToEnd_DraftPersistence(t *testing.T) {
	tempDir := t.TempDir()
	draftPath := filepath.Join(tempDir, "draft.txt")

	// === Session 1: Enter capture mode, type content, "crash" ===
	model1 := New(nil)
	model1.agenda = &service.MultiDayAgenda{}
	model1.draftPath = draftPath

	// Enter capture mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	newModel, _ := model1.Update(msg)
	m1 := newModel.(Model)

	if !m1.captureMode.active {
		t.Fatal("should be in capture mode")
	}

	// Type ". Test task" - note: "." at line start auto-converts to "• " (with space)
	// So ". Test" becomes "•  Test" (two spaces: one from conversion, one from input)
	for _, r := range ". Test task" {
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		newModel, _ = m1.Update(msg)
		m1 = newModel.(Model)
	}

	expectedContent := "•  Test task" // "." converts to "• ", then " Test task" is added
	if m1.captureMode.content != expectedContent {
		t.Fatalf("expected content %q, got %q", expectedContent, m1.captureMode.content)
	}

	// Verify draft file was saved (simulating "crash" - we just check the file exists)
	savedDraft, exists := LoadDraft(draftPath)
	if !exists {
		t.Fatal("draft file should exist after typing")
	}
	if savedDraft != expectedContent {
		t.Fatalf("saved draft should be %q, got %q", expectedContent, savedDraft)
	}

	// === Session 2: New model (app restart), enter capture mode, see draft prompt ===
	model2 := New(nil)
	model2.agenda = &service.MultiDayAgenda{}
	model2.draftPath = draftPath

	// Enter capture mode
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	newModel, _ = model2.Update(msg)
	m2 := newModel.(Model)

	if !m2.captureMode.draftExists {
		t.Fatal("draftExists should be true on re-entry")
	}
	if m2.captureMode.draftContent != expectedContent {
		t.Fatalf("draftContent should be %q, got %q", expectedContent, m2.captureMode.draftContent)
	}

	// Press 'y' to restore
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	newModel, _ = m2.Update(msg)
	m2 = newModel.(Model)

	if m2.captureMode.draftExists {
		t.Fatal("draftExists should be false after restore")
	}
	if m2.captureMode.content != expectedContent {
		t.Fatalf("content should be restored to %q, got %q", expectedContent, m2.captureMode.content)
	}

	// === Session 3: Type more, then save with Ctrl+X ===
	// Add more content
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}
	newModel, _ = m2.Update(msg)
	m2 = newModel.(Model)

	// Save with Ctrl+X
	msg = tea.KeyMsg{Type: tea.KeyCtrlX}
	newModel, _ = m2.Update(msg)
	m2 = newModel.(Model)

	if m2.captureMode.active {
		t.Fatal("should exit capture mode after Ctrl+X")
	}

	// Verify draft was deleted
	if _, exists := LoadDraft(draftPath); exists {
		t.Fatal("draft file should be deleted after save")
	}

	t.Log("End-to-end draft persistence test passed!")
}
