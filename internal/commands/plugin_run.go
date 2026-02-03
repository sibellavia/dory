package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sibellavia/dory/internal/plugin"
	"github.com/spf13/cobra"
)

const pluginRunMethod = "dory.command.run"

var pluginRunCmd = &cobra.Command{
	Use:   "run <plugin> [command] [args...]",
	Short: "Run a plugin command capability",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		name := args[0]
		requested := args[1:]
		timeout, _ := cmd.Flags().GetDuration("timeout")

		plugins, _, err := plugin.Discover(doryRoot)
		CheckError(err)

		selected := findPluginByName(plugins, name)
		if selected == nil {
			CheckError(fmt.Errorf("plugin %q not found", name))
		}
		if !selected.Enabled {
			CheckError(fmt.Errorf("plugin %q is disabled (run `dory plugin enable %s` first)", name, name))
		}

		commandName, commandArgs, err := resolvePluginCommand(requested, selected.Capabilities.Commands)
		CheckError(err)

		cwd, err := os.Getwd()
		CheckError(err)

		result, stderr, durationMS, err := plugin.Invoke(*selected, pluginRunMethod, map[string]interface{}{
			"api_version": plugin.APIVersionV1,
			"plugin":      selected.Name,
			"command":     commandName,
			"args":        commandArgs,
			"cwd":         cwd,
		}, timeout)
		if err != nil {
			if stderr != "" {
				CheckError(fmt.Errorf("%v: %s", err, stderr))
			}
			CheckError(err)
		}

		payload := map[string]interface{}{
			"plugin":      selected.Name,
			"command":     commandName,
			"args":        commandArgs,
			"duration_ms": durationMS,
			"result":      result,
		}
		if stderr != "" {
			payload["stderr"] = stderr
		}

		OutputResult(cmd, payload, func() {
			if output, ok := result["output"].(string); ok && output != "" {
				fmt.Print(output)
				if !strings.HasSuffix(output, "\n") {
					fmt.Println()
				}
			} else if message, ok := result["message"].(string); ok && message != "" {
				fmt.Println(message)
			} else {
				fmt.Printf("Plugin %s command %s completed in %dms\n", selected.Name, commandName, durationMS)
			}
			if stderr != "" {
				fmt.Fprintf(os.Stderr, "Warning: plugin stderr: %s\n", stderr)
			}
		})
	},
}

func resolvePluginCommand(requested []string, available []string) (string, []string, error) {
	if len(available) == 0 {
		return "", nil, fmt.Errorf("plugin does not expose command capabilities")
	}
	if len(requested) == 0 {
		if len(available) == 1 {
			return available[0], nil, nil
		}
		return "", nil, fmt.Errorf("command is required (available: %s)", strings.Join(available, ", "))
	}

	candidate := requested[0]
	for _, command := range available {
		if command == candidate {
			return candidate, requested[1:], nil
		}
	}

	if len(available) == 1 {
		// If only one command exists, treat requested args as command args.
		return available[0], requested, nil
	}

	return "", nil, fmt.Errorf("unknown command %q (available: %s)", candidate, strings.Join(available, ", "))
}

func init() {
	pluginRunCmd.Flags().Duration("timeout", 5*time.Second, "Plugin command timeout")
	pluginCmd.AddCommand(pluginRunCmd)
}
