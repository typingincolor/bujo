package domain

func ComputeDiff(original []Entry, parsed *EditableDocument) *Changeset {
	changeset := &Changeset{
		Operations: make([]DiffOperation, 0),
		Errors:     make([]ParseError, 0),
	}

	originalByEntityID := make(map[EntityID]Entry)
	for _, entry := range original {
		originalByEntityID[entry.EntityID] = entry
	}

	pendingDeleteSet := make(map[EntityID]bool)
	for _, id := range parsed.PendingDeletes {
		pendingDeleteSet[id] = true
	}

	seenEntityIDs := make(map[EntityID]bool)

	var parentStack []EntityID

	for _, line := range parsed.Lines {
		if line.IsHeader || !line.IsValid {
			continue
		}

		if line.Depth > 0 && len(parentStack) == 0 {
			changeset.Errors = append(changeset.Errors, ParseError{
				LineNumber: line.LineNumber,
				Message:    "Orphan child: no parent at depth 0",
			})
			continue
		}

		for len(parentStack) > line.Depth {
			parentStack = parentStack[:len(parentStack)-1]
		}

		var currentParentID *EntityID
		if len(parentStack) > 0 && line.Depth > 0 {
			parent := parentStack[len(parentStack)-1]
			currentParentID = &parent
		}

		if line.MigrateTarget != nil && line.EntityID != nil {
			changeset.Operations = append(changeset.Operations, DiffOperation{
				Type:        DiffOpMigrate,
				EntityID:    line.EntityID,
				MigrateDate: line.MigrateTarget,
				LineNumber:  line.LineNumber,
			})
			seenEntityIDs[*line.EntityID] = true

			if line.EntityID != nil {
				parentStack = appendOrUpdateStack(parentStack, line.Depth, *line.EntityID)
			}
			continue
		}

		if line.EntityID == nil {
			newEntry := Entry{
				Type:     line.Symbol,
				Content:  line.Content,
				Priority: line.Priority,
				Depth:    line.Depth,
			}
			if currentParentID != nil {
				newEntry.ParentEntityID = currentParentID
			}

			changeset.Operations = append(changeset.Operations, DiffOperation{
				Type:       DiffOpInsert,
				Entry:      newEntry,
				LineNumber: line.LineNumber,
			})

			parentStack = appendOrUpdateStack(parentStack, line.Depth, NewEntityID())
			continue
		}

		seenEntityIDs[*line.EntityID] = true

		originalEntry, exists := originalByEntityID[*line.EntityID]
		if !exists {
			continue
		}

		parentChanged := false
		if currentParentID == nil && originalEntry.ParentEntityID != nil {
			parentChanged = true
		} else if currentParentID != nil && originalEntry.ParentEntityID == nil {
			parentChanged = true
		} else if currentParentID != nil && originalEntry.ParentEntityID != nil && *currentParentID != *originalEntry.ParentEntityID {
			parentChanged = true
		}

		if parentChanged {
			changeset.Operations = append(changeset.Operations, DiffOperation{
				Type:        DiffOpReparent,
				EntityID:    line.EntityID,
				NewParentID: currentParentID,
				LineNumber:  line.LineNumber,
			})
		}

		contentChanged := line.Content != originalEntry.Content
		typeChanged := line.Symbol != originalEntry.Type
		priorityChanged := line.Priority != originalEntry.Priority

		if contentChanged || typeChanged || priorityChanged {
			updatedEntry := Entry{
				EntityID: *line.EntityID,
				Type:     line.Symbol,
				Content:  line.Content,
				Priority: line.Priority,
				Depth:    line.Depth,
			}
			if currentParentID != nil {
				updatedEntry.ParentEntityID = currentParentID
			}

			changeset.Operations = append(changeset.Operations, DiffOperation{
				Type:       DiffOpUpdate,
				EntityID:   line.EntityID,
				Entry:      updatedEntry,
				LineNumber: line.LineNumber,
			})
		}

		parentStack = appendOrUpdateStack(parentStack, line.Depth, *line.EntityID)
	}

	for _, id := range parsed.PendingDeletes {
		if _, exists := originalByEntityID[id]; exists {
			changeset.Operations = append(changeset.Operations, DiffOperation{
				Type:     DiffOpDelete,
				EntityID: &id,
			})
		}
	}

	return changeset
}

func appendOrUpdateStack(stack []EntityID, depth int, id EntityID) []EntityID {
	if depth >= len(stack) {
		return append(stack, id)
	}
	stack[depth] = id
	return stack[:depth+1]
}
