package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	outputFormat string
	agentMode    bool
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
	RootCmd.PersistentFlags().BoolVar(&agentMode, "agent", false, "Agent mode: machine-oriented defaults (YAML output, no interactive prompts)")

	// Hide the auto-generated completion command
	RootCmd.CompletionOptions.HiddenDefaultCmd = true
}

// GetOutputFormat returns the output format from flags
func GetOutputFormat(cmd *cobra.Command) string {
	jsonFlag, _ := cmd.Flags().GetBool("json")
	yamlFlag, _ := cmd.Flags().GetBool("yaml")
	if jsonFlag && yamlFlag {
		CheckError(fmt.Errorf("cannot use --json and --yaml together"))
		return "human"
	}

	if jsonFlag {
		return "json"
	}
	if yamlFlag {
		return "yaml"
	}
	if outputFormat != "" {
		switch outputFormat {
		case "json", "yaml":
			return outputFormat
		default:
			CheckError(fmt.Errorf("invalid --format value %q (expected: json, yaml)", outputFormat))
			return "human"
		}
	}
	if agentMode {
		return "yaml"
	}
	return "human"
}

func resolveDoryRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	for {
		root := filepath.Join(dir, ".dory")
		indexPath := filepath.Join(root, "index.yaml")
		if info, err := os.Stat(indexPath); err == nil && !info.IsDir() {
			return root, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

func requireInteractive(force bool, flagName string) {
	if agentMode && !force {
		CheckError(fmt.Errorf("%s is required in --agent mode", flagName))
	}
	if force {
		return
	}
	info, err := os.Stdin.Stat()
	if err != nil {
		return
	}
	// If stdin is piped/non-tty, prompt-based commands would hang.
	if info.Mode()&os.ModeCharDevice == 0 {
		CheckError(fmt.Errorf("interactive confirmation requires a TTY; pass %s", flagName))
	}
}

// OutputResult outputs the result in the requested format
func OutputResult(cmd *cobra.Command, data interface{}, humanOutput func()) {
	format := GetOutputFormat(cmd)
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		CheckError(enc.Encode(data))
	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		enc.SetIndent(2)
		CheckError(enc.Encode(data))
		CheckError(enc.Close())
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
	root, err := resolveDoryRoot(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Dory not initialized. Run 'dory init' first.")
		os.Exit(1)
	}
	doryRoot = root
}

// resolveTag returns the tag value, checking --tag first then falling back to
// the legacy flag (--topic or --domain) for backwards compatibility.
func resolveTag(cmd *cobra.Command, legacyFlag string) string {
	tag, _ := cmd.Flags().GetString("tag")
	if tag != "" {
		return tag
	}
	legacy, _ := cmd.Flags().GetString(legacyFlag)
	return legacy
}
