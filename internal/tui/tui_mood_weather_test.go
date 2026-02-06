package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newJournalModel() Model {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeJournal
	return model
}

func TestMood_MKeyActivatesMoodMode(t *testing.T) {
	model := newJournalModel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if !m.presetPicker.active {
		t.Error("expected preset picker to be active after pressing 'm'")
	}
	if m.presetPicker.kind != pickerMood {
		t.Error("expected picker kind to be pickerMood")
	}
	if !m.presetPicker.pickerMode {
		t.Error("expected pickerMode to be true")
	}
}

func TestMood_EscCancelsMoodMode(t *testing.T) {
	model := newJournalModel()
	model.presetPicker = presetPickerState{
		active:     true,
		pickerMode: true,
		kind:       pickerMood,
		input:      textinput.New(),
		items:      moodPresets,
	}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.presetPicker.active {
		t.Error("expected preset picker to be inactive after Esc")
	}
}

func TestMood_UpDownNavigatesPresets(t *testing.T) {
	model := newJournalModel()
	model.presetPicker = presetPickerState{
		active:      true,
		pickerMode:  true,
		kind:        pickerMood,
		input:       textinput.New(),
		items:       moodPresets,
		selectedIdx: 0,
	}

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.presetPicker.selectedIdx != 1 {
		t.Errorf("expected selectedIdx=1 after Down, got %d", m.presetPicker.selectedIdx)
	}

	msg = tea.KeyMsg{Type: tea.KeyUp}
	updated, _ = m.Update(msg)
	m = updated.(Model)

	if m.presetPicker.selectedIdx != 0 {
		t.Errorf("expected selectedIdx=0 after Up, got %d", m.presetPicker.selectedIdx)
	}
}

func TestMood_UpDoesNotGoBelowZero(t *testing.T) {
	model := newJournalModel()
	model.presetPicker = presetPickerState{
		active:      true,
		pickerMode:  true,
		kind:        pickerMood,
		input:       textinput.New(),
		items:       moodPresets,
		selectedIdx: 0,
	}

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.presetPicker.selectedIdx != 0 {
		t.Errorf("expected selectedIdx=0, got %d", m.presetPicker.selectedIdx)
	}
}

func TestMood_DownDoesNotExceedPresets(t *testing.T) {
	model := newJournalModel()
	model.presetPicker = presetPickerState{
		active:      true,
		pickerMode:  true,
		kind:        pickerMood,
		input:       textinput.New(),
		items:       moodPresets,
		selectedIdx: len(moodPresets) - 1,
	}

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.presetPicker.selectedIdx != len(moodPresets)-1 {
		t.Errorf("expected selectedIdx=%d, got %d", len(moodPresets)-1, m.presetPicker.selectedIdx)
	}
}

func TestMood_EnterSubmitsSelectedPreset(t *testing.T) {
	model := newJournalModel()
	model.presetPicker = presetPickerState{
		active:      true,
		pickerMode:  true,
		kind:        pickerMood,
		input:       textinput.New(),
		items:       moodPresets,
		selectedIdx: 2,
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if m.presetPicker.active {
		t.Error("expected preset picker to be inactive after Enter")
	}
	if cmd == nil {
		t.Error("expected a command to be returned")
	}
}

func TestMood_EnterSubmitsCustomInput(t *testing.T) {
	model := newJournalModel()
	input := textinput.New()
	input.SetValue("energetic")
	model.presetPicker = presetPickerState{
		active:      true,
		pickerMode:  true,
		kind:        pickerMood,
		input:       input,
		items:       moodPresets,
		selectedIdx: 0,
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if m.presetPicker.active {
		t.Error("expected preset picker to be inactive after Enter")
	}
	if cmd == nil {
		t.Error("expected a command to be returned")
	}
}

func TestMood_EnterNoOpOnEmptyInput(t *testing.T) {
	model := newJournalModel()
	model.presetPicker = presetPickerState{
		active:     true,
		pickerMode: false,
		kind:       pickerMood,
		input:      textinput.New(),
		items:      nil,
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if m.presetPicker.active {
		t.Error("expected preset picker to be inactive after Enter")
	}
	if cmd != nil {
		t.Error("expected no command when input is empty and no presets")
	}
}

func TestMood_PresetsContainExpectedValues(t *testing.T) {
	expected := []string{"happy", "neutral", "sad", "frustrated", "tired", "sick", "anxious", "grateful"}
	if len(moodPresets) != len(expected) {
		t.Fatalf("expected %d mood presets, got %d", len(expected), len(moodPresets))
	}
	for i, preset := range moodPresets {
		if preset != expected[i] {
			t.Errorf("expected preset[%d]=%q, got %q", i, expected[i], preset)
		}
	}
}

func TestMood_RenderShowsPresets(t *testing.T) {
	model := newJournalModel()
	model.presetPicker = presetPickerState{
		active:      true,
		pickerMode:  true,
		kind:        pickerMood,
		input:       textinput.New(),
		items:       moodPresets,
		selectedIdx: 0,
		title:       "Set mood:",
		pickerLabel: "Mood presets:",
	}

	output := model.renderPresetPicker()
	if output == "" {
		t.Error("expected non-empty render output")
	}
}

// --- Weather tests ---

func TestWeather_WKeyActivatesWeatherMode(t *testing.T) {
	model := newJournalModel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'W'}}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if !m.presetPicker.active {
		t.Error("expected preset picker to be active after pressing 'W'")
	}
	if m.presetPicker.kind != pickerWeather {
		t.Error("expected picker kind to be pickerWeather")
	}
	if !m.presetPicker.pickerMode {
		t.Error("expected pickerMode to be true")
	}
}

func TestWeather_EscCancelsWeatherMode(t *testing.T) {
	model := newJournalModel()
	model.presetPicker = presetPickerState{
		active:     true,
		pickerMode: true,
		kind:       pickerWeather,
		input:      textinput.New(),
		items:      weatherPresets,
	}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.presetPicker.active {
		t.Error("expected preset picker to be inactive after Esc")
	}
}

func TestWeather_UpDownNavigatesPresets(t *testing.T) {
	model := newJournalModel()
	model.presetPicker = presetPickerState{
		active:      true,
		pickerMode:  true,
		kind:        pickerWeather,
		input:       textinput.New(),
		items:       weatherPresets,
		selectedIdx: 0,
	}

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.presetPicker.selectedIdx != 1 {
		t.Errorf("expected selectedIdx=1 after Down, got %d", m.presetPicker.selectedIdx)
	}

	msg = tea.KeyMsg{Type: tea.KeyUp}
	updated, _ = m.Update(msg)
	m = updated.(Model)

	if m.presetPicker.selectedIdx != 0 {
		t.Errorf("expected selectedIdx=0 after Up, got %d", m.presetPicker.selectedIdx)
	}
}

func TestWeather_EnterSubmitsSelectedPreset(t *testing.T) {
	model := newJournalModel()
	model.presetPicker = presetPickerState{
		active:      true,
		pickerMode:  true,
		kind:        pickerWeather,
		input:       textinput.New(),
		items:       weatherPresets,
		selectedIdx: 1,
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if m.presetPicker.active {
		t.Error("expected preset picker to be inactive after Enter")
	}
	if cmd == nil {
		t.Error("expected a command to be returned")
	}
}

func TestWeather_PresetsContainExpectedValues(t *testing.T) {
	expected := []string{"sunny", "partly-cloudy", "cloudy", "rainy", "stormy", "snowy"}
	if len(weatherPresets) != len(expected) {
		t.Fatalf("expected %d weather presets, got %d", len(expected), len(weatherPresets))
	}
	for i, preset := range weatherPresets {
		if preset != expected[i] {
			t.Errorf("expected preset[%d]=%q, got %q", i, expected[i], preset)
		}
	}
}

func TestWeather_RenderShowsPresets(t *testing.T) {
	model := newJournalModel()
	model.presetPicker = presetPickerState{
		active:      true,
		pickerMode:  true,
		kind:        pickerWeather,
		input:       textinput.New(),
		items:       weatherPresets,
		selectedIdx: 0,
		title:       "Set weather:",
		pickerLabel: "Weather presets:",
	}

	output := model.renderPresetPicker()
	if output == "" {
		t.Error("expected non-empty render output")
	}
}

func TestMood_MKeyNoOpInSearchView(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeSearch
	input := textinput.New()
	input.Blur()
	model.searchView = searchViewState{
		input: input,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.presetPicker.active {
		t.Error("expected preset picker NOT to activate in search view")
	}
}

func TestWeather_WKeyNoOpInSearchView(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeSearch
	input := textinput.New()
	input.Blur()
	model.searchView = searchViewState{
		input: input,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'W'}}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.presetPicker.active {
		t.Error("expected preset picker NOT to activate in search view")
	}
}

func TestMergePresets_CaseInsensitiveDedup(t *testing.T) {
	defaults := []string{"happy", "sad"}
	history := []string{"Happy", "excited"}

	result := mergePresets(defaults, history)

	expected := []string{"happy", "sad", "excited"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d presets, got %d: %v", len(expected), len(result), result)
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("expected result[%d]=%q, got %q", i, v, result[i])
		}
	}
}

func TestMood_MKeyWorksInReviewView(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeReview

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if !m.presetPicker.active {
		t.Error("expected preset picker to be active in review view")
	}
}
