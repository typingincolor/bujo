package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	detectors := []ChangeDetector{
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
	detectors := []ChangeDetector{
		&mockChangeDetector{lastModified: time.Time{}},
		&mockChangeDetector{lastModified: time.Time{}},
	}

	svc := NewChangeDetectionService(detectors)
	ctx := context.Background()

	lastModified, err := svc.GetLastModified(ctx)

	require.NoError(t, err)
	assert.True(t, lastModified.IsZero(), "should return zero time when all detectors return zero")
}

func TestChangeDetectionService_HasChangedSince_DetectsNewChanges(t *testing.T) {
	now := time.Now()
	since := now.Add(-1 * time.Hour)

	detectors := []ChangeDetector{
		&mockChangeDetector{lastModified: now},
	}

	svc := NewChangeDetectionService(detectors)
	ctx := context.Background()

	changed, err := svc.HasChangedSince(ctx, since)

	require.NoError(t, err)
	assert.True(t, changed, "should detect changes when lastModified is after since")
}

func TestChangeDetectionService_HasChangedSince_NoChanges(t *testing.T) {
	older := time.Now().Add(-2 * time.Hour)
	since := time.Now().Add(-1 * time.Hour)

	detectors := []ChangeDetector{
		&mockChangeDetector{lastModified: older},
	}

	svc := NewChangeDetectionService(detectors)
	ctx := context.Background()

	changed, err := svc.HasChangedSince(ctx, since)

	require.NoError(t, err)
	assert.False(t, changed, "should not detect changes when lastModified is before since")
}
