package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj/go-naturaldate"
	"github.com/typingincolor/bujo/internal/adapter/cli"
	"github.com/typingincolor/bujo/internal/service"
)

var (
	inspectFrom string
	inspectTo   string
)

var habitInspectCmd = &cobra.Command{
	Use:   "inspect <habit-name|#id>",
	Short: "Show habit details and log history",
	Long: `Show detailed information about a habit including individual log entries.

By default shows the last 30 days. Use --from and --to to specify a date range.

Examples:
  bujo habit inspect Gym
  bujo habit inspect #1
  bujo habit inspect Gym --from 2025-12-01
  bujo habit inspect Gym --from "last month"
  bujo habit inspect Gym --from 2025-12-01 --to 2025-12-31`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]
		today := time.Now()

		// Default: last 30 days
		from := today.AddDate(0, 0, -30)
		to := today

		if inspectFrom != "" {
			parsed, err := parseInspectDate(inspectFrom)
			if err != nil {
				return err
			}
			from = parsed
		}

		if inspectTo != "" {
			parsed, err := parseInspectDate(inspectTo)
			if err != nil {
				return err
			}
			to = parsed
		}

		var details *service.HabitDetails
		var err error

		if strings.HasPrefix(nameOrID, "#") {
			habitID, parseErr := strconv.ParseInt(nameOrID[1:], 10, 64)
			if parseErr != nil {
				return fmt.Errorf("invalid habit ID: %s", nameOrID)
			}
			details, err = habitService.InspectHabitByID(cmd.Context(), habitID, from, to, today)
		} else {
			details, err = habitService.InspectHabit(cmd.Context(), nameOrID, from, to, today)
		}

		if err != nil {
			return fmt.Errorf("failed to inspect habit: %w", err)
		}

		fmt.Print(cli.RenderHabitInspect(details))
		return nil
	},
}

func init() {
	habitInspectCmd.Flags().StringVar(&inspectFrom, "from", "", "Start date (e.g., '2025-12-01', 'last month')")
	habitInspectCmd.Flags().StringVar(&inspectTo, "to", "", "End date (e.g., '2025-12-31', 'yesterday')")
	habitCmd.AddCommand(habitInspectCmd)
}

func parseInspectDate(s string) (time.Time, error) {
	now := time.Now()

	// Try standard date format first
	if parsed, err := time.Parse("2006-01-02", s); err == nil {
		return parsed, nil
	}

	// Try natural language parsing
	parsed, err := naturaldate.Parse(s, now, naturaldate.WithDirection(naturaldate.Past))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %s", s)
	}

	return parsed, nil
}
