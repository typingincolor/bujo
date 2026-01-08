package service

import (
	"context"
	"fmt"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type ListRepository interface {
	Create(ctx context.Context, name string) (*domain.List, error)
	GetByID(ctx context.Context, id int64) (*domain.List, error)
	GetByName(ctx context.Context, name string) (*domain.List, error)
	GetAll(ctx context.Context) ([]domain.List, error)
	Rename(ctx context.Context, id int64, newName string) error
	Delete(ctx context.Context, id int64) error
	GetItemCount(ctx context.Context, listID int64) (int, error)
	GetDoneCount(ctx context.Context, listID int64) (int, error)
}

type ListEntryRepository interface {
	Insert(ctx context.Context, entry domain.Entry) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Entry, error)
	GetByListID(ctx context.Context, listID int64) ([]domain.Entry, error)
	Update(ctx context.Context, entry domain.Entry) error
	Delete(ctx context.Context, id int64) error
	DeleteWithChildren(ctx context.Context, id int64) error
}

type ListService struct {
	listRepo  ListRepository
	entryRepo ListEntryRepository
}

func NewListService(listRepo ListRepository, entryRepo ListEntryRepository) *ListService {
	return &ListService{
		listRepo:  listRepo,
		entryRepo: entryRepo,
	}
}

func (s *ListService) getListByID(ctx context.Context, id int64) (*domain.List, error) {
	list, err := s.listRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return nil, fmt.Errorf("list not found: %d", id)
	}
	return list, nil
}

func (s *ListService) getEntryByID(ctx context.Context, id int64) (*domain.Entry, error) {
	entry, err := s.entryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, fmt.Errorf("entry not found: %d", id)
	}
	return entry, nil
}

func (s *ListService) CreateList(ctx context.Context, name string) (*domain.List, error) {
	return s.listRepo.Create(ctx, name)
}

func (s *ListService) GetListByID(ctx context.Context, id int64) (*domain.List, error) {
	return s.getListByID(ctx, id)
}

func (s *ListService) GetListByName(ctx context.Context, name string) (*domain.List, error) {
	list, err := s.listRepo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return nil, fmt.Errorf("list not found: %s", name)
	}
	return list, nil
}

func (s *ListService) GetAllLists(ctx context.Context) ([]domain.List, error) {
	return s.listRepo.GetAll(ctx)
}

func (s *ListService) RenameList(ctx context.Context, id int64, newName string) error {
	if _, err := s.getListByID(ctx, id); err != nil {
		return err
	}
	return s.listRepo.Rename(ctx, id, newName)
}

func (s *ListService) DeleteList(ctx context.Context, id int64, force bool) error {
	if _, err := s.getListByID(ctx, id); err != nil {
		return err
	}

	count, err := s.listRepo.GetItemCount(ctx, id)
	if err != nil {
		return err
	}

	if count > 0 && !force {
		return fmt.Errorf("list has items (%d); use force to delete anyway", count)
	}

	if count > 0 {
		items, err := s.entryRepo.GetByListID(ctx, id)
		if err != nil {
			return err
		}
		for _, item := range items {
			if err := s.entryRepo.DeleteWithChildren(ctx, item.ID); err != nil {
				return err
			}
		}
	}

	return s.listRepo.Delete(ctx, id)
}

func (s *ListService) AddItem(ctx context.Context, listID int64, entryType domain.EntryType, content string) (int64, error) {
	if _, err := s.getListByID(ctx, listID); err != nil {
		return 0, err
	}

	entry := domain.Entry{
		Type:      entryType,
		Content:   content,
		ListID:    &listID,
		CreatedAt: time.Now(),
	}

	return s.entryRepo.Insert(ctx, entry)
}

func (s *ListService) GetListItems(ctx context.Context, listID int64) ([]domain.Entry, error) {
	return s.entryRepo.GetByListID(ctx, listID)
}

func (s *ListService) RemoveItem(ctx context.Context, entryID int64) error {
	entry, err := s.getEntryByID(ctx, entryID)
	if err != nil {
		return err
	}
	if entry.ListID == nil {
		return fmt.Errorf("entry %d is not a list item", entryID)
	}
	return s.entryRepo.Delete(ctx, entryID)
}

func (s *ListService) MarkDone(ctx context.Context, entryID int64) error {
	entry, err := s.getEntryByID(ctx, entryID)
	if err != nil {
		return err
	}
	if entry.ListID == nil {
		return fmt.Errorf("entry %d is not a list item", entryID)
	}

	entry.Type = domain.EntryTypeDone
	return s.entryRepo.Update(ctx, *entry)
}

func (s *ListService) MarkUndone(ctx context.Context, entryID int64) error {
	entry, err := s.getEntryByID(ctx, entryID)
	if err != nil {
		return err
	}
	if entry.ListID == nil {
		return fmt.Errorf("entry %d is not a list item", entryID)
	}

	entry.Type = domain.EntryTypeTask
	return s.entryRepo.Update(ctx, *entry)
}

func (s *ListService) MoveItem(ctx context.Context, entryID int64, targetListID int64) error {
	entry, err := s.getEntryByID(ctx, entryID)
	if err != nil {
		return err
	}
	if entry.ListID == nil {
		return fmt.Errorf("entry %d is not a list item", entryID)
	}

	if _, err := s.getListByID(ctx, targetListID); err != nil {
		return err
	}

	entry.ListID = &targetListID
	return s.entryRepo.Update(ctx, *entry)
}

type ListSummary struct {
	ID         int64
	Name       string
	TotalItems int
	DoneItems  int
}

func (s *ListService) GetListSummary(ctx context.Context, listID int64) (*ListSummary, error) {
	list, err := s.getListByID(ctx, listID)
	if err != nil {
		return nil, err
	}

	total, err := s.listRepo.GetItemCount(ctx, listID)
	if err != nil {
		return nil, err
	}

	done, err := s.listRepo.GetDoneCount(ctx, listID)
	if err != nil {
		return nil, err
	}

	return &ListSummary{
		ID:         list.ID,
		Name:       list.Name,
		TotalItems: total,
		DoneItems:  done,
	}, nil
}
