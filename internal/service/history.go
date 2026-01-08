package service

import (
	"context"
	"fmt"

	"github.com/typingincolor/bujo/internal/domain"
)

type HistoryService struct {
	listItemRepo domain.ListItemRepository
}

func NewHistoryService(listItemRepo domain.ListItemRepository) *HistoryService {
	return &HistoryService{
		listItemRepo: listItemRepo,
	}
}

func (s *HistoryService) GetItemHistory(ctx context.Context, entityID domain.EntityID) ([]domain.ListItem, error) {
	return s.listItemRepo.GetHistory(ctx, entityID)
}

func (s *HistoryService) GetItemAtVersion(ctx context.Context, entityID domain.EntityID, version int) (*domain.ListItem, error) {
	return s.listItemRepo.GetAtVersion(ctx, entityID, version)
}

func (s *HistoryService) RestoreItem(ctx context.Context, entityID domain.EntityID, version int) error {
	oldVersion, err := s.listItemRepo.GetAtVersion(ctx, entityID, version)
	if err != nil {
		return err
	}
	if oldVersion == nil {
		return fmt.Errorf("version not found: %d", version)
	}

	current, err := s.listItemRepo.GetByEntityID(ctx, entityID)
	if err != nil {
		return err
	}
	if current == nil {
		return fmt.Errorf("entity not found: %s", entityID)
	}

	current.Content = oldVersion.Content
	current.Type = oldVersion.Type
	return s.listItemRepo.Update(ctx, *current)
}
