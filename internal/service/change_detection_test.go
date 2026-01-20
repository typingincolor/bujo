package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

type mockChangeDetector struct {
	lastModified time.Time
	err          error
}

func (m *mockChangeDetector) GetLastModified(ctx context.Context) (time.Time, error) {
	return m.lastModified, m.err
}

func TestChangeDetectionService_GetLastModified_ReturnsLatestAcrossAllDetectors(t *testing.T) {
	now := time.Now()
	older := now.Add(-1 * time.Hour)
	oldest := now.Add(-2 * time.Hour)

	detectors := []domain.ChangeDetector{
		&mockChangeDetector{lastModified: oldest},
		&mockChangeDetector{lastModified: now},
		&mockChangeDetector{lastModified: older},
	}

	svc := NewChangeDetectionService(detectors)
	ctx := context.Background()

	lastModified, err := svc.GetLastModified(ctx)

	require.NoError(t, err)
	assert.Equal(t, now.Unix(), lastModified.Unix(), "should return the latest timestamp across all detectors")
}

func TestChangeDetectionService_GetLastModified_EmptyDetectors_ReturnsZeroTime(t *testing.T) {
	svc := NewChangeDetectionService(nil)
	ctx := context.Background()

	lastModified, err := svc.GetLastModified(ctx)

	require.NoError(t, err)
	assert.True(t, lastModified.IsZero(), "should return zero time when no detectors")
}

func TestChangeDetectionService_GetLastModified_AllZeroTimes_ReturnsZeroTime(t *testing.T) {
	detectors := []domain.ChangeDetector{
		&mockChangeDetector{lastModified: time.Time{}},
		&mockChangeDetector{lastModified: time.Time{}},
	}

	svc := NewChangeDetectionService(detectors)
	ctx := context.Background()

	lastModified, err := svc.GetLastModified(ctx)

	require.NoError(t, err)
	assert.True(t, lastModified.IsZero(), "should return zero time when all detectors return zero")
}
