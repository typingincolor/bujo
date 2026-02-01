package service

import (
	"context"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type ExportEntryRepository interface {
	GetAll(ctx context.Context) ([]domain.Entry, error)
	GetByDateRange(ctx context.Context, from, to time.Time) ([]domain.Entry, error)
}

type ExportHabitRepository interface {
	GetAll(ctx context.Context) ([]domain.Habit, error)
}

type ExportHabitLogRepository interface {
	GetAll(ctx context.Context) ([]domain.HabitLog, error)
}

type ExportDayContextRepository interface {
	GetAll(ctx context.Context) ([]domain.DayContext, error)
}

type ExportSummaryRepository interface {
	GetAll(ctx context.Context) ([]domain.Summary, error)
}

type ExportListRepository interface {
	GetAll(ctx context.Context) ([]domain.List, error)
}

type ExportListItemRepository interface {
	GetAll(ctx context.Context) ([]domain.ListItem, error)
}

type ExportGoalRepository interface {
	GetAll(ctx context.Context) ([]domain.Goal, error)
}

type ExportService struct {
	entryRepo      ExportEntryRepository
	habitRepo      ExportHabitRepository
	habitLogRepo   ExportHabitLogRepository
	dayContextRepo ExportDayContextRepository
	summaryRepo    ExportSummaryRepository
	listRepo       ExportListRepository
	listItemRepo   ExportListItemRepository
	goalRepo       ExportGoalRepository
}

func NewExportService(
	entryRepo ExportEntryRepository,
	habitRepo ExportHabitRepository,
	habitLogRepo ExportHabitLogRepository,
	dayContextRepo ExportDayContextRepository,
	summaryRepo ExportSummaryRepository,
	listRepo ExportListRepository,
	listItemRepo ExportListItemRepository,
	goalRepo ExportGoalRepository,
) *ExportService {
	return &ExportService{
		entryRepo:      entryRepo,
		habitRepo:      habitRepo,
		habitLogRepo:   habitLogRepo,
		dayContextRepo: dayContextRepo,
		summaryRepo:    summaryRepo,
		listRepo:       listRepo,
		listItemRepo:   listItemRepo,
		goalRepo:       goalRepo,
	}
}

type ImportEntryRepository interface {
	Insert(ctx context.Context, entry domain.Entry) (int64, error)
	DeleteAll(ctx context.Context) error
}

type ImportHabitRepository interface {
	Insert(ctx context.Context, habit domain.Habit) (int64, error)
	GetByEntityID(ctx context.Context, entityID domain.EntityID) (*domain.Habit, error)
	DeleteAll(ctx context.Context) error
}

type ImportHabitLogRepository interface {
	Insert(ctx context.Context, log domain.HabitLog) (int64, error)
	DeleteAll(ctx context.Context) error
}

type ImportDayContextRepository interface {
	Upsert(ctx context.Context, dc domain.DayContext) error
	DeleteAll(ctx context.Context) error
}

type ImportSummaryRepository interface {
	Insert(ctx context.Context, s domain.Summary) (int64, error)
	DeleteAll(ctx context.Context) error
}

type ImportListRepository interface {
	InsertWithEntityID(ctx context.Context, list domain.List) (int64, error)
	GetByEntityID(ctx context.Context, entityID domain.EntityID) (*domain.List, error)
	DeleteAll(ctx context.Context) error
}

type ImportListItemRepository interface {
	Insert(ctx context.Context, item domain.ListItem) (int64, error)
	DeleteAll(ctx context.Context) error
}

type ImportGoalRepository interface {
	Insert(ctx context.Context, goal domain.Goal) (int64, error)
	GetByEntityID(ctx context.Context, entityID domain.EntityID) (*domain.Goal, error)
	DeleteAll(ctx context.Context) error
}

type ImportService struct {
	entryRepo      ImportEntryRepository
	habitRepo      ImportHabitRepository
	habitLogRepo   ImportHabitLogRepository
	dayContextRepo ImportDayContextRepository
	summaryRepo    ImportSummaryRepository
	listRepo       ImportListRepository
	listItemRepo   ImportListItemRepository
	goalRepo       ImportGoalRepository
}

func NewImportService(
	entryRepo ImportEntryRepository,
	habitRepo ImportHabitRepository,
	habitLogRepo ImportHabitLogRepository,
	dayContextRepo ImportDayContextRepository,
	summaryRepo ImportSummaryRepository,
	listRepo ImportListRepository,
	listItemRepo ImportListItemRepository,
	goalRepo ImportGoalRepository,
) *ImportService {
	return &ImportService{
		entryRepo:      entryRepo,
		habitRepo:      habitRepo,
		habitLogRepo:   habitLogRepo,
		dayContextRepo: dayContextRepo,
		summaryRepo:    summaryRepo,
		listRepo:       listRepo,
		listItemRepo:   listItemRepo,
		goalRepo:       goalRepo,
	}
}

func (s *ImportService) Import(ctx context.Context, data *domain.ExportData, opts domain.ImportOptions) error {
	if opts.Mode == domain.ImportModeReplace {
		if err := s.clearAllData(ctx); err != nil {
			return err
		}
	}

	for _, entry := range data.Entries {
		if _, err := s.entryRepo.Insert(ctx, entry); err != nil {
			return err
		}
	}

	for _, habit := range data.Habits {
		shouldInsert := true
		if opts.Mode == domain.ImportModeMerge {
			existing, err := s.habitRepo.GetByEntityID(ctx, habit.EntityID)
			if err != nil {
				return err
			}
			shouldInsert = existing == nil
		}

		if shouldInsert {
			if _, err := s.habitRepo.Insert(ctx, habit); err != nil {
				return err
			}
		}
	}

	for _, log := range data.HabitLogs {
		if _, err := s.habitLogRepo.Insert(ctx, log); err != nil {
			return err
		}
	}

	for _, dc := range data.DayContexts {
		if err := s.dayContextRepo.Upsert(ctx, dc); err != nil {
			return err
		}
	}

	for _, summary := range data.Summaries {
		if _, err := s.summaryRepo.Insert(ctx, summary); err != nil {
			return err
		}
	}

	for _, list := range data.Lists {
		shouldInsert := true
		if opts.Mode == domain.ImportModeMerge {
			existing, err := s.listRepo.GetByEntityID(ctx, list.EntityID)
			if err != nil {
				return err
			}
			shouldInsert = existing == nil
		}

		if shouldInsert {
			if _, err := s.listRepo.InsertWithEntityID(ctx, list); err != nil {
				return err
			}
		}
	}

	for _, item := range data.ListItems {
		if _, err := s.listItemRepo.Insert(ctx, item); err != nil {
			return err
		}
	}

	for _, goal := range data.Goals {
		shouldInsert := true
		if opts.Mode == domain.ImportModeMerge {
			existing, err := s.goalRepo.GetByEntityID(ctx, goal.EntityID)
			if err != nil {
				return err
			}
			shouldInsert = existing == nil
		}

		if shouldInsert {
			if _, err := s.goalRepo.Insert(ctx, goal); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *ImportService) clearAllData(ctx context.Context) error {
	if err := s.listItemRepo.DeleteAll(ctx); err != nil {
		return err
	}
	if err := s.listRepo.DeleteAll(ctx); err != nil {
		return err
	}
	if err := s.goalRepo.DeleteAll(ctx); err != nil {
		return err
	}
	if err := s.summaryRepo.DeleteAll(ctx); err != nil {
		return err
	}
	if err := s.dayContextRepo.DeleteAll(ctx); err != nil {
		return err
	}
	if err := s.habitLogRepo.DeleteAll(ctx); err != nil {
		return err
	}
	if err := s.habitRepo.DeleteAll(ctx); err != nil {
		return err
	}
	if err := s.entryRepo.DeleteAll(ctx); err != nil {
		return err
	}
	return nil
}

func (s *ExportService) Export(ctx context.Context, opts domain.ExportOptions) (*domain.ExportData, error) {
	data := &domain.ExportData{
		Version:    domain.ExportVersion,
		ExportedAt: time.Now(),
	}

	var err error

	data.Entries, err = s.entryRepo.GetAll(ctx)
	if opts.DateFrom != nil && opts.DateTo != nil {
		data.Entries, err = s.entryRepo.GetByDateRange(ctx, *opts.DateFrom, *opts.DateTo)
	}
	if err != nil {
		return nil, err
	}
	if data.Entries == nil {
		data.Entries = []domain.Entry{}
	}

	data.Habits, err = s.habitRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if data.Habits == nil {
		data.Habits = []domain.Habit{}
	}

	data.HabitLogs, err = s.habitLogRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if data.HabitLogs == nil {
		data.HabitLogs = []domain.HabitLog{}
	}

	data.DayContexts, err = s.dayContextRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if data.DayContexts == nil {
		data.DayContexts = []domain.DayContext{}
	}

	data.Summaries, err = s.summaryRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if data.Summaries == nil {
		data.Summaries = []domain.Summary{}
	}

	data.Lists, err = s.listRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if data.Lists == nil {
		data.Lists = []domain.List{}
	}

	data.ListItems, err = s.listItemRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if data.ListItems == nil {
		data.ListItems = []domain.ListItem{}
	}

	data.Goals, err = s.goalRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if data.Goals == nil {
		data.Goals = []domain.Goal{}
	}

	return data, nil
}
