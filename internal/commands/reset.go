package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sibellavia/dory/internal/doryfile"
	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

		s := store.New("")
		defer s.Close()

		// Count items to show user
		items, err := s.List("", "", "", time.Time{}, time.Time{})
		CheckError(err)
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
			answer, err := reader.ReadString('\n')
			CheckError(err)
			answer = strings.TrimSpace(strings.ToLower(answer))

			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted")
				return
			}
		}

		if full {
			projectName := "project"
			description := ""
			if dump, err := s.DumpIndex(); err == nil {
				var index doryfile.Index
				if err := yaml.Unmarshal([]byte(dump), &index); err == nil {
					if index.Project != "" {
						projectName = index.Project
					}
					description = index.Description
				}
			}

			// Full reset - remove everything and reinitialize
			s.Close() // Close before removing

			// Remove .dory directory
			if err := os.RemoveAll(".dory"); err != nil {
				CheckError(fmt.Errorf("failed to remove .dory: %w", err))
			}

			// Reinitialize
			s2 := store.New("")
			err := s2.Init(projectName, description)
			CheckError(err)
			s2.Close()

			OutputResult(cmd, map[string]interface{}{
				"status":      "reset",
				"full":        true,
				"project":     projectName,
				"description": description,
			}, func() {
				fmt.Println("Dory fully reset")
			})
		} else {
			// Partial reset - remove all items, keep state
			for _, item := range items {
				CheckError(s.Remove(item.ID))
			}
			CheckError(s.Compact())
			OutputResult(cmd, map[string]interface{}{
				"status":        "cleared",
				"cleared_items": itemCount,
			}, func() {
				fmt.Printf("Cleared %d knowledge items\n", itemCount)
			})
		}
	},
}

func init() {
	resetCmd.Flags().Bool("full", false, "Full reset (reinitialize completely)")
	resetCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	RootCmd.AddCommand(resetCmd)
}
