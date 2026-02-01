package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/simonebellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var recallCmd = &cobra.Command{
	Use:   "recall <topic>",
	Short: "Get all knowledge for a topic",
	Long:  `Returns all lessons, decisions, and patterns for a specific topic with summaries.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		topic := args[0]
		s := store.New("")

		format := GetOutputFormat(cmd)

		if format == "json" || format == "yaml" {
			items, err := s.List(topic, "", "", time.Time{}, time.Time{})
			CheckError(err)

			result := map[string]interface{}{
				"topic": topic,
				"items": items,
			}
			OutputResult(cmd, result, func() {})
			return
		}

		content, err := s.Recall(topic)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(content)
	},
}

func init() {
	RootCmd.AddCommand(recallCmd)
}
