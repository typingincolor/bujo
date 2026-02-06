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

	if !m.setMoodMode.active {
		t.Error("expected setMoodMode to be active after pressing 'm'")
	}
	if !m.setMoodMode.pickerMode {
		t.Error("expected pickerMode to be true")
	}
}

func TestMood_EscCancelsMoodMode(t *testing.T) {
	model := newJournalModel()
	model.setMoodMode = setMoodState{
		active:     true,
		pickerMode: true,
		input:      textinput.New(),
		presets:    moodPresets,
	}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.setMoodMode.active {
		t.Error("expected setMoodMode to be inactive after Esc")
	}
}

func TestMood_UpDownNavigatesPresets(t *testing.T) {
	model := newJournalModel()
	model.setMoodMode = setMoodState{
		active:      true,
		pickerMode:  true,
		input:       textinput.New(),
		presets:     moodPresets,
		selectedIdx: 0,
	}

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.setMoodMode.selectedIdx != 1 {
		t.Errorf("expected selectedIdx=1 after Down, got %d", m.setMoodMode.selectedIdx)
	}

	msg = tea.KeyMsg{Type: tea.KeyUp}
	updated, _ = m.Update(msg)
	m = updated.(Model)

	if m.setMoodMode.selectedIdx != 0 {
		t.Errorf("expected selectedIdx=0 after Up, got %d", m.setMoodMode.selectedIdx)
	}
}

func TestMood_UpDoesNotGoBelowZero(t *testing.T) {
	model := newJournalModel()
	model.setMoodMode = setMoodState{
		active:      true,
		pickerMode:  true,
		input:       textinput.New(),
		presets:     moodPresets,
		selectedIdx: 0,
	}

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.setMoodMode.selectedIdx != 0 {
		t.Errorf("expected selectedIdx=0, got %d", m.setMoodMode.selectedIdx)
	}
}

func TestMood_DownDoesNotExceedPresets(t *testing.T) {
	model := newJournalModel()
	model.setMoodMode = setMoodState{
		active:      true,
		pickerMode:  true,
		input:       textinput.New(),
		presets:     moodPresets,
		selectedIdx: len(moodPresets) - 1,
	}

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.setMoodMode.selectedIdx != len(moodPresets)-1 {
		t.Errorf("expected selectedIdx=%d, got %d", len(moodPresets)-1, m.setMoodMode.selectedIdx)
	}
}

func TestMood_EnterSubmitsSelectedPreset(t *testing.T) {
	model := newJournalModel()
	model.setMoodMode = setMoodState{
		active:      true,
		pickerMode:  true,
		input:       textinput.New(),
		presets:     moodPresets,
		selectedIdx: 2,
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if m.setMoodMode.active {
		t.Error("expected setMoodMode to be inactive after Enter")
	}
	if cmd == nil {
		t.Error("expected a command to be returned")
	}
}

func TestMood_EnterSubmitsCustomInput(t *testing.T) {
	model := newJournalModel()
	input := textinput.New()
	input.SetValue("energetic")
	model.setMoodMode = setMoodState{
		active:      true,
		pickerMode:  true,
		input:       input,
		presets:     moodPresets,
		selectedIdx: 0,
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if m.setMoodMode.active {
		t.Error("expected setMoodMode to be inactive after Enter")
	}
	if cmd == nil {
		t.Error("expected a command to be returned")
	}
}

func TestMood_EnterNoOpOnEmptyInput(t *testing.T) {
	model := newJournalModel()
	model.setMoodMode = setMoodState{
		active:     true,
		pickerMode: false,
		input:      textinput.New(),
		presets:    nil,
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if m.setMoodMode.active {
		t.Error("expected setMoodMode to be inactive after Enter")
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
	model.setMoodMode = setMoodState{
		active:      true,
		pickerMode:  true,
		input:       textinput.New(),
		presets:     moodPresets,
		selectedIdx: 0,
	}

	output := model.renderSetMoodInput()
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

	if !m.setWeatherMode.active {
		t.Error("expected setWeatherMode to be active after pressing 'W'")
	}
	if !m.setWeatherMode.pickerMode {
		t.Error("expected pickerMode to be true")
	}
}

func TestWeather_EscCancelsWeatherMode(t *testing.T) {
	model := newJournalModel()
	model.setWeatherMode = setWeatherState{
		active:     true,
		pickerMode: true,
		input:      textinput.New(),
		presets:    weatherPresets,
	}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.setWeatherMode.active {
		t.Error("expected setWeatherMode to be inactive after Esc")
	}
}

func TestWeather_UpDownNavigatesPresets(t *testing.T) {
	model := newJournalModel()
	model.setWeatherMode = setWeatherState{
		active:      true,
		pickerMode:  true,
		input:       textinput.New(),
		presets:     weatherPresets,
		selectedIdx: 0,
	}

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := model.Update(msg)
	m := updated.(Model)

	if m.setWeatherMode.selectedIdx != 1 {
		t.Errorf("expected selectedIdx=1 after Down, got %d", m.setWeatherMode.selectedIdx)
	}

	msg = tea.KeyMsg{Type: tea.KeyUp}
	updated, _ = m.Update(msg)
	m = updated.(Model)

	if m.setWeatherMode.selectedIdx != 0 {
		t.Errorf("expected selectedIdx=0 after Up, got %d", m.setWeatherMode.selectedIdx)
	}
}

func TestWeather_EnterSubmitsSelectedPreset(t *testing.T) {
	model := newJournalModel()
	model.setWeatherMode = setWeatherState{
		active:      true,
		pickerMode:  true,
		input:       textinput.New(),
		presets:     weatherPresets,
		selectedIdx: 1,
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := model.Update(msg)
	m := updated.(Model)

	if m.setWeatherMode.active {
		t.Error("expected setWeatherMode to be inactive after Enter")
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
	model.setWeatherMode = setWeatherState{
		active:      true,
		pickerMode:  true,
		input:       textinput.New(),
		presets:     weatherPresets,
		selectedIdx: 0,
	}

	output := model.renderSetWeatherInput()
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

	if m.setMoodMode.active {
		t.Error("expected setMoodMode NOT to activate in search view")
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

	if m.setWeatherMode.active {
		t.Error("expected setWeatherMode NOT to activate in search view")
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

	if !m.setMoodMode.active {
		t.Error("expected setMoodMode to be active in review view")
	}
}
