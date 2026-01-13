package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/ai"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
	"github.com/typingincolor/bujo/internal/service"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	dbPath  string
	verbose bool

	db             *sql.DB
	bujoService    *service.BujoService
	habitService   *service.HabitService
	listService    *service.ListService
	goalService    *service.GoalService
	summaryService *service.SummaryService
	statsService   *service.StatsService
)

var rootCmd = &cobra.Command{
	Use:              "bujo",
	Short:            "A command-line Bullet Journal",
	Long:             `bujo is a CLI-based Bullet Journal for rapid task capture, habit tracking, and AI-powered reflections.`,
	TraverseChildren: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}

		var err error
		db, err = sqlite.OpenAndMigrate(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		backupDir := getDefaultBackupDir()
		backupSvc := service.NewBackupService(db, backupDir)
		created, path, err := backupSvc.EnsureRecentBackup(cmd.Context(), 7)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to ensure backup: %v\n", err)
		} else if created {
			fmt.Fprintf(os.Stderr, "Creating backup... %s\n", path)
		}

		entryRepo := sqlite.NewEntryRepository(db)
		dayCtxRepo := sqlite.NewDayContextRepository(db)
		habitRepo := sqlite.NewHabitRepository(db)
		habitLogRepo := sqlite.NewHabitLogRepository(db)
		listRepo := sqlite.NewListRepository(db)
		listItemRepo := sqlite.NewListItemRepository(db)
		goalRepo := sqlite.NewGoalRepository(db)
		parser := domain.NewTreeParser()

		bujoService = service.NewBujoService(entryRepo, dayCtxRepo, parser)
		habitService = service.NewHabitService(habitRepo, habitLogRepo)
		listService = service.NewListService(listRepo, listItemRepo)
		goalService = service.NewGoalService(goalRepo)
		statsService = service.NewStatsService(entryRepo, habitRepo, habitLogRepo)

		summaryRepo := sqlite.NewSummaryRepository(db)
		if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
			geminiClient, err := ai.NewGeminiClient(cmd.Context(), apiKey)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to initialize AI: %v\n", err)
			} else {
				generator := ai.NewGeminiGenerator(geminiClient)
				summaryService = service.NewSummaryService(entryRepo, summaryRepo, generator)
			}
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if db != nil {
			_ = db.Close()
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	_ = godotenv.Load() // .env in current directory
	if home, err := os.UserHomeDir(); err == nil {
		_ = godotenv.Load(filepath.Join(home, ".bujo", ".env"))
	}

	defaultDBPath := getDefaultDBPath()

	rootCmd.PersistentFlags().StringVar(&dbPath, "db-path", defaultDBPath, "Path to the database file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

func getDefaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "bujo.db"
	}

	bujoDir := filepath.Join(home, ".bujo")
	if err := os.MkdirAll(bujoDir, 0755); err != nil {
		return "bujo.db"
	}

	return filepath.Join(bujoDir, "bujo.db")
}
