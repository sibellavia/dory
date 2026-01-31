package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	outputFormat string
)

// RootCmd is the root command for dory
var RootCmd = &cobra.Command{
	Use:   "dory",
	Short: "Knowledge memory for coding agents",
	Long: `Dory - A lightweight knowledge store that gives coding agents
persistent memory across sessions.

Named after the forgetful fish from Finding Nemo, Dory helps agents
remember lessons learned, decisions made, and patterns established.`,
}

func init() {
	RootCmd.PersistentFlags().StringVar(&outputFormat, "format", "", "Output format: json, yaml (default: human-readable)")
	RootCmd.PersistentFlags().Bool("json", false, "Output in JSON format (shorthand for --format=json)")
	RootCmd.PersistentFlags().Bool("yaml", false, "Output in YAML format (shorthand for --format=yaml)")
}

// GetOutputFormat returns the output format from flags
func GetOutputFormat(cmd *cobra.Command) string {
	if f, _ := cmd.Flags().GetBool("json"); f {
		return "json"
	}
	if f, _ := cmd.Flags().GetBool("yaml"); f {
		return "yaml"
	}
	if outputFormat != "" {
		return outputFormat
	}
	return "human"
}

// OutputResult outputs the result in the requested format
func OutputResult(cmd *cobra.Command, data interface{}, humanOutput func()) {
	format := GetOutputFormat(cmd)
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(data)
	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		enc.SetIndent(2)
		enc.Encode(data)
	default:
		humanOutput()
	}
}

// CheckError prints error and exits if err is not nil
func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// RequireStore ensures the dory store exists
func RequireStore() {
	if _, err := os.Stat(".dory/index.yaml"); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "Error: Dory not initialized. Run 'dory init' first.")
		os.Exit(1)
	}
}
