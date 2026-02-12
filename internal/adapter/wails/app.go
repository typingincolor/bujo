package wails

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/typingincolor/bujo/cmd/bujo/cmd"
	bujohttp "github.com/typingincolor/bujo/internal/adapter/http"
	"github.com/typingincolor/bujo/internal/app"
	"github.com/typingincolor/bujo/internal/dateutil"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ListWithItems struct {
	ID    int64
	Name  string
	Items []domain.ListItem
}

const (
	changePollingInterval = 2 * time.Second
	eventDataChanged      = "data:changed"
)

type App struct {
	ctx          context.Context
	services     *app.Services
	lastModified time.Time
	stopPolling  chan struct{}
	httpServer   *bujohttp.Server
}

func NewApp(services *app.Services) *App {
	return &App{
		services: services,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.stopPolling = make(chan struct{})

	if a.services.ChangeDetection != nil {
		lastMod, _ := a.services.ChangeDetection.GetLastModified(ctx)
		a.lastModified = lastMod
		go a.pollForChanges()
	}

	a.httpServer = bujohttp.NewServer(a.services.Bujo, 0)
	if _, err := a.httpServer.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to start HTTP API: %v\n", err)
	}
}

func (a *App) pollForChanges() {
	ticker := time.NewTicker(changePollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopPolling:
			return
		case <-ticker.C:
			a.checkForChanges()
		}
	}
}

func (a *App) checkForChanges() {
	if a.services.ChangeDetection == nil {
		return
	}

	currentMod, err := a.services.ChangeDetection.GetLastModified(a.ctx)
	if err != nil {
		return
	}

	if currentMod.After(a.lastModified) {
		a.lastModified = currentMod
		runtime.EventsEmit(a.ctx, eventDataChanged)
	}
}

func (a *App) Shutdown(_ context.Context) {
	if a.stopPolling != nil {
		close(a.stopPolling)
	}
	if a.httpServer != nil {
		_ = a.httpServer.Stop()
	}
}

func (a *App) Greet(name string) string {
	return "Hello " + name + ", from Bujo!"
}

func (a *App) GetDayEntries(from, to time.Time) ([]service.DayEntries, error) {
	return a.services.Bujo.GetDayEntries(a.ctx, from, to)
}

func (a *App) GetOverdue() ([]domain.Entry, error) {
	return a.services.Bujo.GetOverdue(a.ctx)
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

func (a *App) AddChildEntry(parentID int64, input string, date time.Time) ([]int64, error) {
	return a.services.Bujo.LogEntries(a.ctx, input, service.LogEntriesOptions{Date: date, ParentID: &parentID})
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

func (a *App) SearchByTags(tags []string) ([]domain.Entry, error) {
	opts := domain.NewSearchOptions("").WithTags(tags)
	return a.services.Bujo.SearchEntries(a.ctx, opts)
}

func (a *App) GetAllTags() ([]string, error) {
	return a.services.Bujo.GetAllTags(a.ctx)
}

func (a *App) SearchByMentions(mentions []string) ([]domain.Entry, error) {
	opts := domain.NewSearchOptions("").WithMentions(mentions)
	return a.services.Bujo.SearchEntries(a.ctx, opts)
}

func (a *App) GetAllMentions() ([]string, error) {
	return a.services.Bujo.GetAllMentions(a.ctx)
}

func (a *App) GetEntry(id int64) (*domain.Entry, error) {
	return a.services.Bujo.GetEntry(a.ctx, id)
}

func (a *App) GetEntryAncestors(id int64) ([]domain.Entry, error) {
	return a.services.Bujo.GetEntryAncestors(a.ctx, id)
}

func (a *App) GetEntryContext(id int64) ([]domain.Entry, error) {
	return a.services.Bujo.GetEntryContext(a.ctx, id, 100)
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

func (a *App) SetHabitWeeklyGoal(habitID int64, weeklyGoal int) error {
	return a.services.Habit.SetHabitWeeklyGoalByID(a.ctx, habitID, weeklyGoal)
}

func (a *App) SetHabitMonthlyGoal(habitID int64, monthlyGoal int) error {
	return a.services.Habit.SetHabitMonthlyGoalByID(a.ctx, habitID, monthlyGoal)
}

func (a *App) AnswerQuestion(questionID int64, answerText string) error {
	return a.services.Bujo.MarkAnswered(a.ctx, questionID, answerText)
}

func (a *App) GetOutstandingQuestions() ([]domain.Entry, error) {
	opts := domain.NewSearchOptions("").WithType(domain.EntryTypeQuestion).WithLimit(100)
	return a.services.Bujo.SearchEntries(a.ctx, opts)
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

func (a *App) RetypeEntry(id int64, newType string) error {
	return a.services.Bujo.RetypeEntry(a.ctx, id, domain.EntryType(newType))
}

func (a *App) MoveEntryToRoot(id int64) error {
	moveToRoot := true
	return a.services.Bujo.MoveEntry(a.ctx, id, service.MoveOptions{MoveToRoot: &moveToRoot})
}

func (a *App) MoveEntryToList(entryID int64, listID int64) error {
	return a.services.Bujo.MoveEntryToList(a.ctx, entryID, listID)
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

func (a *App) GetVersion() string {
	return cmd.Version()
}

const maxFileSize = 1024 * 1024

func (a *App) ReadFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.Size() > maxFileSize {
		return "", fmt.Errorf("file too large: %d bytes (max %d bytes)", info.Size(), maxFileSize)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (a *App) OpenFileDialog() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Import Entries",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Text Files (*.txt, *.md)",
				Pattern:     "*.txt;*.md",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	return a.ReadFile(path)
}

type ValidationError struct {
	LineNumber int    `json:"lineNumber"`
	Message    string `json:"message"`
}

type ValidationResult struct {
	IsValid bool              `json:"isValid"`
	Errors  []ValidationError `json:"errors"`
}

type ApplyResult struct {
	Inserted int `json:"inserted"`
	Deleted  int `json:"deleted"`
}

type ResolvedDate struct {
	ISO     string `json:"iso"`
	Display string `json:"display"`
}

func (a *App) GetEditableDocument(date time.Time) (string, error) {
	return a.services.EditableView.GetEditableDocument(a.ctx, date)
}

func (a *App) ValidateEditableDocument(doc string) ValidationResult {
	result := a.services.EditableView.ValidateDocument(doc)
	errors := make([]ValidationError, len(result.Errors))
	for i, err := range result.Errors {
		errors[i] = ValidationError{
			LineNumber: err.LineNumber,
			Message:    err.Message,
		}
	}
	return ValidationResult{
		IsValid: result.IsValid,
		Errors:  errors,
	}
}

func (a *App) ApplyEditableDocument(doc string, date time.Time) (*ApplyResult, error) {
	result, err := a.services.EditableView.ApplyChanges(a.ctx, doc, date)
	if err != nil {
		return nil, err
	}

	return &ApplyResult{
		Inserted: result.Inserted,
		Deleted:  result.Deleted,
	}, nil
}

func (a *App) ApplyEditableDocumentWithActions(doc string, date time.Time, migrateDate *time.Time, listID *int64) (*ApplyResult, error) {
	actions := service.ApplyActions{
		MigrateDate: migrateDate,
		ListID:      listID,
	}
	result, err := a.services.EditableView.ApplyChangesWithActions(a.ctx, doc, date, actions)
	if err != nil {
		return nil, err
	}

	return &ApplyResult{
		Inserted: result.Inserted,
		Deleted:  result.Deleted,
	}, nil
}

func (a *App) IsInsightsAvailable() bool {
	return a.services.InsightsRepo.IsAvailable()
}

func (a *App) GetInsightsDashboard() (*domain.InsightsDashboard, error) {
	repo := a.services.InsightsRepo
	if !repo.IsAvailable() {
		return &domain.InsightsDashboard{Status: "not_initialized"}, nil
	}

	latest, err := repo.GetLatestSummary(a.ctx)
	if err != nil {
		return nil, err
	}

	initiatives, err := repo.GetActiveInitiatives(a.ctx, 5)
	if err != nil {
		return nil, err
	}

	actions, err := repo.GetPendingActions(a.ctx)
	if err != nil {
		return nil, err
	}

	var highPriority []domain.InsightsAction
	for _, action := range actions {
		if action.Priority == "high" {
			highPriority = append(highPriority, action)
		}
	}

	decisions, err := repo.GetRecentDecisions(a.ctx, 3)
	if err != nil {
		return nil, err
	}

	days, err := repo.GetDaysSinceLastSummary(a.ctx)
	if err != nil {
		return nil, err
	}

	status := "ready"
	if latest == nil {
		status = "empty"
	}

	return &domain.InsightsDashboard{
		LatestSummary:        latest,
		ActiveInitiatives:    initiatives,
		HighPriorityActions:  highPriority,
		RecentDecisions:      decisions,
		DaysSinceLastSummary: days,
		Status:               status,
	}, nil
}

func (a *App) GetInsightsSummaries(limit int) ([]domain.InsightsSummary, error) {
	repo := a.services.InsightsRepo
	if !repo.IsAvailable() {
		return []domain.InsightsSummary{}, nil
	}
	return repo.GetSummaries(a.ctx, limit)
}

func (a *App) GetInsightsSummaryDetail(summaryID int64) ([]domain.InsightsTopic, error) {
	repo := a.services.InsightsRepo
	if !repo.IsAvailable() {
		return []domain.InsightsTopic{}, nil
	}
	return repo.GetTopicsForSummary(a.ctx, summaryID)
}

func (a *App) GetInsightsActions() ([]domain.InsightsAction, error) {
	repo := a.services.InsightsRepo
	if !repo.IsAvailable() {
		return []domain.InsightsAction{}, nil
	}
	return repo.GetPendingActions(a.ctx)
}

type WeekSummaryDetail struct {
	Summary *domain.InsightsSummary
	Topics  []domain.InsightsTopic
}

func nextWeekStartFrom(weekStart string) (string, error) {
	t, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return "", fmt.Errorf("invalid weekStart date: %w", err)
	}
	return t.AddDate(0, 0, 7).Format("2006-01-02"), nil
}

func (a *App) GetInsightsSummaryForWeek(weekStart string) (*WeekSummaryDetail, error) {
	repo := a.services.InsightsRepo
	if !repo.IsAvailable() {
		return &WeekSummaryDetail{Topics: []domain.InsightsTopic{}}, nil
	}
	nextWeek, err := nextWeekStartFrom(weekStart)
	if err != nil {
		return nil, err
	}
	summary, err := repo.GetSummaryForWeek(a.ctx, weekStart, nextWeek)
	if err != nil {
		return nil, err
	}
	if summary == nil {
		return &WeekSummaryDetail{Topics: []domain.InsightsTopic{}}, nil
	}
	topics, err := repo.GetTopicsForSummary(a.ctx, summary.ID)
	if err != nil {
		return nil, err
	}
	return &WeekSummaryDetail{Summary: summary, Topics: topics}, nil
}

func (a *App) GetInsightsActionsForWeek(weekStart string) ([]domain.InsightsAction, error) {
	repo := a.services.InsightsRepo
	if !repo.IsAvailable() {
		return []domain.InsightsAction{}, nil
	}
	nextWeek, err := nextWeekStartFrom(weekStart)
	if err != nil {
		return nil, err
	}
	return repo.GetActionsForWeek(a.ctx, weekStart, nextWeek)
}

func (a *App) GetInsightsInitiativePortfolio() ([]domain.InsightsInitiativePortfolio, error) {
	repo := a.services.InsightsRepo
	if !repo.IsAvailable() {
		return []domain.InsightsInitiativePortfolio{}, nil
	}
	return repo.GetInitiativePortfolio(a.ctx)
}

func (a *App) GetInsightsInitiativeDetail(initiativeID int64) (*domain.InsightsInitiativeDetail, error) {
	repo := a.services.InsightsRepo
	if !repo.IsAvailable() {
		return &domain.InsightsInitiativeDetail{}, nil
	}
	return repo.GetInitiativeDetail(a.ctx, initiativeID)
}

func (a *App) GetInsightsDistinctTopics() ([]string, error) {
	repo := a.services.InsightsRepo
	if !repo.IsAvailable() {
		return []string{}, nil
	}
	return repo.GetDistinctTopics(a.ctx)
}

func (a *App) GetInsightsTopicTimeline(topic string) ([]domain.InsightsTopicTimeline, error) {
	repo := a.services.InsightsRepo
	if !repo.IsAvailable() {
		return []domain.InsightsTopicTimeline{}, nil
	}
	return repo.GetTopicTimeline(a.ctx, topic)
}

func (a *App) GetInsightsDecisionLog() ([]domain.InsightsDecisionWithInitiatives, error) {
	repo := a.services.InsightsRepo
	if !repo.IsAvailable() {
		return []domain.InsightsDecisionWithInitiatives{}, nil
	}
	return repo.GetDecisionsWithInitiatives(a.ctx)
}

func (a *App) GetInsightsWeeklyReport(weekStart string) (*domain.InsightsWeeklyReport, error) {
	repo := a.services.InsightsRepo
	if !repo.IsAvailable() {
		return &domain.InsightsWeeklyReport{}, nil
	}
	t, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return nil, err
	}
	nextWeekStart := t.AddDate(0, 0, 7).Format("2006-01-02")
	return repo.GetWeeklyReport(a.ctx, weekStart, nextWeekStart)
}

func (a *App) GetAttentionScores(ids []int64) (map[int64]domain.AttentionResult, error) {
	result, err := a.services.Bujo.GetAttentionScores(a.ctx, ids)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAttentionResultType forces Wails to generate TypeScript bindings for domain.AttentionResult.
// Without this method, Wails doesn't generate types for structs only used as map values.
func (a *App) GetAttentionResultType() domain.AttentionResult {
	return domain.AttentionResult{}
}

func (a *App) ResolveDate(input string) (*ResolvedDate, error) {
	parsed, err := dateutil.ParseFuture(input)
	if err != nil {
		return nil, err
	}

	return &ResolvedDate{
		ISO:     parsed.Format("2006-01-02"),
		Display: parsed.Format("Mon, Jan 2, 2006"),
	}, nil
}
