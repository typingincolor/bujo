package tui

import "github.com/typingincolor/bujo/internal/domain"

func CanCancel(entry domain.Entry) bool {
	return entry.CanCancel()
}

func CanUncancel(entry domain.Entry) bool {
	return entry.CanUncancel()
}

func CanCycleType(entry domain.Entry) bool {
	return entry.CanCycleType()
}

func CanEdit(entry domain.Entry) bool {
	return entry.CanEdit()
}

func CanMigrate(entry domain.Entry) bool {
	return entry.CanMigrate()
}

func CanAnswer(entry domain.Entry) bool {
	return entry.CanAnswer()
}

func CanAddChild(entry domain.Entry) bool {
	return entry.CanAddChild()
}

func CanMoveToList(entry domain.Entry) bool {
	return entry.CanMoveToList()
}

func CanMoveToRoot(entry domain.Entry) bool {
	return entry.CanMoveToRoot()
}

func CanCyclePriority(entry domain.Entry) bool {
	return entry.CanCyclePriority()
}

func CanDelete(entry domain.Entry) bool {
	return entry.CanDelete()
}

func UpdateKeyMapForEntry(km *KeyMap, entry domain.Entry) {
	km.CancelEntry.SetEnabled(CanCancel(entry))
	km.UncancelEntry.SetEnabled(CanUncancel(entry))
	km.Edit.SetEnabled(CanEdit(entry))
	km.Retype.SetEnabled(CanCycleType(entry))
	km.AddChild.SetEnabled(CanAddChild(entry))
	km.Migrate.SetEnabled(CanMigrate(entry))
	km.MoveToList.SetEnabled(CanMoveToList(entry))
	km.Answer.SetEnabled(CanAnswer(entry))
}

func ResetKeyMapEnabled(km *KeyMap) {
	km.CancelEntry.SetEnabled(true)
	km.UncancelEntry.SetEnabled(true)
	km.Edit.SetEnabled(true)
	km.Retype.SetEnabled(true)
	km.AddChild.SetEnabled(true)
	km.Migrate.SetEnabled(true)
	km.MoveToList.SetEnabled(true)
	km.Answer.SetEnabled(true)
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
