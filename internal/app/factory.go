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
	Summary         *service.SummaryService
	ChangeDetection *service.ChangeDetectionService
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

	cleanup := func() {
		_ = db.Close()
	}

	services := f.createServices(db)
	return services, cleanup, nil
}

func (f *ServiceFactory) createServices(db *sql.DB) *Services {
	entryRepo := sqlite.NewEntryRepository(db)
	dayCtxRepo := sqlite.NewDayContextRepository(db)
	habitRepo := sqlite.NewHabitRepository(db)
	habitLogRepo := sqlite.NewHabitLogRepository(db)
	listRepo := sqlite.NewListRepository(db)
	listItemRepo := sqlite.NewListItemRepository(db)
	goalRepo := sqlite.NewGoalRepository(db)
	parser := domain.NewTreeParser()

	changeDetectors := []service.ChangeDetector{
		entryRepo,
		dayCtxRepo,
		habitRepo,
		habitLogRepo,
		listRepo,
		listItemRepo,
		goalRepo,
	}

	return &Services{
		DB:              db,
		Bujo:            service.NewBujoService(entryRepo, dayCtxRepo, parser),
		Habit:           service.NewHabitService(habitRepo, habitLogRepo),
		List:            service.NewListService(listRepo, listItemRepo),
		Goal:            service.NewGoalService(goalRepo),
		Stats:           service.NewStatsService(entryRepo, habitRepo, habitLogRepo),
		ChangeDetection: service.NewChangeDetectionService(changeDetectors),
	}
}
