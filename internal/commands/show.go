package commands

import (
	"fmt"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show full content for an item",
	Long:  `Returns the complete content for a specific item.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		id := args[0]
		s := store.New("")
		defer s.Close()

		content, err := s.Show(id)
		CheckError(err)

		format := GetOutputFormat(cmd)
		if format == "human" {
			fmt.Print(content)
			return
		}

		result := map[string]string{
			"id":      id,
			"content": content,
		}
		if format == "json" || format == "yaml" {
			OutputResult(cmd, result, func() {})
			return
		}

		// Fallback for unknown format values.
		OutputResult(cmd, result, func() {
			fmt.Print(content)
		})
	},
}

func init() {
	RootCmd.AddCommand(showCmd)
}
