package commands

import "github.com/spf13/cobra"

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage Dory plugins",
	Long: `Manage project plugins.

Current plugin support includes install/remove lifecycle, discovery, enablement,
inspection, diagnostics, command execution, hooks, custom types, and TUI extension declarations.
Plugins are discovered from .dory/plugins/<plugin>/plugin.yaml and are opt-in.`,
}

func init() {
	RootCmd.AddCommand(pluginCmd)
}
