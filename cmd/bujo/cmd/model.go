package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/domain"
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

var modelStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show model storage status",
	Long:  `Display disk usage for downloaded models and available space.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		modelsDir := getDefaultModelsDir()
		svc := service.NewModelService(modelsDir)

		models, err := svc.List(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list models: %w", err)
		}

		bold := color.New(color.Bold).SprintFunc()
		dimmed := color.New(color.Faint).SprintFunc()

		var totalSize int64
		var downloadedCount int

		fmt.Printf("%s %s\n", bold("Models directory:"), modelsDir)
		fmt.Println()

		downloaded := []domain.ModelInfo{}
		for _, model := range models {
			if model.IsDownloaded() {
				downloaded = append(downloaded, model)
				totalSize += model.Size
				downloadedCount++
			}
		}

		if downloadedCount == 0 {
			fmt.Println(dimmed("No models downloaded yet."))
			fmt.Println()
			fmt.Println("Download a model with: bujo model pull <name>")
			return nil
		}

		totalGB := float64(totalSize) / (1024 * 1024 * 1024)
		fmt.Printf("%s %.2f GB\n", bold("Total size:"), totalGB)
		fmt.Println()

		fmt.Println(bold("Downloaded models:"))
		for _, model := range downloaded {
			sizeGB := float64(model.Size) / (1024 * 1024 * 1024)
			fmt.Printf("  %-18s  %.2f GB\n", model.Spec, sizeGB)
		}

		return nil
	},
}

var modelRmCmd = &cobra.Command{
	Use:   "rm <model>",
	Short: "Remove a downloaded model",
	Long:  `Remove a downloaded AI model and free up disk space.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		modelName := args[0]

		spec, err := domain.ParseModelSpec(modelName)
		if err != nil {
			return fmt.Errorf("invalid model name: %w", err)
		}

		modelsDir := getDefaultModelsDir()
		svc := service.NewModelService(modelsDir)

		model, err := svc.FindModel(cmd.Context(), spec)
		if err != nil {
			return fmt.Errorf("model not found: %w", err)
		}

		if !model.IsDownloaded() {
			return fmt.Errorf("model %s is not downloaded", spec)
		}

		sizeMB := model.Size / (1024 * 1024)
		sizeGB := float64(model.Size) / (1024 * 1024 * 1024)

		var sizeStr string
		if sizeGB >= 1.0 {
			sizeStr = fmt.Sprintf("%.1f GB", sizeGB)
		} else {
			sizeStr = fmt.Sprintf("%d MB", sizeMB)
		}

		if err := svc.Remove(cmd.Context(), spec); err != nil {
			return fmt.Errorf("failed to remove model: %w", err)
		}

		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("%s Removed %s (freed %s)\n", green("✓"), spec, sizeStr)

		return nil
	},
}

func init() {
	modelCmd.AddCommand(modelListCmd)
	modelCmd.AddCommand(modelStatusCmd)
	modelCmd.AddCommand(modelRmCmd)
	rootCmd.AddCommand(modelCmd)
}

func getDefaultModelsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "models"
	}

	return filepath.Join(home, ".bujo", "models")
}
