package commands

import (
	"fmt"
	"strings"

	"github.com/sibellavia/dory/internal/plugin"
	"github.com/spf13/cobra"
)

var pluginInspectCmd = &cobra.Command{
	Use:   "inspect <name>",
	Short: "Inspect a plugin manifest and status",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		name := args[0]
		plugins, issues, err := plugin.Discover(doryRoot)
		CheckError(err)

		selected := findPluginByName(plugins, name)
		if selected == nil {
			CheckError(fmt.Errorf("plugin %q not found", name))
		}

		result := map[string]interface{}{
			"plugin": selected,
		}
		if len(issues) > 0 {
			result["issues"] = issues
		}

		OutputResult(cmd, result, func() {
			state := "disabled"
			if selected.Enabled {
				state = "enabled"
			}
			fmt.Printf("Name:        %s\n", selected.Name)
			fmt.Printf("Version:     %s\n", selected.Version)
			fmt.Printf("API Version: %s\n", selected.APIVersion)
			fmt.Printf("State:       %s\n", state)
			fmt.Printf("Manifest:    %s\n", pluginPathDisplay(selected.ManifestPath))
			fmt.Printf("Command:     %s\n", strings.Join(selected.Command, " "))
			if selected.Description != "" {
				fmt.Printf("Description: %s\n", selected.Description)
			}
			fmt.Printf("Capabilities: %s\n", capabilitySummary(selected.Capabilities))
			if len(selected.Capabilities.Commands) > 0 {
				fmt.Printf("  Commands: %s\n", strings.Join(selected.Capabilities.Commands, ", "))
			}
			if len(selected.Capabilities.Hooks) > 0 {
				fmt.Printf("  Hooks: %s\n", strings.Join(selected.Capabilities.Hooks, ", "))
			}
			if len(selected.Capabilities.Types) > 0 {
				fmt.Printf("  Types: %s\n", strings.Join(selected.Capabilities.Types, ", "))
			}
			if len(selected.Capabilities.TUI) > 0 {
				fmt.Printf("  TUI: %s\n", strings.Join(selected.Capabilities.TUI, ", "))
			}
			for _, issue := range issues {
				fmt.Printf("Warning: %s (%s)\n", issue.Error, pluginPathDisplay(issue.Path))
			}
		})
	},
}

func init() {
	pluginCmd.AddCommand(pluginInspectCmd)
}
