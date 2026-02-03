package commands

import (
	"fmt"
	"time"

	"github.com/sibellavia/dory/internal/plugin"
	"github.com/spf13/cobra"
)

var pluginDoctorCmd = &cobra.Command{
	Use:   "doctor [name]",
	Short: "Run health checks for plugins",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		timeout, _ := cmd.Flags().GetDuration("timeout")
		plugins, issues, err := plugin.Discover(doryRoot)
		CheckError(err)

		if len(args) == 1 {
			name := args[0]
			selected := findPluginByName(plugins, name)
			if selected == nil {
				CheckError(fmt.Errorf("plugin %q not found", name))
			}
			plugins = []plugin.PluginInfo{*selected}
		}

		results := make([]plugin.HealthStatus, 0, len(plugins))
		for _, p := range plugins {
			results = append(results, plugin.HealthCheck(p, timeout))
		}

		payload := map[string]interface{}{
			"results": results,
		}
		if len(issues) > 0 {
			payload["issues"] = issues
		}

		OutputResult(cmd, payload, func() {
			if len(results) == 0 {
				fmt.Println("No plugins to check")
			}
			for _, result := range results {
				label := "[ERROR]"
				if result.Status == "ok" {
					label = "[OK]"
				} else if result.Status == "warning" {
					label = "[WARN]"
				}
				msg := result.Message
				if msg == "" {
					msg = result.Error
				} else if result.Error != "" && result.Error != result.Message {
					msg = fmt.Sprintf("%s (%s)", msg, result.Error)
				}
				fmt.Printf("%s %-20s %s\n", label, result.Name, msg)
			}
			for _, issue := range issues {
				fmt.Printf("Warning: %s (%s)\n", issue.Error, pluginPathDisplay(issue.Path))
			}
		})
	},
}

func init() {
	pluginDoctorCmd.Flags().Duration("timeout", 3*time.Second, "Health-check timeout per plugin")
	pluginCmd.AddCommand(pluginDoctorCmd)
}
