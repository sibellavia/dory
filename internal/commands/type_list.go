package commands

import (
	"fmt"

	"github.com/sibellavia/dory/internal/plugin"
	"github.com/spf13/cobra"
)

var typeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List core and plugin-provided knowledge types",
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		customTypes, issues, err := plugin.DiscoverCustomTypes(doryRoot)
		CheckError(err)

		coreTypes := []string{"lesson", "decision", "pattern"}
		payload := map[string]interface{}{
			"core":   coreTypes,
			"custom": customTypes,
		}
		if len(issues) > 0 {
			payload["issues"] = issues
		}

		OutputResult(cmd, payload, func() {
			fmt.Println("Core types:")
			for _, t := range coreTypes {
				fmt.Printf("- %s\n", t)
			}

			if len(customTypes) == 0 {
				fmt.Println("\nCustom types: (none)")
			} else {
				fmt.Println("\nCustom types:")
				for _, t := range customTypes {
					state := "disabled"
					if t.Enabled {
						state = "enabled"
					}
					fmt.Printf("- %s (plugin: %s, %s)\n", t.Name, t.Plugin, state)
				}
			}

			for _, issue := range issues {
				fmt.Printf("Warning: %s (%s)\n", issue.Error, pluginPathDisplay(issue.Path))
			}
		})
	},
}

func init() {
	typeCmd.AddCommand(typeListCmd)
}
