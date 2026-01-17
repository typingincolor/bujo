package wails

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/app"
)

func TestNewApp_AcceptsServices(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)

	assert.NotNil(t, wailsApp)
	assert.NotNil(t, wailsApp.services)
}

func TestApp_Startup_StoresContext(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	assert.NotNil(t, wailsApp.ctx)
}

func TestApp_GetAgenda_ReturnsMultiDayAgenda(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	agenda, err := wailsApp.GetAgenda(today, today.AddDate(0, 0, 7))

	require.NoError(t, err)
	assert.NotNil(t, agenda)
	assert.NotNil(t, agenda.Days)
}

func TestApp_GetHabits_ReturnsTrackerStatus(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	status, err := wailsApp.GetHabits(7)

	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.NotNil(t, status.Habits)
}
