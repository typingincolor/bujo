package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/cli"
	"github.com/typingincolor/bujo/internal/service"
)

var (
	habitShowFrom string
	habitShowTo   string
)

var habitShowCmd = &cobra.Command{
	Use:   "show <habit-name|#id>",
	Short: "Show habit details and log history",
	Long: `Show detailed information about a habit including individual log entries.

By default shows the last 30 days. Use --from and --to to specify a date range.

Examples:
  bujo habit show Gym
  bujo habit show #1
  bujo habit show Gym --from 2025-12-01
  bujo habit show Gym --from "last month"
  bujo habit show Gym --from 2025-12-01 --to 2025-12-31`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, id, isID, err := parseHabitNameOrID(args[0])
		if err != nil {
			return err
		}

		if !isID && isPureNumber(args[0]) {
			fmt.Printf("'%s' looks like an ID. Did you mean to use #%s? [y/N]: ", args[0], args[0])
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			confirm := strings.TrimSpace(strings.ToLower(input))

			if confirm == "y" || confirm == "yes" {
				id, _ = strconv.ParseInt(args[0], 10, 64)
				isID = true
			}
		}

		today := time.Now()

		from := today.AddDate(0, 0, -30)
		to := today

		if habitShowFrom != "" {
			parsed, err := parsePastDate(habitShowFrom)
			if err != nil {
				return err
			}
			from = parsed
		}

		if habitShowTo != "" {
			parsed, err := parsePastDate(habitShowTo)
			if err != nil {
				return err
			}
			to = parsed
		}

		if err := validateDateRange(from, to); err != nil {
			return err
		}

		var details *service.HabitDetails

		if isID {
			details, err = services.Habit.InspectHabitByID(cmd.Context(), id, from, to, today)
		} else {
			details, err = services.Habit.InspectHabit(cmd.Context(), name, from, to, today)
		}

		if err != nil {
			return fmt.Errorf("failed to show habit: %w", err)
		}

		fmt.Print(cli.RenderHabitInspect(details))
		return nil
	},
}

func init() {
	habitShowCmd.Flags().StringVar(&habitShowFrom, "from", "", "Start date (e.g., '2025-12-01', 'last month')")
	habitShowCmd.Flags().StringVar(&habitShowTo, "to", "", "End date (e.g., '2025-12-31', 'yesterday')")
	habitCmd.AddCommand(habitShowCmd)
}
