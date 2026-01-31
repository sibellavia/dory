package commands

import (
	"fmt"

	"github.com/simonebellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show full content for an item",
	Long:  `Returns the complete .eng file content for a specific item.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		id := args[0]
		s := store.New("")

		content, err := s.Show(id)
		CheckError(err)

		format := GetOutputFormat(cmd)
		if format == "json" {
			result := map[string]string{
				"id":      id,
				"content": content,
			}
			OutputResult(cmd, result, func() {})
			return
		}

		fmt.Print(content)
	},
}

func init() {
	RootCmd.AddCommand(showCmd)
}
