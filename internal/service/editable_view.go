package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type EditableViewService struct {
	entryRepo        domain.EntryRepository
	entryToListMover domain.EntryToListMover
	listRepo         domain.ListRepository
	tagRepo          domain.TagRepository
	mentionRepo      domain.MentionRepository
}

func NewEditableViewService(entryRepo domain.EntryRepository, entryToListMover domain.EntryToListMover, listRepo domain.ListRepository, tagRepo domain.TagRepository, mentionRepo domain.MentionRepository) *EditableViewService {
	return &EditableViewService{
		entryRepo:        entryRepo,
		entryToListMover: entryToListMover,
		listRepo:         listRepo,
		tagRepo:          tagRepo,
		mentionRepo:      mentionRepo,
	}
}

func (s *EditableViewService) GetEditableDocument(ctx context.Context, date time.Time) (string, error) {
	entries, err := s.entryRepo.GetByDate(ctx, date)
	if err != nil {
		return "", err
	}

	return domain.Serialize(entries), nil
}

type ValidationResult struct {
	IsValid     bool
	Errors      []domain.ParseError
	ParsedLines []domain.ParsedLine
}

func (s *EditableViewService) ValidateDocument(doc string) ValidationResult {
	parser := domain.NewEditableDocumentParser()
	parsed, err := parser.Parse(doc)
	if err != nil {
		return ValidationResult{
			IsValid: false,
			Errors:  []domain.ParseError{{LineNumber: 0, Message: err.Error()}},
		}
	}

	errors := make([]domain.ParseError, 0)

	var parentStack []int
	for _, line := range parsed.Lines {
		if strings.TrimSpace(line.Raw) == "" {
			continue
		}

		if !line.IsValid && !line.IsHeader {
			errors = append(errors, domain.ParseError{
				LineNumber: line.LineNumber,
				Message:    line.ErrorMessage,
			})
		}

		if line.IsValid && !line.IsHeader {
			if line.Depth > 0 && len(parentStack) == 0 {
				errors = append(errors, domain.ParseError{
					LineNumber: line.LineNumber,
					Message:    "Orphan child: no parent at depth 0",
				})
			}

			for len(parentStack) > line.Depth {
				parentStack = parentStack[:len(parentStack)-1]
			}

			if line.Depth >= len(parentStack) {
				parentStack = append(parentStack, line.LineNumber)
			} else {
				parentStack[line.Depth] = line.LineNumber
				parentStack = parentStack[:line.Depth+1]
			}
		}
	}

	validLines := make([]domain.ParsedLine, 0)
	for _, line := range parsed.Lines {
		if strings.TrimSpace(line.Raw) != "" {
			validLines = append(validLines, line)
		}
	}

	return ValidationResult{
		IsValid:     len(errors) == 0,
		Errors:      errors,
		ParsedLines: validLines,
	}
}

type ApplyChangesResult struct {
	Inserted int
	Deleted  int
}

func (s *EditableViewService) ApplyChanges(ctx context.Context, doc string, date time.Time) (*ApplyChangesResult, error) {
	validation := s.ValidateDocument(doc)
	if !validation.IsValid {
		return nil, fmt.Errorf("validation failed: %s", validation.Errors[0].Message)
	}

	existing, err := s.entryRepo.GetByDate(ctx, date)
	if err != nil {
		return nil, err
	}
	deletedCount := len(existing)

	err = s.entryRepo.DeleteByDate(ctx, date)
	if err != nil {
		return nil, err
	}

	result := &ApplyChangesResult{Deleted: deletedCount}

	type insertedEntry struct {
		id       int64
		symbol   domain.EntryType
		depth    int
		parentID *int64
	}

	var depthStack []int64
	var inserted []insertedEntry
	for _, line := range validation.ParsedLines {
		if !line.IsValid || line.IsHeader {
			continue
		}

		entry := domain.Entry{
			Type:          line.Symbol,
			Content:       line.Content,
			Priority:      line.Priority,
			Depth:         line.Depth,
			ScheduledDate: &date,
			CreatedAt:     time.Now(),
		}

		if line.Depth > 0 && len(depthStack) > line.Depth-1 {
			parentID := depthStack[line.Depth-1]
			entry.ParentID = &parentID
		}

		rowID, err := s.entryRepo.Insert(ctx, entry)
		if err != nil {
			return nil, err
		}

		if s.tagRepo != nil {
			tags := domain.ExtractTags(line.Content)
			if len(tags) > 0 {
				if err := s.tagRepo.InsertEntryTags(ctx, rowID, tags); err != nil {
					return nil, err
				}
			}
		}

		if s.mentionRepo != nil {
			mentions := domain.ExtractMentions(line.Content)
			if len(mentions) > 0 {
				if err := s.mentionRepo.InsertEntryMentions(ctx, rowID, mentions); err != nil {
					return nil, err
				}
			}
		}

		inserted = append(inserted, insertedEntry{
			id:       rowID,
			symbol:   line.Symbol,
			depth:    line.Depth,
			parentID: entry.ParentID,
		})

		if line.Depth >= len(depthStack) {
			depthStack = append(depthStack, rowID)
		} else {
			depthStack[line.Depth] = rowID
			depthStack = depthStack[:line.Depth+1]
		}

		result.Inserted++
	}

	questionsWithChildren := make(map[int64]bool)
	for _, ie := range inserted {
		if ie.parentID != nil {
			questionsWithChildren[*ie.parentID] = true
		}
	}

	for _, ie := range inserted {
		if ie.symbol == domain.EntryTypeQuestion && questionsWithChildren[ie.id] {
			entry, err := s.entryRepo.GetByID(ctx, ie.id)
			if err != nil {
				return nil, err
			}
			entry.Type = domain.EntryTypeAnswered
			if err := s.entryRepo.Update(ctx, *entry); err != nil {
				return nil, err
			}
		}

		if ie.parentID != nil {
			for _, parent := range inserted {
				if parent.id == *ie.parentID && parent.symbol == domain.EntryTypeQuestion {
					entry, err := s.entryRepo.GetByID(ctx, ie.id)
					if err != nil {
						return nil, err
					}
					entry.Type = domain.EntryTypeAnswer
					if err := s.entryRepo.Update(ctx, *entry); err != nil {
						return nil, err
					}
					break
				}
			}
		}
	}

	return result, nil
}

type ApplyActions struct {
	MigrateDate *time.Time
	ListID      *int64
}

func (s *EditableViewService) ApplyChangesWithActions(ctx context.Context, doc string, date time.Time, actions ApplyActions) (*ApplyChangesResult, error) {
	result, err := s.ApplyChanges(ctx, doc, date)
	if err != nil {
		return nil, err
	}

	entries, err := s.entryRepo.GetByDate(ctx, date)
	if err != nil {
		return nil, err
	}

	if actions.MigrateDate != nil {
		for _, entry := range entries {
			if entry.Type != domain.EntryTypeMigrated {
				continue
			}

			tree, err := s.entryRepo.GetWithChildren(ctx, entry.ID)
			if err != nil {
				return nil, err
			}

			descendants := tree[1:]

			origTypes := make(map[int64]domain.EntryType, len(descendants))
			for _, d := range descendants {
				origTypes[d.ID] = d.Type
			}

			for _, d := range descendants {
				d.Type = domain.EntryTypeMigrated
				if err := s.entryRepo.Update(ctx, d); err != nil {
					return nil, err
				}
			}

			newParentID, err := s.entryRepo.Insert(ctx, domain.Entry{
				Type:          domain.EntryTypeTask,
				Content:       entry.Content,
				Priority:      entry.Priority,
				ScheduledDate: actions.MigrateDate,
				CreatedAt:     time.Now(),
			})
			if err != nil {
				return nil, err
			}

			idMap := map[int64]int64{entry.ID: newParentID}
			for _, d := range descendants {
				var newParent int64
				if d.ParentID != nil {
					newParent = idMap[*d.ParentID]
				}
				newChildID, err := s.entryRepo.Insert(ctx, domain.Entry{
					Type:          origTypes[d.ID],
					Content:       d.Content,
					Priority:      d.Priority,
					Depth:         d.Depth,
					ParentID:      &newParent,
					ScheduledDate: actions.MigrateDate,
					CreatedAt:     time.Now(),
				})
				if err != nil {
					return nil, err
				}
				idMap[d.ID] = newChildID
			}
		}
	}

	if actions.ListID != nil && s.listRepo != nil && s.entryToListMover != nil {
		list, err := s.listRepo.GetByID(ctx, *actions.ListID)
		if err != nil {
			return nil, fmt.Errorf("list not found: %w", err)
		}
		for _, entry := range entries {
			if entry.Type == domain.EntryTypeMovedToList {
				if err := s.entryToListMover.MoveEntryToList(ctx, entry, list.EntityID); err != nil {
					return nil, err
				}
			}
		}
	}

	return result, nil
}
