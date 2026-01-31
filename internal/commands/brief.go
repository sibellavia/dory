package commands

import (
	"fmt"

	"github.com/simonebellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var briefCmd = &cobra.Command{
	Use:   "brief",
	Short: "Output index and state for agent bootstrap",
	Long:  `Returns the index.yaml and state.yaml content for agent session start.`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		s := store.New("")
		format := GetOutputFormat(cmd)

		if format == "yaml" {
			content, err := s.GetBriefYAML()
			CheckError(err)
			fmt.Print(content)
			return
		}

		if format == "json" {
			index, err := s.LoadIndex()
			CheckError(err)
			state, err := s.LoadState()
			CheckError(err)

			result := map[string]interface{}{
				"index": index,
				"state": state,
			}
			OutputResult(cmd, result, func() {})
			return
		}

		// Human readable
		brief, err := s.Brief()
		CheckError(err)
		fmt.Print(brief)
	},
}

func init() {
	RootCmd.AddCommand(briefCmd)
}
