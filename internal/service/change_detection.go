package service

import (
	"context"
	"time"
)

type ChangeDetector interface {
	GetLastModified(ctx context.Context) (time.Time, error)
}

type ChangeDetectionService struct {
	detectors []ChangeDetector
}

func NewChangeDetectionService(detectors []ChangeDetector) *ChangeDetectionService {
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

func (s *ChangeDetectionService) HasChangedSince(ctx context.Context, since time.Time) (bool, error) {
	lastModified, err := s.GetLastModified(ctx)
	if err != nil {
		return false, err
	}

	return lastModified.After(since), nil
}
