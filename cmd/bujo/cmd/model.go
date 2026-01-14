package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/service"
)

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage local AI models",
	Long:  `Download, list, and manage local AI models for offline summarization and Q&A.`,
}

var modelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available models",
	Long:  `List all available AI models, showing which are downloaded and their sizes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		modelsDir := getDefaultModelsDir()
		svc := service.NewModelService(modelsDir)

		models, err := svc.List(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list models: %w", err)
		}

		bold := color.New(color.Bold).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		dimmed := color.New(color.Faint).SprintFunc()

		fmt.Println(bold("Available models:"))
		fmt.Println()

		for _, model := range models {
			modelName := model.Spec.String()
			sizeMB := model.Size / (1024 * 1024)
			sizeGB := float64(model.Size) / (1024 * 1024 * 1024)

			var sizeStr string
			if sizeGB >= 1.0 {
				sizeStr = fmt.Sprintf("%.1f GB", sizeGB)
			} else {
				sizeStr = fmt.Sprintf("%d MB", sizeMB)
			}

			status := ""
			if model.IsDownloaded() {
				status = green("[downloaded]")
			}

			fmt.Printf("  %-18s %-10s  %s  %s\n",
				modelName,
				fmt.Sprintf("(%s)", sizeStr),
				status,
				dimmed(model.Description))
		}

		fmt.Println()
		fmt.Println(dimmed("Download a model with: bujo model pull <name>"))

		return nil
	},
}

func init() {
	modelCmd.AddCommand(modelListCmd)
	rootCmd.AddCommand(modelCmd)
}

func getDefaultModelsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "models"
	}

	return filepath.Join(home, ".bujo", "models")
}
