package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/remarkable"
	"github.com/typingincolor/bujo/internal/domain"
)

var remarkableCmd = &cobra.Command{
	Use:   "remarkable",
	Short: "reMarkable cloud integration (test harness)",
	Long:  `Commands for testing reMarkable cloud API integration. No database required.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var remarkableRegisterCmd = &cobra.Command{
	Use:   "register <code>",
	Short: "Register device with reMarkable cloud using one-time code",
	Long: `Register this device with the reMarkable cloud API.

Get a code from: my.remarkable.com/device/browser/connect
Then run: bujo remarkable register <8-char-code>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		code := args[0]
		client := remarkable.NewClient(remarkable.DefaultAuthHost)

		fmt.Println("Registering with reMarkable cloud...")
		deviceToken, err := client.RegisterDevice(context.Background(), code)
		if err != nil {
			return fmt.Errorf("registration failed: %w", err)
		}

		configPath, err := remarkable.DefaultConfigPath()
		if err != nil {
			return fmt.Errorf("failed to determine config path: %w", err)
		}

		cfg := remarkable.Config{
			DeviceToken: deviceToken,
		}
		if err := remarkable.SaveConfig(configPath, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Device registered. Token saved to %s\n", configPath)
		return nil
	},
}

var remarkableListCmd = &cobra.Command{
	Use:   "list",
	Short: "List documents from reMarkable cloud",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := remarkable.DefaultConfigPath()
		if err != nil {
			return err
		}
		cfg, err := remarkable.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("not registered — run 'bujo remarkable register <code>' first: %w", err)
		}

		client := remarkable.NewClient(remarkable.DefaultAuthHost)
		client.SetSyncHost(remarkable.DefaultSyncHost)

		docs, err := client.ListDocuments(context.Background(), cfg.DeviceToken)
		if err != nil {
			return fmt.Errorf("failed to list documents: %w", err)
		}

		if len(docs) == 0 {
			fmt.Println("No documents found.")
			return nil
		}

		fmt.Printf("%-40s %-10s %-20s %s\n", "NAME", "TYPE", "MODIFIED", "ID")
		for _, doc := range docs {
			fmt.Printf("%-40s %-10s %-20s %s\n", doc.VisibleName, doc.FileType, doc.LastModified, doc.ID)
		}
		return nil
	},
}

var remarkableImportCmd = &cobra.Command{
	Use:   "import <doc-id>",
	Short: "Download notebook, OCR pages, parse bujo entries, print to stdout",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docID := args[0]
		providerName, _ := cmd.Flags().GetString("provider")

		configPath, err := remarkable.DefaultConfigPath()
		if err != nil {
			return err
		}
		cfg, err := remarkable.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("not registered — run 'bujo remarkable register <code>' first: %w", err)
		}

		provider, err := newOCRProvider(providerName)
		if err != nil {
			return err
		}

		client := remarkable.NewClient(remarkable.DefaultAuthHost)
		client.SetSyncHost(remarkable.DefaultSyncHost)

		ctx := context.Background()
		fmt.Fprintf(os.Stderr, "Downloading pages for %s...\n", docID)
		pages, err := client.DownloadPages(ctx, cfg.DeviceToken, docID)
		if err != nil {
			return fmt.Errorf("failed to download pages: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Downloaded %d pages\n", len(pages))

		tmpDir, err := os.MkdirTemp("", "remarkable-import-*")
		if err != nil {
			return err
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		parser := domain.NewTreeParser()

		for i, page := range pages {
			fmt.Fprintf(os.Stderr, "Rendering page %d/%d...\n", i+1, len(pages))
			pngPath, err := remarkable.RenderPageToPNG(tmpDir, page.PageID, page.Data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: render failed for page %s: %v\n", page.PageID, err)
				continue
			}

			fmt.Fprintf(os.Stderr, "OCR page %d/%d (provider: %s)...\n", i+1, len(pages), providerName)
			results, err := provider.RecognizeText(ctx, pngPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: OCR failed for page %s: %v\n", page.PageID, err)
				continue
			}

			reconstructed := remarkable.ReconstructTextWithConfidence(results, remarkable.DefaultConfidenceThreshold)

			fmt.Printf("\n--- Page %d ---\n", i+1)
			fmt.Printf("Reconstructed text:\n%s\n", reconstructed.Text)
			if reconstructed.LowConfidenceCount > 0 {
				fmt.Printf("Low confidence lines: %v (%d total)\n", reconstructed.LowConfidenceLines, reconstructed.LowConfidenceCount)
			}
			if len(reconstructed.ConcatenatedLines) > 0 {
				fmt.Printf("Concatenated lines: %v\n", reconstructed.ConcatenatedLines)
			}
			if len(reconstructed.UncertainLines) > 0 {
				fmt.Printf("Uncertain lines (candidate disagreement): %v\n", reconstructed.UncertainLines)
			}

			text := reconstructed.Text

			entries, err := parser.Parse(text)
			if err != nil {
				fmt.Printf("Parse error: %v\n", err)
				continue
			}

			fmt.Printf("\nParsed %d entries:\n", len(entries))
			for _, e := range entries {
				indent := strings.Repeat("  ", e.Depth)
				fmt.Printf("%s%s %s", indent, e.Type, e.Content)
				if e.Priority != domain.PriorityNone {
					fmt.Printf(" [%s]", e.Priority)
				}
				if len(e.Tags) > 0 {
					fmt.Printf(" tags:%v", e.Tags)
				}
				fmt.Println()
			}
		}
		return nil
	},
}

var remarkableRenderCmd = &cobra.Command{
	Use:   "render <doc-id>",
	Short: "Download notebook pages and render to PNG",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docID := args[0]
		outDir, _ := cmd.Flags().GetString("out-dir")

		configPath, err := remarkable.DefaultConfigPath()
		if err != nil {
			return err
		}
		cfg, err := remarkable.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("not registered — run 'bujo remarkable register <code>' first: %w", err)
		}

		client := remarkable.NewClient(remarkable.DefaultAuthHost)
		client.SetSyncHost(remarkable.DefaultSyncHost)

		fmt.Printf("Downloading pages for %s...\n", docID)
		pages, err := client.DownloadPages(context.Background(), cfg.DeviceToken, docID)
		if err != nil {
			return fmt.Errorf("failed to download pages: %w", err)
		}
		fmt.Printf("Downloaded %d pages\n", len(pages))

		if outDir == "" {
			outDir, err = os.MkdirTemp("", "remarkable-render-*")
			if err != nil {
				return err
			}
		} else {
			if err := os.MkdirAll(outDir, 0755); err != nil {
				return err
			}
		}

		for i, page := range pages {
			fmt.Printf("Rendering page %d/%d (%s)...\n", i+1, len(pages), page.PageID)
			pngPath, err := remarkable.RenderPageToPNG(outDir, page.PageID, page.Data)
			if err != nil {
				return fmt.Errorf("failed to render page %s: %w", page.PageID, err)
			}
			fmt.Printf("  → %s\n", pngPath)
		}

		fmt.Printf("\nPNGs saved to: %s\n", outDir)
		return nil
	},
}

func findOCRTool() string {
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		// Check next to executable (Homebrew: bin/remarkable-ocr)
		if p := filepath.Join(dir, "remarkable-ocr"); fileExists(p) {
			return p
		}
		// Check tools subdirectory
		if p := filepath.Join(dir, "tools", "remarkable-ocr", "remarkable-ocr"); fileExists(p) {
			return p
		}
		if p := filepath.Join(dir, "..", "tools", "remarkable-ocr", "remarkable-ocr"); fileExists(p) {
			return p
		}
	}
	// Check dev mode: relative to working directory
	if cwd, err := os.Getwd(); err == nil {
		if p := filepath.Join(cwd, "tools", "remarkable-ocr", "remarkable-ocr"); fileExists(p) {
			return p
		}
	}
	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func newOCRProvider(providerName string) (remarkable.OCRProvider, error) {
	switch providerName {
	case "apple":
		ocrTool := findOCRTool()
		if ocrTool == "" {
			return nil, fmt.Errorf("OCR tool not found — build with: make ocr")
		}
		return &remarkable.AppleVisionOCR{ToolPath: ocrTool}, nil
	case "google":
		return &remarkable.GoogleVisionOCR{}, nil
	default:
		return nil, fmt.Errorf("unknown OCR provider: %s (supported: apple, google)", providerName)
	}
}

var remarkableOcrCmd = &cobra.Command{
	Use:   "ocr <png-path-or-dir>",
	Short: "Run OCR on PNG(s), output text with bounding boxes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		providerName, _ := cmd.Flags().GetString("provider")

		provider, err := newOCRProvider(providerName)
		if err != nil {
			return err
		}

		info, err := os.Stat(target)
		if err != nil {
			return fmt.Errorf("cannot access %s: %w", target, err)
		}

		var pngFiles []string
		if info.IsDir() {
			entries, err := os.ReadDir(target)
			if err != nil {
				return err
			}
			for _, e := range entries {
				if filepath.Ext(e.Name()) == ".png" {
					pngFiles = append(pngFiles, filepath.Join(target, e.Name()))
				}
			}
		} else {
			pngFiles = []string{target}
		}

		ctx := context.Background()
		for i, png := range pngFiles {
			fmt.Fprintf(os.Stderr, "OCR page %d/%d: %s (provider: %s)\n", i+1, len(pngFiles), filepath.Base(png), providerName)
			results, err := provider.RecognizeText(ctx, png)
			if err != nil {
				return fmt.Errorf("OCR failed on %s: %w", png, err)
			}

			text := remarkable.ReconstructText(results)
			fmt.Printf("--- Page %d ---\n%s\n\n", i+1, text)
		}
		return nil
	},
}

func init() {
	remarkableRenderCmd.Flags().String("out-dir", "", "Output directory for PNGs (default: temp dir)")
	remarkableOcrCmd.Flags().String("provider", "apple", "OCR provider: apple, google")
	remarkableImportCmd.Flags().String("provider", "apple", "OCR provider: apple, google")
	remarkableCmd.AddCommand(remarkableRegisterCmd)
	remarkableCmd.AddCommand(remarkableListCmd)
	remarkableCmd.AddCommand(remarkableImportCmd)
	remarkableCmd.AddCommand(remarkableRenderCmd)
	remarkableCmd.AddCommand(remarkableOcrCmd)
	rootCmd.AddCommand(remarkableCmd)
}
