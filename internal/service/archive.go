package service

import (
	"context"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type ArchiveService struct {
	listItemRepo domain.ListItemRepository
}

func NewArchiveService(listItemRepo domain.ListItemRepository) *ArchiveService {
	return &ArchiveService{
		listItemRepo: listItemRepo,
	}
}

func (s *ArchiveService) GetArchivableCount(ctx context.Context, olderThan time.Time) (int, error) {
	return s.listItemRepo.CountArchivable(ctx, olderThan)
}

func (s *ArchiveService) Archive(ctx context.Context, olderThan time.Time) (int, error) {
	return s.listItemRepo.DeleteArchivable(ctx, olderThan)
}

func (s *ArchiveService) DryRun(ctx context.Context, olderThan time.Time) (int, error) {
	return s.listItemRepo.CountArchivable(ctx, olderThan)
}
