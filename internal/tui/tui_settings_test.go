package tui

import (
	"strings"
	"testing"
)

func newSettingsModel() Model {
	model := NewWithConfig(Config{
		Version: "1.2.3",
		Commit:  "abc1234",
		Date:    "2026-02-06",
		DBPath:  "/home/user/.bujo/bujo.db",
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeSettings
	return model
}

func TestSettings_ShowsVersion(t *testing.T) {
	model := newSettingsModel()
	output := model.renderSettingsContent()

	if !strings.Contains(output, "1.2.3") {
		t.Error("expected settings to show version")
	}
}

func TestSettings_ShowsCommit(t *testing.T) {
	model := newSettingsModel()
	output := model.renderSettingsContent()

	if !strings.Contains(output, "abc1234") {
		t.Error("expected settings to show commit hash")
	}
}

func TestSettings_ShowsBuildDate(t *testing.T) {
	model := newSettingsModel()
	output := model.renderSettingsContent()

	if !strings.Contains(output, "2026-02-06") {
		t.Error("expected settings to show build date")
	}
}

func TestSettings_ShowsDBPath(t *testing.T) {
	model := newSettingsModel()
	output := model.renderSettingsContent()

	if !strings.Contains(output, "/home/user/.bujo/bujo.db") {
		t.Error("expected settings to show database path")
	}
}

func TestSettings_DefaultValuesWhenNotConfigured(t *testing.T) {
	model := NewWithConfig(Config{})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeSettings

	output := model.renderSettingsContent()

	if !strings.Contains(output, "dev") {
		t.Error("expected default version 'dev' when not configured")
	}
}

func TestSettings_ShowsKeyboardShortcuts(t *testing.T) {
	model := newSettingsModel()
	output := model.renderSettingsContent()

	if !strings.Contains(output, "Keyboard") {
		t.Error("expected settings to show keyboard shortcuts section")
	}
}
