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

func (a *App) MigrateGoal(goalID int64, toMonth time.Time) (int64, error) {
	return a.services.Goal.MigrateGoal(a.ctx, goalID, toMonth)
}

func (a *App) UpdateGoal(goalID int64, content string) error {
	return a.services.Goal.UpdateGoal(a.ctx, goalID, content)
}

func (a *App) CancelGoal(goalID int64) error {
	return a.services.Goal.CancelGoal(a.ctx, goalID)
}

func (a *App) UncancelGoal(goalID int64) error {
	return a.services.Goal.UncancelGoal(a.ctx, goalID)
}

func (a *App) Search(query string) ([]domain.Entry, error) {
	opts := domain.NewSearchOptions(query)
	return a.services.Bujo.SearchEntries(a.ctx, opts)
}

func (a *App) GetEntryAncestors(id int64) ([]domain.Entry, error) {
	return a.services.Bujo.GetEntryAncestors(a.ctx, id)
}

func (a *App) EditEntry(id int64, newContent string) error {
	return a.services.Bujo.EditEntry(a.ctx, id, newContent)
}

func (a *App) DeleteEntry(id int64) error {
	return a.services.Bujo.DeleteEntry(a.ctx, id)
}

func (a *App) HasChildren(id int64) (bool, error) {
	return a.services.Bujo.HasChildren(a.ctx, id)
}

func (a *App) CreateHabit(name string) (int64, error) {
	return a.services.Habit.CreateHabit(a.ctx, name)
}

func (a *App) DeleteHabit(habitID int64) error {
	return a.services.Habit.DeleteHabitByID(a.ctx, habitID)
}

func (a *App) LogHabitForDate(habitID int64, count int, date time.Time) error {
	return a.services.Habit.LogHabitByIDForDate(a.ctx, habitID, count, date)
}

func (a *App) UndoHabitLog(habitID int64) error {
	return a.services.Habit.UndoLastLogByID(a.ctx, habitID)
}

func (a *App) UndoHabitLogForDate(habitID int64, date time.Time) error {
	return a.services.Habit.RemoveHabitLogForDateByID(a.ctx, habitID, date)
}

func (a *App) SetHabitGoal(habitID int64, dailyGoal int) error {
	return a.services.Habit.SetHabitGoalByID(a.ctx, habitID, dailyGoal)
}

func (a *App) AnswerQuestion(questionID int64, answerText string) error {
	return a.services.Bujo.MarkAnswered(a.ctx, questionID, answerText)
}

func (a *App) CreateList(name string) (int64, error) {
	list, err := a.services.List.CreateList(a.ctx, name)
	if err != nil {
		return 0, err
	}
	return list.ID, nil
}

func (a *App) DeleteList(listID int64, force bool) error {
	return a.services.List.DeleteList(a.ctx, listID, force)
}

func (a *App) RenameList(listID int64, newName string) error {
	return a.services.List.RenameList(a.ctx, listID, newName)
}

func (a *App) EditListItem(itemID int64, content string) error {
	return a.services.List.EditItem(a.ctx, itemID, content)
}

func (a *App) CancelListItem(itemID int64) error {
	return a.services.List.Cancel(a.ctx, itemID)
}

func (a *App) UncancelListItem(itemID int64) error {
	return a.services.List.Uncancel(a.ctx, itemID)
}

func (a *App) MoveListItem(itemID int64, targetListID int64) error {
	return a.services.List.MoveItem(a.ctx, itemID, targetListID)
}

func (a *App) SetMood(date time.Time, mood string) error {
	return a.services.Bujo.SetMood(a.ctx, date, mood)
}

func (a *App) SetWeather(date time.Time, weather string) error {
	return a.services.Bujo.SetWeather(a.ctx, date, weather)
}

func (a *App) CancelEntry(id int64) error {
	return a.services.Bujo.CancelEntry(a.ctx, id)
}

func (a *App) UncancelEntry(id int64) error {
	return a.services.Bujo.UncancelEntry(a.ctx, id)
}

func (a *App) SetPriority(id int64, priority string) error {
	p, err := domain.ParsePriority(priority)
	if err != nil {
		return err
	}
	return a.services.Bujo.EditEntryPriority(a.ctx, id, p)
}

func (a *App) CyclePriority(id int64) error {
	return a.services.Bujo.CyclePriority(a.ctx, id)
}

func (a *App) MigrateEntry(id int64, toDate time.Time) (int64, error) {
	return a.services.Bujo.MigrateEntry(a.ctx, id, toDate)
}

func (a *App) SetLocation(date time.Time, location string) error {
	return a.services.Bujo.SetLocation(a.ctx, date, location)
}

const locationHistoryMonths = 6

func (a *App) GetLocationHistory() ([]string, error) {
	now := time.Now()
	from := now.AddDate(0, -locationHistoryMonths, 0)
	history, err := a.services.Bujo.GetLocationHistory(a.ctx, from, now)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var locations []string
	for _, dayCtx := range history {
		if dayCtx.Location != nil && *dayCtx.Location != "" && !seen[*dayCtx.Location] {
			seen[*dayCtx.Location] = true
			locations = append(locations, *dayCtx.Location)
		}
	}
	return locations, nil
}

func (a *App) GetSummary(date time.Time) (string, error) {
	if a.services.Summary == nil {
		return "", nil
	}
	summary, err := a.services.Summary.GetSummary(a.ctx, domain.SummaryHorizonDaily, date)
	if err != nil {
		return "", err
	}
	if summary == nil {
		return "", nil
	}
	return summary.Content, nil
}
