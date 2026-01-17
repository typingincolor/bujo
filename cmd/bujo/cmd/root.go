package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/ai"
	"github.com/typingincolor/bujo/internal/app"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
	"github.com/typingincolor/bujo/internal/service"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	dbPath  string
	verbose bool

	services        *app.Services
	servicesCleanup func()
	summaryService  *service.SummaryService
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

		factory := app.NewServiceFactory()
		var err error
		services, servicesCleanup, err = factory.Create(cmd.Context(), dbPath)
		if err != nil {
			return err
		}

		backupDir := getDefaultBackupDir()
		backupRepo := sqlite.NewBackupRepository(services.DB)
		backupSvc := service.NewBackupService(backupRepo)
		created, path, backupErr := backupSvc.EnsureRecentBackup(cmd.Context(), backupDir, 7)
		if backupErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to ensure backup: %v\n", backupErr)
		} else if created {
			fmt.Fprintf(os.Stderr, "Creating backup... %s\n", path)
		}

		summaryService = initSummaryService(cmd)

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if servicesCleanup != nil {
			servicesCleanup()
		}
	},
}

func initSummaryService(cmd *cobra.Command) *service.SummaryService {
	entryRepo := sqlite.NewEntryRepository(services.DB)
	summaryRepo := sqlite.NewSummaryRepository(services.DB)

	aiClient, err := ai.NewAIClient(cmd.Context())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize AI: %v\n", err)
		return nil
	}

	promptsDir := getDefaultPromptsDir()
	promptLoader := ai.NewPromptLoader(promptsDir)
	if err := promptLoader.EnsureDefaults(cmd.Context()); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create default prompts: %v\n", err)
	}
	generator := ai.NewGeminiGeneratorWithLoader(aiClient, promptLoader)
	return service.NewSummaryService(entryRepo, summaryRepo, generator)
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

func getDefaultPromptsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".bujo", "prompts")
}
