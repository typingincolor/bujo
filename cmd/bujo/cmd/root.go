package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
	"github.com/typingincolor/bujo/internal/service"
)

var (
	// Version info set by goreleaser via ldflags
	version = "dev"
	commit  = "none"
	date    = "unknown"

	dbPath  string
	verbose bool

	db           *sql.DB
	bujoService  *service.BujoService
	habitService *service.HabitService
)

var rootCmd = &cobra.Command{
	Use:   "bujo",
	Short: "A command-line Bullet Journal",
	Long:  `bujo is a CLI-based Bullet Journal for rapid task capture, habit tracking, and AI-powered reflections.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip DB initialization for help commands
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}

		var err error
		db, err = sqlite.OpenAndMigrate(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		// Initialize repositories
		entryRepo := sqlite.NewEntryRepository(db)
		dayCtxRepo := sqlite.NewDayContextRepository(db)
		habitRepo := sqlite.NewHabitRepository(db)
		habitLogRepo := sqlite.NewHabitLogRepository(db)
		parser := domain.NewTreeParser()

		// Initialize services
		bujoService = service.NewBujoService(entryRepo, dayCtxRepo, parser)
		habitService = service.NewHabitService(habitRepo, habitLogRepo)

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if db != nil {
			db.Close()
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
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
