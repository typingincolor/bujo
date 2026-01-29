package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type EditableViewService struct {
	entryRepo   domain.EntryRepository
	bujoService *BujoService
}

func NewEditableViewService(entryRepo domain.EntryRepository, bujoService *BujoService) *EditableViewService {
	return &EditableViewService{
		entryRepo:   entryRepo,
		bujoService: bujoService,
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
	dateParser := func(dateStr string) (time.Time, error) {
		return time.Parse("2006-01-02", dateStr)
	}
	parser := domain.NewEditableDocumentParser(dateParser)
	parsed, err := parser.Parse(doc, nil)
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
	Updated  int
	Deleted  int
	Migrated int
}

func (s *EditableViewService) ApplyChanges(ctx context.Context, doc string, date time.Time, pendingDeletes []domain.EntityID) (*ApplyChangesResult, error) {
	original, err := s.entryRepo.GetByDate(ctx, date)
	if err != nil {
		return nil, err
	}

	dateParser := func(dateStr string) (time.Time, error) {
		return time.Parse("2006-01-02", dateStr)
	}
	parser := domain.NewEditableDocumentParser(dateParser)
	parsed, err := parser.Parse(doc, original)
	if err != nil {
		return nil, err
	}

	parsed.PendingDeletes = pendingDeletes

	validation := s.validateParsed(parsed)
	if !validation.IsValid {
		return nil, fmt.Errorf("validation failed: %s", validation.Errors[0].Message)
	}

	changeset := domain.ComputeDiff(original, parsed)

	if len(changeset.Errors) > 0 {
		return nil, fmt.Errorf("diff error: %s", changeset.Errors[0].Message)
	}

	result := &ApplyChangesResult{}

	entityIDToRowID := make(map[domain.EntityID]int64)

	for _, op := range changeset.Operations {
		switch op.Type {
		case domain.DiffOpInsert:
			op.Entry.ScheduledDate = &date
			op.Entry.CreatedAt = time.Now()
			if op.Entry.ParentEntityID != nil {
				if rowID, ok := entityIDToRowID[*op.Entry.ParentEntityID]; ok {
					op.Entry.ParentID = &rowID
				} else {
					parent, err := s.entryRepo.GetByEntityID(ctx, *op.Entry.ParentEntityID)
					if err != nil {
						return nil, err
					}
					if parent != nil {
						op.Entry.ParentID = &parent.ID
					}
				}
			}
			rowID, err := s.entryRepo.Insert(ctx, op.Entry)
			if err != nil {
				return nil, err
			}
			if !op.Entry.EntityID.IsEmpty() {
				entityIDToRowID[op.Entry.EntityID] = rowID
			}
			result.Inserted++

		case domain.DiffOpUpdate:
			existing, err := s.entryRepo.GetByEntityID(ctx, *op.EntityID)
			if err != nil {
				return nil, err
			}
			if existing == nil {
				continue
			}
			existing.Content = op.Entry.Content
			existing.Type = op.Entry.Type
			existing.Priority = op.Entry.Priority
			err = s.entryRepo.Update(ctx, *existing)
			if err != nil {
				return nil, err
			}
			result.Updated++

		case domain.DiffOpDelete:
			existing, err := s.entryRepo.GetByEntityID(ctx, *op.EntityID)
			if err != nil {
				return nil, err
			}
			if existing == nil {
				continue
			}
			err = s.entryRepo.Delete(ctx, existing.ID)
			if err != nil {
				return nil, err
			}
			result.Deleted++

		case domain.DiffOpMigrate:
			existing, err := s.entryRepo.GetByEntityID(ctx, *op.EntityID)
			if err != nil {
				return nil, err
			}
			if existing == nil {
				continue
			}
			_, err = s.bujoService.MigrateEntry(ctx, existing.ID, *op.MigrateDate)
			if err != nil {
				return nil, err
			}
			result.Migrated++

		case domain.DiffOpReparent:
			existing, err := s.entryRepo.GetByEntityID(ctx, *op.EntityID)
			if err != nil {
				return nil, err
			}
			if existing == nil {
				continue
			}
			if op.NewParentID != nil {
				parent, err := s.entryRepo.GetByEntityID(ctx, *op.NewParentID)
				if err != nil {
					return nil, err
				}
				if parent != nil {
					existing.ParentID = &parent.ID
					existing.Depth = parent.Depth + 1
				}
			} else {
				existing.ParentID = nil
				existing.Depth = 0
			}
			err = s.entryRepo.Update(ctx, *existing)
			if err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func (s *EditableViewService) validateParsed(parsed *domain.EditableDocument) ValidationResult {
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
