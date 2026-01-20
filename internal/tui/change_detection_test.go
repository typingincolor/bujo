package tui

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockChangeDetectionService struct {
	lastModified time.Time
	err          error
}

func (m *mockChangeDetectionService) GetLastModified(ctx context.Context) (time.Time, error) {
	return m.lastModified, m.err
}

func TestModel_CheckChangesMsg_TriggersReloadWhenDataChanged(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Second)

	mockSvc := &mockChangeDetectionService{lastModified: later}

	model := NewWithConfig(Config{
		ChangeDetection: mockSvc,
	})
	model.lastCheckedModified = now

	msg := checkChangesMsg{}
	m, cmd := model.Update(msg)
	updated := m.(Model)

	assert.Equal(t, later.Unix(), updated.lastCheckedModified.Unix(), "should update lastCheckedModified")
	require.NotNil(t, cmd, "should return a command to reload data")
}

func TestModel_CheckChangesMsg_NoReloadWhenNoChanges(t *testing.T) {
	now := time.Now()

	mockSvc := &mockChangeDetectionService{lastModified: now}

	model := NewWithConfig(Config{
		ChangeDetection: mockSvc,
	})
	model.lastCheckedModified = now

	msg := checkChangesMsg{}
	m, cmd := model.Update(msg)
	updated := m.(Model)

	assert.Equal(t, now.Unix(), updated.lastCheckedModified.Unix())
	// Should only return the next tick command, not a reload command
	require.NotNil(t, cmd, "should return a tick command")
	// Execute to verify it produces a message (tick)
	result := cmd()
	assert.NotNil(t, result, "tick command should produce a message")
}

func TestModel_Init_StartsChangeDetectionTicker(t *testing.T) {
	mockSvc := &mockChangeDetectionService{lastModified: time.Now()}

	model := NewWithConfig(Config{
		ChangeDetection: mockSvc,
	})

	cmd := model.Init()
	require.NotNil(t, cmd, "Init should return a command")
}

func TestModel_CheckChangesMsg_HandlesNilService(t *testing.T) {
	model := NewWithConfig(Config{
		ChangeDetection: nil,
	})

	msg := checkChangesMsg{}
	m, cmd := model.Update(msg)
	_ = m.(Model)

	// Should not panic and should return a tick command
	if cmd != nil {
		// Just verify it doesn't panic
		_ = cmd()
	}
}
