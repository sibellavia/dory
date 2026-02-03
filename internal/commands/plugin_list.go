package commands

import (
	"fmt"

	"github.com/sibellavia/dory/internal/plugin"
	"github.com/spf13/cobra"
)

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List discovered plugins",
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		plugins, issues, err := plugin.Discover(doryRoot)
		CheckError(err)

		result := map[string]interface{}{
			"plugins": plugins,
		}
		if len(issues) > 0 {
			result["issues"] = issues
		}

		OutputResult(cmd, result, func() {
			if len(plugins) == 0 {
				fmt.Println("No plugins found in .dory/plugins")
			} else {
				for _, p := range plugins {
					state := "disabled"
					if p.Enabled {
						state = "enabled"
					}
					fmt.Printf("%-20s  %-10s  %-8s  %s\n",
						p.Name,
						p.Version,
						state,
						capabilitySummary(p.Capabilities),
					)
				}
			}
			for _, issue := range issues {
				fmt.Printf("Warning: %s (%s)\n", issue.Error, pluginPathDisplay(issue.Path))
			}
		})
	},
}

func init() {
	pluginCmd.AddCommand(pluginListCmd)
}
