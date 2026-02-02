package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset dory's memory",
	Long: `Clear all knowledge from dory.

By default, keeps project config and state, clears all knowledge items.
Use --full to completely reinitialize (like rm -rf .dory && dory init).`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		full, _ := cmd.Flags().GetBool("full")
		force, _ := cmd.Flags().GetBool("force")

		s := store.NewSingle("")
		defer s.Close()

		// Count items to show user
		items, _ := s.List("", "", "", time.Time{}, time.Time{})
		itemCount := len(items)

		if !force {
			var prompt string
			if full {
				prompt = "This will completely reset dory. Are you sure? [y/N] "
			} else {
				prompt = fmt.Sprintf("This will clear %d knowledge items. Are you sure? [y/N] ", itemCount)
			}
			fmt.Print(prompt)

			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))

			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted")
				return
			}
		}

		if full {
			// Full reset - remove everything and reinitialize
			s.Close() // Close before removing

			// Remove .dory directory
			if err := os.RemoveAll(".dory"); err != nil {
				CheckError(fmt.Errorf("failed to remove .dory: %w", err))
			}

			// Reinitialize
			s2 := store.NewSingle("")
			err := s2.Init("project", "")
			CheckError(err)
			s2.Close()

			fmt.Println("Dory fully reset")
		} else {
			// Partial reset - remove all items, keep state
			for _, item := range items {
				s.Remove(item.ID)
			}
			s.Compact()
			fmt.Printf("Cleared %d knowledge items\n", itemCount)
		}
	},
}

func init() {
	resetCmd.Flags().Bool("full", false, "Full reset (reinitialize completely)")
	resetCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	RootCmd.AddCommand(resetCmd)
}
