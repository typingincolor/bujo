package service

import (
	"context"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type ChangeDetectionService struct {
	detectors []domain.ChangeDetector
}

func NewChangeDetectionService(detectors []domain.ChangeDetector) *ChangeDetectionService {
	return &ChangeDetectionService{detectors: detectors}
}

func (s *ChangeDetectionService) GetLastModified(ctx context.Context) (time.Time, error) {
	var latest time.Time

	for _, detector := range s.detectors {
		t, err := detector.GetLastModified(ctx)
		if err != nil {
			return time.Time{}, err
		}
		if t.After(latest) {
			latest = t
		}
	}

	return latest, nil
}
