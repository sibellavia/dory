package commands

import (
	"fmt"

	"github.com/sibellavia/dory/internal/plugin"
	"github.com/spf13/cobra"
)

var pluginTUICmd = &cobra.Command{
	Use:   "tui",
	Short: "List declared plugin TUI extension points",
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		extensions, issues, err := plugin.DiscoverTUIExtensions(doryRoot)
		CheckError(err)

		payload := map[string]interface{}{
			"extensions": extensions,
		}
		if len(issues) > 0 {
			payload["issues"] = issues
		}

		OutputResult(cmd, payload, func() {
			if len(extensions) == 0 {
				fmt.Println("No TUI extensions declared")
			} else {
				for _, ext := range extensions {
					state := "disabled"
					if ext.Enabled {
						state = "enabled"
					}
					fmt.Printf("%-20s  %-16s  %s\n", ext.Name, ext.Plugin, state)
				}
			}
			for _, issue := range issues {
				fmt.Printf("Warning: %s (%s)\n", issue.Error, pluginPathDisplay(issue.Path))
			}
		})
	},
}

func init() {
	pluginCmd.AddCommand(pluginTUICmd)
}
