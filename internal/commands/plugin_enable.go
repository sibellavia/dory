package commands

import (
	"fmt"

	"github.com/sibellavia/dory/internal/plugin"
	"github.com/spf13/cobra"
)

var pluginEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a plugin for this project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		name := args[0]
		plugins, _, err := plugin.Discover(doryRoot)
		CheckError(err)

		if findPluginByName(plugins, name) == nil {
			CheckError(fmt.Errorf("plugin %q not found in .dory/plugins", name))
		}

		CheckError(plugin.SetPluginEnabled(doryRoot, name, true))
		OutputResult(cmd, map[string]interface{}{
			"status": "enabled",
			"name":   name,
		}, func() {
			fmt.Printf("Enabled plugin %s\n", name)
		})
	},
}

func init() {
	pluginCmd.AddCommand(pluginEnableCmd)
}
