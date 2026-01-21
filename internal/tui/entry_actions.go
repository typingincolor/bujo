package tui

import "github.com/typingincolor/bujo/internal/domain"

func UpdateKeyMapForEntry(km *KeyMap, entry domain.Entry) {
	km.CancelEntry.SetEnabled(entry.CanCancel())
	km.UncancelEntry.SetEnabled(entry.CanUncancel())
	km.Edit.SetEnabled(entry.CanEdit())
	km.Retype.SetEnabled(entry.CanCycleType())
	km.AddChild.SetEnabled(entry.CanAddChild())
	km.Migrate.SetEnabled(entry.CanMigrate())
	km.MoveToList.SetEnabled(entry.CanMoveToList())
	km.MoveToRoot.SetEnabled(entry.CanMoveToRoot())
	km.Answer.SetEnabled(entry.CanAnswer())
	km.Priority.SetEnabled(entry.CanCyclePriority())
	km.Delete.SetEnabled(entry.CanDelete())
}

func ResetKeyMapEnabled(km *KeyMap) {
	km.CancelEntry.SetEnabled(true)
	km.UncancelEntry.SetEnabled(true)
	km.Edit.SetEnabled(true)
	km.Retype.SetEnabled(true)
	km.AddChild.SetEnabled(true)
	km.Migrate.SetEnabled(true)
	km.MoveToList.SetEnabled(true)
	km.MoveToRoot.SetEnabled(true)
	km.Answer.SetEnabled(true)
	km.Priority.SetEnabled(true)
	km.Delete.SetEnabled(true)
}

func (m Model) syncKeyMapToSelection() Model {
	if len(m.entries) > 0 && m.selectedIdx < len(m.entries) {
		entry := m.entries[m.selectedIdx].Entry
		UpdateKeyMapForEntry(&m.keyMap, entry)
	} else {
		ResetKeyMapEnabled(&m.keyMap)
	}
	return m
}
