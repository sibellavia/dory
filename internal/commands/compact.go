package commands

import (
	"fmt"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var compactCmd = &cobra.Command{
	Use:   "compact",
	Short: "Compact the knowledge file",
	Long: `Removes deleted entries from the knowledge file and reclaims disk space.

Dory uses an append-only storage format for reliability. When items are deleted,
they're marked as deleted but remain in the file. Running compact physically
removes deleted entries and rebuilds the file.

This is safe to run at any time - no data will be lost.`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		s := store.New("")
		defer s.Close()

		err := s.Compact()
		CheckError(err)

		result := map[string]string{
			"status": "compacted",
		}

		OutputResult(cmd, result, func() {
			fmt.Println("Knowledge file compacted")
		})
	},
}

func init() {
	RootCmd.AddCommand(compactCmd)
}
