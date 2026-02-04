package app

import (
	"context"
	"database/sql"
	"fmt"

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
	InsightsRepo    *sqlite.InsightsRepository
}

type ServiceFactory struct{}

func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{}
}

func (f *ServiceFactory) Create(ctx context.Context, dbPath string) (*Services, func(), error) {
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

	bujoService := service.NewBujoServiceWithLists(entryRepo, dayCtxRepo, parser, listRepo, listItemRepo, entryToListMover)

	return &Services{
		DB:              db,
		Bujo:            bujoService,
		Habit:           service.NewHabitService(habitRepo, habitLogRepo),
		List:            service.NewListService(listRepo, listItemRepo),
		Goal:            service.NewGoalService(goalRepo),
		Stats:           service.NewStatsService(entryRepo, habitRepo, habitLogRepo),
		ChangeDetection: service.NewChangeDetectionService(changeDetectors),
		EditableView:    service.NewEditableViewService(entryRepo, entryToListMover, listRepo),
		InsightsRepo:    sqlite.NewInsightsRepository(insightsDB),
	}
}
