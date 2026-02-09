package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
	"github.com/typingincolor/bujo/internal/service"
)

type Services struct {
	DB              *sql.DB
	Bujo            *service.BujoService
	Habit           *service.HabitService
	List            *service.ListService
	Goal            *service.GoalService
	Stats           *service.StatsService
	ChangeDetection *service.ChangeDetectionService
	EditableView    *service.EditableViewService
	Backup          *service.BackupService
	InsightsRepo    *sqlite.InsightsRepository
}

func DefaultBackupDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "backups"
	}
	return filepath.Join(home, ".bujo", "backups")
}

type CreateOption func(*createOptions)

type createOptions struct {
	backupDir string
}

func WithBackupDir(dir string) CreateOption {
	return func(o *createOptions) {
		o.backupDir = dir
	}
}

type ServiceFactory struct{}

func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{}
}

func (f *ServiceFactory) Create(ctx context.Context, dbPath string, opts ...CreateOption) (*Services, func(), error) {
	options := createOptions{
		backupDir: DefaultBackupDir(),
	}
	for _, opt := range opts {
		opt(&options)
	}

	db, err := sqlite.OpenAndMigrate(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	insightsDB, _ := OpenInsightsDB(DefaultInsightsDBPath())

	cleanup := func() {
		_ = db.Close()
		if insightsDB != nil {
			_ = insightsDB.Close()
		}
	}

	services := f.createServices(db, insightsDB)

	if options.backupDir != "" {
		_, _, _ = services.Backup.EnsureRecentBackup(ctx, options.backupDir, 7)
	}

	return services, cleanup, nil
}

func (f *ServiceFactory) createServices(db *sql.DB, insightsDB *sql.DB) *Services {
	entryRepo := sqlite.NewEntryRepository(db)
	dayCtxRepo := sqlite.NewDayContextRepository(db)
	habitRepo := sqlite.NewHabitRepository(db)
	habitLogRepo := sqlite.NewHabitLogRepository(db)
	listRepo := sqlite.NewListRepository(db)
	listItemRepo := sqlite.NewListItemRepository(db)
	goalRepo := sqlite.NewGoalRepository(db)
	entryToListMover := sqlite.NewEntryToListMover(db)
	parser := domain.NewTreeParser()

	changeDetectors := []domain.ChangeDetector{
		entryRepo,
		dayCtxRepo,
		habitRepo,
		habitLogRepo,
		listRepo,
		listItemRepo,
		goalRepo,
	}

	tagRepo := sqlite.NewTagRepository(db)
	mentionRepo := sqlite.NewMentionRepository(db)
	backupRepo := sqlite.NewBackupRepository(db)

	bujoService := service.NewBujoServiceWithLists(entryRepo, dayCtxRepo, parser, listRepo, listItemRepo, entryToListMover, tagRepo, mentionRepo)

	return &Services{
		DB:              db,
		Bujo:            bujoService,
		Habit:           service.NewHabitService(habitRepo, habitLogRepo),
		List:            service.NewListService(listRepo, listItemRepo),
		Goal:            service.NewGoalService(goalRepo),
		Stats:           service.NewStatsService(entryRepo, habitRepo, habitLogRepo),
		ChangeDetection: service.NewChangeDetectionService(changeDetectors),
		EditableView:    service.NewEditableViewService(entryRepo, entryToListMover, listRepo, tagRepo, mentionRepo),
		Backup:          service.NewBackupService(backupRepo),
		InsightsRepo:    sqlite.NewInsightsRepository(insightsDB),
	}
}
