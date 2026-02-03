package commands

import (
	"fmt"

	"github.com/sibellavia/dory/internal/plugin"
	"github.com/spf13/cobra"
)

var pluginDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable a plugin for this project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		name := args[0]
		CheckError(plugin.SetPluginEnabled(doryRoot, name, false))
		OutputResult(cmd, map[string]interface{}{
			"status": "disabled",
			"name":   name,
		}, func() {
			fmt.Printf("Disabled plugin %s\n", name)
		})
	},
}

func init() {
	pluginCmd.AddCommand(pluginDisableCmd)
}
