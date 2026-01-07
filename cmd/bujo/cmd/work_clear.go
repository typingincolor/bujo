package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var workClearDate string

var workClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear location for a day",
	Long: `Clear the location for a specific day.

Defaults to today if no date specified.

Examples:
  bujo work clear
  bujo work clear --date yesterday
  bujo work clear -d "last monday"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDate, err := parseDateOrToday(workClearDate)
		if err != nil {
			return err
		}

		err = bujoService.ClearLocation(cmd.Context(), targetDate)
		if err != nil {
			return fmt.Errorf("failed to clear location: %w", err)
		}

		if workClearDate == "" {
			fmt.Fprintf(os.Stderr, "✓ Location cleared for today\n")
		} else {
			fmt.Fprintf(os.Stderr, "✓ Location cleared for %s\n", targetDate.Format("Jan 2, 2006"))
		}
		return nil
	},
}

func init() {
	workClearCmd.Flags().StringVarP(&workClearDate, "date", "d", "", "Date to clear location for (e.g., 'yesterday', '2026-01-05')")
	workCmd.AddCommand(workClearCmd)
}
