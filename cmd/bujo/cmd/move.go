package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj/go-naturaldate"
	"github.com/typingincolor/bujo/internal/service"
)

var (
	moveParent string
	moveLogged string
	moveRoot   bool
)

var moveCmd = &cobra.Command{
	Use:   "move <id>",
	Short: "Move an entry (change parent or logged date)",
	Long: `Move an entry to a different parent or logged date.

Unlike migrate (which reschedules tasks to future dates), move reorganizes
entries within the journal. Use it to:
  - Change an entry's parent (--parent)
  - Move an entry to be a root entry (--root)
  - Change the day an entry was logged (--logged)

Examples:
  bujo move 42 --parent 10       # Make entry 42 a child of entry 10
  bujo move 42 --root            # Make entry 42 a root entry (no parent)
  bujo move 42 --logged yesterday  # Change logged date to yesterday
  bujo move 42 --parent 10 --logged "last monday"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid entry ID: %s", args[0])
		}

		if moveParent == "" && moveLogged == "" && !moveRoot {
			return fmt.Errorf("at least one of --parent, --root, or --logged is required")
		}

		if moveParent != "" && moveRoot {
			return fmt.Errorf("cannot use both --parent and --root")
		}

		opts := service.MoveOptions{}

		if moveRoot {
			opts.MoveToRoot = &moveRoot
		} else if moveParent != "" {
			parentID, err := strconv.ParseInt(moveParent, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid parent ID: %s", moveParent)
			}
			opts.NewParentID = &parentID
		}

		if moveLogged != "" {
			loggedDate, err := parseMoveDate(moveLogged)
			if err != nil {
				return err
			}
			opts.NewLoggedDate = &loggedDate
		}

		if err := bujoService.MoveEntry(cmd.Context(), id, opts); err != nil {
			return fmt.Errorf("failed to move entry: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Moved entry #%d\n", id)
		return nil
	},
}

func init() {
	moveCmd.Flags().StringVar(&moveParent, "parent", "", "New parent entry ID")
	moveCmd.Flags().BoolVar(&moveRoot, "root", false, "Move entry to root (no parent)")
	moveCmd.Flags().StringVar(&moveLogged, "logged", "", "New logged date (e.g., 'yesterday', '2026-01-05')")
	rootCmd.AddCommand(moveCmd)
}

func parseMoveDate(s string) (time.Time, error) {
	now := time.Now()

	if parsed, err := time.Parse("2006-01-02", s); err == nil {
		return parsed, nil
	}

	parsed, err := naturaldate.Parse(s, now)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %s", s)
	}

	return parsed, nil
}
