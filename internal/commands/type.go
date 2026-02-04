package commands

import "github.com/spf13/cobra"

var typeCmd = &cobra.Command{
	Use:   "type",
	Short: "Manage knowledge types",
	Long: `Discover and use built-in and plugin-provided knowledge types.

Core types remain lesson, decision, and pattern.
Custom types are provided by enabled plugins.`,
}

func init() {
	typeCmd.Hidden = true
	RootCmd.AddCommand(typeCmd)
}
