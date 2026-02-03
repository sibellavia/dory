package commands

import (
	"fmt"

	"github.com/sibellavia/dory/internal/plugin"
	"github.com/spf13/cobra"
)

var pluginInstallCmd = &cobra.Command{
	Use:   "install <path>",
	Short: "Install a plugin from a local directory or plugin.yaml path",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		source := args[0]
		force, _ := cmd.Flags().GetBool("force")
		enable, _ := cmd.Flags().GetBool("enable")

		info, err := plugin.Install(doryRoot, source, plugin.InstallOptions{
			Force: force,
		})
		CheckError(err)

		if enable {
			CheckError(plugin.SetPluginEnabled(doryRoot, info.Name, true))
			info.Enabled = true
		}

		OutputResult(cmd, map[string]interface{}{
			"status": "installed",
			"plugin": info,
			"source": source,
		}, func() {
			state := "disabled"
			if info.Enabled {
				state = "enabled"
			}
			fmt.Printf("Installed plugin %s (%s)\n", info.Name, state)
			fmt.Printf("Location: %s\n", pluginPathDisplay(info.Dir))
		})
	},
}

func init() {
	pluginInstallCmd.Flags().Bool("force", false, "Overwrite existing plugin directory")
	pluginInstallCmd.Flags().Bool("enable", false, "Enable plugin immediately after install")
	pluginCmd.AddCommand(pluginInstallCmd)
}
