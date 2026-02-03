package commands

import (
	"time"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var recallCmd = &cobra.Command{
	Use:    "recall <topic>",
	Short:  "Compatibility alias for list --topic <topic>",
	Long:   `Compatibility alias for list --topic <topic>. Prefer dory list --topic for new workflows.`,
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		topic := args[0]
		s := store.New(doryRoot)
		defer s.Close()

		sortKey, _ := cmd.Flags().GetString("sort")
		sortKey = resolveListSort(cmd, sortKey)
		desc, _ := cmd.Flags().GetBool("desc")
		CheckError(validateListSort(sortKey))

		items, err := s.List(topic, "", "", time.Time{}, time.Time{})
		CheckError(err)
		sortListItems(items, sortKey, desc)

		OutputResult(cmd, items, func() { renderListHuman(items) })
	},
}

func init() {
	recallCmd.Flags().String("sort", "id", "Sort by: id, created")
	recallCmd.Flags().Bool("desc", false, "Sort in descending order")
	RootCmd.AddCommand(recallCmd)
}
