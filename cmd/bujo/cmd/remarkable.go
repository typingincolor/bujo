package cmd

import (
	"fmt"
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
	Short: "Download document, extract text, parse entries, print to stdout",
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

		client := remarkable.NewClient(remarkable.DefaultAuthHost)
		client.SetSyncHost(remarkable.DefaultSyncHost)

		fmt.Printf("Downloading document %s...\n", docID)
		data, err := client.DownloadDocument(cfg.DeviceToken, docID)
		if err != nil {
			return fmt.Errorf("failed to download: %w", err)
		}
		fmt.Printf("Downloaded %d bytes\n", len(data))

		manifest, err := remarkable.ListZIPContents(data)
		if err != nil {
			return fmt.Errorf("failed to read ZIP: %w", err)
		}
		fmt.Println("\nZIP contents:")
		for _, entry := range manifest {
			fmt.Printf("  %s\n", entry)
		}

		texts, err := remarkable.ExtractTextFromZIP(data)
		if err != nil {
			return fmt.Errorf("failed to extract text: %w", err)
		}

		if len(texts) == 0 {
			fmt.Println("\nNo text files found in ZIP. The document may not have been converted to text.")
			fmt.Println("On your reMarkable, select the page → Convert to text, then sync.")
			return nil
		}

		parser := domain.NewTreeParser()
		for i, text := range texts {
			fmt.Printf("\n--- Page %d ---\n", i+1)
			fmt.Printf("Raw text:\n%s\n", text)

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

func init() {
	remarkableCmd.AddCommand(remarkableRegisterCmd)
	remarkableCmd.AddCommand(remarkableListCmd)
	remarkableCmd.AddCommand(remarkableImportCmd)
	rootCmd.AddCommand(remarkableCmd)
}
