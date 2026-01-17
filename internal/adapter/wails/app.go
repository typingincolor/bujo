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

func (a *App) MarkEntryDone(id int64) error {
	return a.services.Bujo.MarkDone(a.ctx, id)
}

func (a *App) MarkEntryUndone(id int64) error {
	return a.services.Bujo.Undo(a.ctx, id)
}

func (a *App) AddEntry(input string, date time.Time) ([]int64, error) {
	return a.services.Bujo.LogEntries(a.ctx, input, service.LogEntriesOptions{Date: date})
}

func (a *App) LogHabit(habitID int64, count int) error {
	return a.services.Habit.LogHabitByID(a.ctx, habitID, count)
}

func (a *App) AddListItem(listID int64, content string) (int64, error) {
	return a.services.List.AddItem(a.ctx, listID, domain.EntryTypeTask, content)
}

func (a *App) MarkListItemDone(itemID int64) error {
	return a.services.List.MarkDone(a.ctx, itemID)
}

func (a *App) MarkListItemUndone(itemID int64) error {
	return a.services.List.MarkUndone(a.ctx, itemID)
}

func (a *App) RemoveListItem(itemID int64) error {
	return a.services.List.RemoveItem(a.ctx, itemID)
}

func (a *App) CreateGoal(content string, month time.Time) (int64, error) {
	return a.services.Goal.CreateGoal(a.ctx, content, month)
}

func (a *App) MarkGoalDone(goalID int64) error {
	return a.services.Goal.MarkDone(a.ctx, goalID)
}

func (a *App) MarkGoalActive(goalID int64) error {
	return a.services.Goal.MarkActive(a.ctx, goalID)
}

func (a *App) DeleteGoal(goalID int64) error {
	return a.services.Goal.DeleteGoal(a.ctx, goalID)
}

func (a *App) Search(query string) ([]domain.Entry, error) {
	opts := domain.NewSearchOptions(query)
	return a.services.Bujo.SearchEntries(a.ctx, opts)
}
