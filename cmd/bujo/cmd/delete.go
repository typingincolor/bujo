package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an entry",
	Long: `Delete an entry by its ID.

If the entry has children, you will be prompted to choose:
  1. Delete entry and all children
  2. Delete entry and reparent children to grandparent
  3. Cancel

Use --force to skip the prompt and delete with children.

Examples:
  bujo delete 42
  bujo delete 1 --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseEntryID(args[0])
		if err != nil {
			return err
		}

		hasChildren, err := bujoService.HasChildren(cmd.Context(), id)
		if err != nil {
			return fmt.Errorf("failed to check entry: %w", err)
		}

		if hasChildren && !deleteForce {
			fmt.Println("This entry has children. What would you like to do?")
			fmt.Println("  1. Delete entry and all children")
			fmt.Println("  2. Delete entry and reparent children to grandparent")
			fmt.Println("  3. Cancel")
			fmt.Print("Choose [1/2/3]: ")

			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			choice := strings.TrimSpace(input)

			switch choice {
			case "1":
				err = bujoService.DeleteEntry(cmd.Context(), id)
			case "2":
				err = bujoService.DeleteEntryAndReparent(cmd.Context(), id)
			case "3":
				fmt.Fprintln(os.Stderr, "Cancelled")
				return nil
			default:
				return fmt.Errorf("invalid choice: %s", choice)
			}
		} else {
			err = bujoService.DeleteEntry(cmd.Context(), id)
		}

		if err != nil {
			return fmt.Errorf("failed to delete entry: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Deleted entry #%d\n", id)
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Delete without prompting (includes children)")
	rootCmd.AddCommand(deleteCmd)
}
