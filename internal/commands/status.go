package commands

import (
	"fmt"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Update session state",
	Long:  `Update the current session state with goal, progress, blockers, and next steps.`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		goal, _ := cmd.Flags().GetString("goal")
		progress, _ := cmd.Flags().GetString("progress")
		blocker, _ := cmd.Flags().GetString("blocker")
		next, _ := cmd.Flags().GetStringSlice("next")
		workingFiles, _ := cmd.Flags().GetStringSlice("working-file")
		openQuestions, _ := cmd.Flags().GetStringSlice("question")

		s := store.New("")
		err := s.UpdateStatus(goal, progress, blocker, next, workingFiles, openQuestions)
		CheckError(err)

		result := map[string]interface{}{
			"status": "updated",
		}
		if goal != "" {
			result["goal"] = goal
		}
		if progress != "" {
			result["progress"] = progress
		}
		if blocker != "" {
			result["blocker"] = blocker
		}
		if len(next) > 0 {
			result["next"] = next
		}

		OutputResult(cmd, result, func() {
			fmt.Println("Status updated")
		})
	},
}

func init() {
	statusCmd.Flags().StringP("goal", "g", "", "Current goal")
	statusCmd.Flags().StringP("progress", "p", "", "Current progress")
	statusCmd.Flags().StringP("blocker", "b", "", "Current blocker")
	statusCmd.Flags().StringSliceP("next", "n", nil, "Next steps (can be repeated)")
	statusCmd.Flags().StringSlice("working-file", nil, "Working files (can be repeated)")
	statusCmd.Flags().StringSliceP("question", "q", nil, "Open questions (can be repeated)")
	RootCmd.AddCommand(statusCmd)
}
