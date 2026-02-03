package commands

import (
	"fmt"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var topicsCmd = &cobra.Command{
	Use:   "topics",
	Short: "List all topics with counts",
	Long:  `List all topics and the number of items in each.`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		s := store.New(doryRoot)
		defer s.Close()
		topics, err := s.Topics()
		CheckError(err)

		OutputResult(cmd, topics, func() {
			if len(topics) == 0 {
				fmt.Println("No topics found")
				return
			}

			for _, t := range topics {
				fmt.Printf("%-20s  %d items\n", t.Name, t.Count)
			}
		})
	},
}

func init() {
	RootCmd.AddCommand(topicsCmd)
}
