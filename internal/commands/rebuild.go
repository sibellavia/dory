package commands

import (
	"fmt"

	"github.com/simonebellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var rebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuild index from knowledge files",
	Long:  `Scans all .eng files and rebuilds the index.yaml. Useful after manual edits.`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		s := store.New("")
		err := s.Rebuild()
		CheckError(err)

		result := map[string]string{
			"status": "rebuilt",
		}

		OutputResult(cmd, result, func() {
			fmt.Println("Index rebuilt")
		})
	},
}

func init() {
	RootCmd.AddCommand(rebuildCmd)
}
