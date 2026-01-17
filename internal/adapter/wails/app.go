package wails

import (
	"context"
	"time"

	"github.com/typingincolor/bujo/internal/app"
	"github.com/typingincolor/bujo/internal/service"
)

type App struct {
	ctx      context.Context
	services *app.Services
}

func NewApp(services *app.Services) *App {
	return &App{
		services: services,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Greet(name string) string {
	return "Hello " + name + ", from Bujo!"
}

func (a *App) GetAgenda(from, to time.Time) (*service.MultiDayAgenda, error) {
	return a.services.Bujo.GetMultiDayAgenda(a.ctx, from, to)
}

func (a *App) GetHabits(days int) (*service.TrackerStatus, error) {
	return a.services.Habit.GetTrackerStatus(a.ctx, time.Now(), days)
}
