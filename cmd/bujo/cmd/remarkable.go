package cmd

import (
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
		deviceToken, err := client.RegisterDevice(code)
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

		docs, err := client.ListDocuments(cfg.DeviceToken)
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

		configPath, err := remarkable.DefaultConfigPath()
		if err != nil {
			return err
		}
		cfg, err := remarkable.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("not registered — run 'bujo remarkable register <code>' first: %w", err)
		}

		ocrTool := filepath.Join(getToolsDir(), "remarkable-ocr", "remarkable-ocr")
		if _, err := os.Stat(ocrTool); os.IsNotExist(err) {
			return fmt.Errorf("OCR tool not found — build with: swiftc -o %s tools/remarkable-ocr/main.swift -framework Vision -framework AppKit", ocrTool)
		}

		client := remarkable.NewClient(remarkable.DefaultAuthHost)
		client.SetSyncHost(remarkable.DefaultSyncHost)

		fmt.Fprintf(os.Stderr, "Downloading pages for %s...\n", docID)
		pages, err := client.DownloadPages(cfg.DeviceToken, docID)
		if err != nil {
			return fmt.Errorf("failed to download pages: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Downloaded %d pages\n", len(pages))

		tmpDir, err := os.MkdirTemp("", "remarkable-import-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)

		parser := domain.NewTreeParser()

		for i, page := range pages {
			fmt.Fprintf(os.Stderr, "Rendering page %d/%d...\n", i+1, len(pages))
			pngPath, err := remarkable.RenderPageToPNG(tmpDir, page.PageID, page.Data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: render failed for page %s: %v\n", page.PageID, err)
				continue
			}

			fmt.Fprintf(os.Stderr, "OCR page %d/%d...\n", i+1, len(pages))
			results, err := remarkable.RunOCR(ocrTool, pngPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: OCR failed for page %s: %v\n", page.PageID, err)
				continue
			}

			text := remarkable.ReconstructText(results)

			fmt.Printf("\n--- Page %d ---\n", i+1)
			fmt.Printf("Reconstructed text:\n%s\n", text)

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
		pages, err := client.DownloadPages(cfg.DeviceToken, docID)
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

func getToolsDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "tools"
	}
	return filepath.Join(filepath.Dir(exe), "..", "tools")
}

var remarkableOcrCmd = &cobra.Command{
	Use:   "ocr <png-path-or-dir>",
	Short: "Run Apple Vision OCR on PNG(s), output text with bounding boxes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

		ocrTool := filepath.Join(getToolsDir(), "remarkable-ocr", "remarkable-ocr")
		if _, err := os.Stat(ocrTool); os.IsNotExist(err) {
			return fmt.Errorf("OCR tool not found at %s — build with: swiftc -o %s tools/remarkable-ocr/main.swift -framework Vision -framework AppKit", ocrTool, ocrTool)
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

		for i, png := range pngFiles {
			fmt.Fprintf(os.Stderr, "OCR page %d/%d: %s\n", i+1, len(pngFiles), filepath.Base(png))
			results, err := remarkable.RunOCR(ocrTool, png)
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
	remarkableCmd.AddCommand(remarkableRegisterCmd)
	remarkableCmd.AddCommand(remarkableListCmd)
	remarkableCmd.AddCommand(remarkableImportCmd)
	remarkableCmd.AddCommand(remarkableRenderCmd)
	remarkableCmd.AddCommand(remarkableOcrCmd)
	rootCmd.AddCommand(remarkableCmd)
}
