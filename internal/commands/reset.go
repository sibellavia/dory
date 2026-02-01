package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/simonebellavia/dory/internal/fileio"
	"github.com/simonebellavia/dory/internal/models"
	"github.com/simonebellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset dory's memory",
	Long: `Clear all knowledge from dory.

By default, keeps project config (name, description) and clears knowledge items.
Use --full to completely reinitialize (like rm -rf .dory && dory init).`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		full, _ := cmd.Flags().GetBool("full")
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			var prompt string
			if full {
				prompt = "This will completely reset dory. Are you sure? [y/N] "
			} else {
				prompt = "This will clear all knowledge items. Are you sure? [y/N] "
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

		s := store.New("")

		if full {
			// Full reset - remove everything and reinitialize
			index, _ := s.LoadIndex()
			project := index.Project
			description := index.Description

			// Remove .dory directory
			if err := os.RemoveAll(s.Root); err != nil {
				CheckError(fmt.Errorf("failed to remove .dory: %w", err))
			}

			// Reinitialize
			err := s.Init(project, description)
			CheckError(err)

			fmt.Println("Dory fully reset")
		} else {
			// Partial reset - clear knowledge, keep config and state
			index, err := s.LoadIndex()
			CheckError(err)

			// Preserve project info and state
			project := index.Project
			description := index.Description
			state := index.State

			// Clear knowledge files
			files, err := fileio.ListEngFiles(s.KnowledgeDir)
			CheckError(err)
			for _, f := range files {
				os.Remove(f)
			}

			// Reset index
			newIndex := models.NewIndex(project)
			newIndex.Description = description
			newIndex.State = state

			err = s.SaveIndex(newIndex)
			CheckError(err)

			fmt.Printf("Cleared %d knowledge items\n", len(files))
		}
	},
}

func init() {
	resetCmd.Flags().Bool("full", false, "Full reset (reinitialize completely)")
	resetCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	RootCmd.AddCommand(resetCmd)
}
