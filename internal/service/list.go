package service

import (
	"context"
	"fmt"

	"github.com/typingincolor/bujo/internal/domain"
)

type ListRepository interface {
	Create(ctx context.Context, name string) (*domain.List, error)
	GetByID(ctx context.Context, id int64) (*domain.List, error)
	GetByName(ctx context.Context, name string) (*domain.List, error)
	GetByEntityID(ctx context.Context, entityID domain.EntityID) (*domain.List, error)
	GetAll(ctx context.Context) ([]domain.List, error)
	Rename(ctx context.Context, id int64, newName string) error
	Delete(ctx context.Context, id int64) error
	GetItemCount(ctx context.Context, listID int64) (int, error)
	GetDoneCount(ctx context.Context, listID int64) (int, error)
}

type ListService struct {
	listRepo     ListRepository
	listItemRepo domain.ListItemRepository
}

func NewListService(listRepo ListRepository, listItemRepo domain.ListItemRepository) *ListService {
	return &ListService{
		listRepo:     listRepo,
		listItemRepo: listItemRepo,
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

func (s *ListService) getItemByID(ctx context.Context, id int64) (*domain.ListItem, error) {
	item, err := s.listItemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("item not found: %d", id)
	}
	return item, nil
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
		items, err := s.listItemRepo.GetByListID(ctx, id)
		if err != nil {
			return err
		}
		for _, item := range items {
			if err := s.listItemRepo.Delete(ctx, item.RowID); err != nil {
				return err
			}
		}
	}

	return s.listRepo.Delete(ctx, id)
}

func (s *ListService) AddItem(ctx context.Context, listID int64, entryType domain.EntryType, content string) (int64, error) {
	list, err := s.getListByID(ctx, listID)
	if err != nil {
		return 0, err
	}

	itemType := domain.ListItemTypeTask
	if entryType == domain.EntryTypeDone {
		itemType = domain.ListItemTypeDone
	}

	item := domain.NewListItem(list.EntityID, itemType, content)
	return s.listItemRepo.Insert(ctx, item)
}

func (s *ListService) GetListItems(ctx context.Context, listID int64) ([]domain.ListItem, error) {
	return s.listItemRepo.GetByListID(ctx, listID)
}

func (s *ListService) RemoveItem(ctx context.Context, itemID int64) error {
	if _, err := s.getItemByID(ctx, itemID); err != nil {
		return err
	}
	return s.listItemRepo.Delete(ctx, itemID)
}

func (s *ListService) MarkDone(ctx context.Context, itemID int64) error {
	item, err := s.getItemByID(ctx, itemID)
	if err != nil {
		return err
	}

	item.Type = domain.ListItemTypeDone
	return s.listItemRepo.Update(ctx, *item)
}

func (s *ListService) MarkUndone(ctx context.Context, itemID int64) error {
	item, err := s.getItemByID(ctx, itemID)
	if err != nil {
		return err
	}

	item.Type = domain.ListItemTypeTask
	return s.listItemRepo.Update(ctx, *item)
}

func (s *ListService) MoveItem(ctx context.Context, itemID int64, targetListID int64) error {
	item, err := s.getItemByID(ctx, itemID)
	if err != nil {
		return err
	}

	targetList, err := s.getListByID(ctx, targetListID)
	if err != nil {
		return err
	}

	item.ListEntityID = targetList.EntityID
	return s.listItemRepo.Update(ctx, *item)
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
