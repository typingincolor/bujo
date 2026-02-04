package service

import (
	"context"
	"testing"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type mockEntryRepoForExport struct {
	entries []domain.Entry
}

func (m *mockEntryRepoForExport) GetAll(ctx context.Context) ([]domain.Entry, error) {
	return m.entries, nil
}

func (m *mockEntryRepoForExport) GetByDateRange(ctx context.Context, from, to time.Time) ([]domain.Entry, error) {
	var result []domain.Entry
	for _, e := range m.entries {
		if e.ScheduledDate != nil && !e.ScheduledDate.Before(from) && !e.ScheduledDate.After(to) {
			result = append(result, e)
		}
	}
	return result, nil
}

type mockHabitRepoForExport struct {
	habits []domain.Habit
}

func (m *mockHabitRepoForExport) GetAll(ctx context.Context) ([]domain.Habit, error) {
	return m.habits, nil
}

type mockHabitLogRepoForExport struct {
	logs []domain.HabitLog
}

func (m *mockHabitLogRepoForExport) GetAll(ctx context.Context) ([]domain.HabitLog, error) {
	return m.logs, nil
}

type mockDayContextRepoForExport struct {
	contexts []domain.DayContext
}

func (m *mockDayContextRepoForExport) GetAll(ctx context.Context) ([]domain.DayContext, error) {
	return m.contexts, nil
}

type mockListRepoForExport struct {
	lists []domain.List
}

func (m *mockListRepoForExport) GetAll(ctx context.Context) ([]domain.List, error) {
	return m.lists, nil
}

type mockListItemRepoForExport struct {
	items []domain.ListItem
}

func (m *mockListItemRepoForExport) GetAll(ctx context.Context) ([]domain.ListItem, error) {
	return m.items, nil
}

type mockGoalRepoForExport struct {
	goals []domain.Goal
}

func (m *mockGoalRepoForExport) GetAll(ctx context.Context) ([]domain.Goal, error) {
	return m.goals, nil
}

func TestExportService_Export_AllData(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	date := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)

	entryRepo := &mockEntryRepoForExport{
		entries: []domain.Entry{
			{ID: 1, Content: "Task 1", Type: domain.EntryTypeTask, ScheduledDate: &date},
			{ID: 2, Content: "Note 1", Type: domain.EntryTypeNote, ScheduledDate: &date},
		},
	}
	habitRepo := &mockHabitRepoForExport{
		habits: []domain.Habit{
			{ID: 1, Name: "Exercise", GoalPerDay: 1},
		},
	}
	habitLogRepo := &mockHabitLogRepoForExport{
		logs: []domain.HabitLog{
			{ID: 1, HabitID: 1, Count: 1, LoggedAt: now},
		},
	}
	dayContextRepo := &mockDayContextRepoForExport{
		contexts: []domain.DayContext{
			{Date: date, EntityID: domain.NewEntityID()},
		},
	}
	listRepo := &mockListRepoForExport{
		lists: []domain.List{
			{ID: 1, Name: "Groceries"},
		},
	}
	listItemRepo := &mockListItemRepoForExport{
		items: []domain.ListItem{
			{VersionInfo: domain.VersionInfo{RowID: 1}, Content: "Milk", Type: domain.ListItemTypeTask},
		},
	}
	goalRepo := &mockGoalRepoForExport{
		goals: []domain.Goal{
			{ID: 1, Content: "Learn Go", Status: domain.GoalStatusActive},
		},
	}

	svc := NewExportService(entryRepo, habitRepo, habitLogRepo, dayContextRepo, listRepo, listItemRepo, goalRepo)

	data, err := svc.Export(ctx, domain.NewExportOptions())
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if data.Version != domain.ExportVersion {
		t.Errorf("Expected version %s, got %s", domain.ExportVersion, data.Version)
	}

	if len(data.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(data.Entries))
	}

	if len(data.Habits) != 1 {
		t.Errorf("Expected 1 habit, got %d", len(data.Habits))
	}

	if len(data.HabitLogs) != 1 {
		t.Errorf("Expected 1 habit log, got %d", len(data.HabitLogs))
	}

	if len(data.DayContexts) != 1 {
		t.Errorf("Expected 1 day context, got %d", len(data.DayContexts))
	}

	if len(data.Lists) != 1 {
		t.Errorf("Expected 1 list, got %d", len(data.Lists))
	}

	if len(data.ListItems) != 1 {
		t.Errorf("Expected 1 list item, got %d", len(data.ListItems))
	}

	if len(data.Goals) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(data.Goals))
	}
}

func TestExportService_Export_WithDateRange(t *testing.T) {
	ctx := context.Background()
	jan5 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	jan10 := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	jan15 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	entryRepo := &mockEntryRepoForExport{
		entries: []domain.Entry{
			{ID: 1, Content: "Entry on Jan 5", Type: domain.EntryTypeTask, ScheduledDate: &jan5},
			{ID: 2, Content: "Entry on Jan 10", Type: domain.EntryTypeTask, ScheduledDate: &jan10},
			{ID: 3, Content: "Entry on Jan 15", Type: domain.EntryTypeTask, ScheduledDate: &jan15},
		},
	}

	svc := NewExportService(entryRepo, &mockHabitRepoForExport{}, &mockHabitLogRepoForExport{},
		&mockDayContextRepoForExport{}, &mockListRepoForExport{},
		&mockListItemRepoForExport{}, &mockGoalRepoForExport{})

	from := time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	opts := domain.NewExportOptions().WithDateRange(from, to)

	data, err := svc.Export(ctx, opts)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if len(data.Entries) != 1 {
		t.Errorf("Expected 1 entry in date range, got %d", len(data.Entries))
	}

	if len(data.Entries) > 0 && data.Entries[0].Content != "Entry on Jan 10" {
		t.Errorf("Expected 'Entry on Jan 10', got '%s'", data.Entries[0].Content)
	}
}

func TestExportService_Export_EmptyData(t *testing.T) {
	ctx := context.Background()

	svc := NewExportService(&mockEntryRepoForExport{}, &mockHabitRepoForExport{}, &mockHabitLogRepoForExport{},
		&mockDayContextRepoForExport{}, &mockListRepoForExport{},
		&mockListItemRepoForExport{}, &mockGoalRepoForExport{})

	data, err := svc.Export(ctx, domain.NewExportOptions())
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if data.Version != domain.ExportVersion {
		t.Errorf("Expected version %s, got %s", domain.ExportVersion, data.Version)
	}

	if data.Entries == nil {
		t.Error("Entries should be empty slice, not nil")
	}
}

type mockImportEntryRepo struct {
	existing map[domain.EntityID]bool
	inserted []domain.Entry
	cleared  bool
}

func (m *mockImportEntryRepo) Insert(ctx context.Context, entry domain.Entry) (int64, error) {
	m.inserted = append(m.inserted, entry)
	return int64(len(m.inserted)), nil
}

func (m *mockImportEntryRepo) DeleteAll(ctx context.Context) error {
	m.cleared = true
	m.existing = make(map[domain.EntityID]bool)
	return nil
}

type mockImportHabitRepo struct {
	existing map[domain.EntityID]bool
	inserted []domain.Habit
	cleared  bool
}

func (m *mockImportHabitRepo) Insert(ctx context.Context, habit domain.Habit) (int64, error) {
	m.inserted = append(m.inserted, habit)
	return int64(len(m.inserted)), nil
}

func (m *mockImportHabitRepo) GetByEntityID(ctx context.Context, entityID domain.EntityID) (*domain.Habit, error) {
	if m.existing[entityID] {
		return &domain.Habit{EntityID: entityID}, nil
	}
	return nil, nil
}

func (m *mockImportHabitRepo) DeleteAll(ctx context.Context) error {
	m.cleared = true
	m.existing = make(map[domain.EntityID]bool)
	return nil
}

type mockImportHabitLogRepo struct {
	inserted []domain.HabitLog
	cleared  bool
}

func (m *mockImportHabitLogRepo) Insert(ctx context.Context, log domain.HabitLog) (int64, error) {
	m.inserted = append(m.inserted, log)
	return int64(len(m.inserted)), nil
}

func (m *mockImportHabitLogRepo) DeleteAll(ctx context.Context) error {
	m.cleared = true
	return nil
}

type mockImportDayContextRepo struct {
	inserted []domain.DayContext
	cleared  bool
}

func (m *mockImportDayContextRepo) Upsert(ctx context.Context, dc domain.DayContext) error {
	m.inserted = append(m.inserted, dc)
	return nil
}

func (m *mockImportDayContextRepo) DeleteAll(ctx context.Context) error {
	m.cleared = true
	return nil
}

type mockImportListRepo struct {
	existing map[domain.EntityID]bool
	inserted []domain.List
	cleared  bool
}

func (m *mockImportListRepo) Create(ctx context.Context, name string) (*domain.List, error) {
	return nil, nil
}

func (m *mockImportListRepo) InsertWithEntityID(ctx context.Context, list domain.List) (int64, error) {
	m.inserted = append(m.inserted, list)
	return int64(len(m.inserted)), nil
}

func (m *mockImportListRepo) GetByEntityID(ctx context.Context, entityID domain.EntityID) (*domain.List, error) {
	if m.existing[entityID] {
		return &domain.List{EntityID: entityID}, nil
	}
	return nil, nil
}

func (m *mockImportListRepo) DeleteAll(ctx context.Context) error {
	m.cleared = true
	m.existing = make(map[domain.EntityID]bool)
	return nil
}

type mockImportListItemRepo struct {
	inserted []domain.ListItem
	cleared  bool
}

func (m *mockImportListItemRepo) Insert(ctx context.Context, item domain.ListItem) (int64, error) {
	m.inserted = append(m.inserted, item)
	return int64(len(m.inserted)), nil
}

func (m *mockImportListItemRepo) DeleteAll(ctx context.Context) error {
	m.cleared = true
	return nil
}

type mockImportGoalRepo struct {
	existing map[domain.EntityID]bool
	inserted []domain.Goal
	cleared  bool
}

func (m *mockImportGoalRepo) Insert(ctx context.Context, goal domain.Goal) (int64, error) {
	m.inserted = append(m.inserted, goal)
	return int64(len(m.inserted)), nil
}

func (m *mockImportGoalRepo) GetByEntityID(ctx context.Context, entityID domain.EntityID) (*domain.Goal, error) {
	if m.existing[entityID] {
		return &domain.Goal{EntityID: entityID}, nil
	}
	return nil, nil
}

func (m *mockImportGoalRepo) DeleteAll(ctx context.Context) error {
	m.cleared = true
	m.existing = make(map[domain.EntityID]bool)
	return nil
}

func TestImportService_Import_MergeMode(t *testing.T) {
	ctx := context.Background()

	existingEntityID := domain.NewEntityID()
	newEntityID := domain.NewEntityID()

	entryRepo := &mockImportEntryRepo{
		existing: map[domain.EntityID]bool{existingEntityID: true},
	}
	habitRepo := &mockImportHabitRepo{existing: make(map[domain.EntityID]bool)}
	habitLogRepo := &mockImportHabitLogRepo{}
	dayContextRepo := &mockImportDayContextRepo{}
	listRepo := &mockImportListRepo{existing: make(map[domain.EntityID]bool)}
	listItemRepo := &mockImportListItemRepo{}
	goalRepo := &mockImportGoalRepo{existing: make(map[domain.EntityID]bool)}

	svc := NewImportService(entryRepo, habitRepo, habitLogRepo, dayContextRepo, listRepo, listItemRepo, goalRepo)

	data := &domain.ExportData{
		Version: domain.ExportVersion,
		Entries: []domain.Entry{
			{EntityID: existingEntityID, Content: "Existing entry"},
			{EntityID: newEntityID, Content: "New entry"},
		},
	}

	opts := domain.NewImportOptions(domain.ImportModeMerge)
	err := svc.Import(ctx, data, opts)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if len(entryRepo.inserted) != 2 {
		t.Errorf("Expected 2 inserted entries (entries always insert without dedup), got %d", len(entryRepo.inserted))
	}
}

func TestImportService_Import_ReplaceMode(t *testing.T) {
	ctx := context.Background()

	entryRepo := &mockImportEntryRepo{
		existing: map[domain.EntityID]bool{domain.NewEntityID(): true},
	}
	habitRepo := &mockImportHabitRepo{existing: make(map[domain.EntityID]bool)}
	habitLogRepo := &mockImportHabitLogRepo{}
	dayContextRepo := &mockImportDayContextRepo{}
	listRepo := &mockImportListRepo{existing: make(map[domain.EntityID]bool)}
	listItemRepo := &mockImportListItemRepo{}
	goalRepo := &mockImportGoalRepo{existing: make(map[domain.EntityID]bool)}

	svc := NewImportService(entryRepo, habitRepo, habitLogRepo, dayContextRepo, listRepo, listItemRepo, goalRepo)

	data := &domain.ExportData{
		Version: domain.ExportVersion,
		Entries: []domain.Entry{
			{EntityID: domain.NewEntityID(), Content: "Imported entry"},
		},
		Habits: []domain.Habit{
			{EntityID: domain.NewEntityID(), Name: "Imported habit"},
		},
	}

	opts := domain.NewImportOptions(domain.ImportModeReplace)
	err := svc.Import(ctx, data, opts)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if !entryRepo.cleared {
		t.Error("Expected entries to be cleared in replace mode")
	}

	if !habitRepo.cleared {
		t.Error("Expected habits to be cleared in replace mode")
	}

	if len(entryRepo.inserted) != 1 {
		t.Errorf("Expected 1 inserted entry, got %d", len(entryRepo.inserted))
	}

	if len(habitRepo.inserted) != 1 {
		t.Errorf("Expected 1 inserted habit, got %d", len(habitRepo.inserted))
	}
}
