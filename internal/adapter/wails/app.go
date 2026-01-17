package wails

import (
	"context"
	"time"

	"github.com/typingincolor/bujo/internal/app"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

type ListWithItems struct {
	ID    int64
	Name  string
	Items []domain.ListItem
}

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

func (a *App) GetLists() ([]ListWithItems, error) {
	lists, err := a.services.List.GetAllLists(a.ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ListWithItems, 0, len(lists))
	for _, list := range lists {
		items, err := a.services.List.GetListItems(a.ctx, list.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, ListWithItems{
			ID:    list.ID,
			Name:  list.Name,
			Items: items,
		})
	}
	return result, nil
}

func (a *App) GetGoals(month time.Time) ([]domain.Goal, error) {
	return a.services.Goal.GetGoalsForMonth(a.ctx, month)
}
