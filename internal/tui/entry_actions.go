package tui

import "github.com/typingincolor/bujo/internal/domain"

var cycleableTypes = map[domain.EntryType]bool{
	domain.EntryTypeTask:     true,
	domain.EntryTypeNote:     true,
	domain.EntryTypeEvent:    true,
	domain.EntryTypeQuestion: true,
}

func CanCancel(entry domain.Entry) bool {
	return entry.Type != domain.EntryTypeCancelled
}

func CanUncancel(entry domain.Entry) bool {
	return entry.Type == domain.EntryTypeCancelled
}

func CanCycleType(entry domain.Entry) bool {
	return cycleableTypes[entry.Type]
}

func CanEdit(entry domain.Entry) bool {
	return entry.Type != domain.EntryTypeCancelled
}

func CanMigrate(entry domain.Entry) bool {
	return entry.Type == domain.EntryTypeTask
}

func CanAnswer(entry domain.Entry) bool {
	return entry.Type == domain.EntryTypeQuestion
}

func CanAddChild(entry domain.Entry) bool {
	return entry.Type != domain.EntryTypeQuestion
}

func CanMoveToList(entry domain.Entry) bool {
	return entry.Type == domain.EntryTypeTask
}

func CanMoveToRoot(entry domain.Entry) bool {
	return entry.ParentID != nil
}

func CanCyclePriority(_ domain.Entry) bool {
	return true
}

func CanDelete(_ domain.Entry) bool {
	return true
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
