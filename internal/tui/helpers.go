package tui

import "github.com/typingincolor/bujo/internal/domain"

func countEntryDescendants(entryID int64, parentMap map[int64][]domain.Entry) int {
	children := parentMap[entryID]
	count := len(children)
	for _, child := range children {
		count += countEntryDescendants(child.ID, parentMap)
	}
	return count
}
