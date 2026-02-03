package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sibellavia/dory/internal/plugin"
	"github.com/sibellavia/dory/internal/store"
	"github.com/sibellavia/dory/templates"
	"github.com/spf13/cobra"
)

var patternCmd = &cobra.Command{
	Use:   "pattern [oneliner]",
	Short: "Record a pattern",
	Long: `Record an established convention or pattern.

If oneliner is provided, creates a quick pattern.
If no oneliner is provided, opens your editor for full content.

Use --body to provide the full markdown content directly:
  dory pattern "Pattern title" --domain mydomain --body "# Full markdown content..."

Use --body - to read markdown content from stdin:
  cat pattern.md | dory pattern "Pattern title" --domain mydomain --body -`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		domain, _ := cmd.Flags().GetString("domain")
		bodyFlag, _ := cmd.Flags().GetString("body")
		refs, _ := cmd.Flags().GetStringSlice("refs")

		if domain == "" {
			CheckError(fmt.Errorf("--domain is required"))
		}

		s := store.New(doryRoot)
		defer s.Close()

		var oneliner, body string

		if len(args) > 0 {
			// Quick mode: oneliner provided
			oneliner = strings.Join(args, " ")

			// Check for body flag
			if bodyFlag == "-" {
				// Read from stdin
				content, err := io.ReadAll(os.Stdin)
				CheckError(err)
				body = string(content)
			} else if bodyFlag != "" {
				body = bodyFlag
			}

		} else {
			// Editor mode
			content, err := openEditor(templates.PatternTemplate)
			CheckError(err)
			if content == "" || content == templates.PatternTemplate {
				CheckError(fmt.Errorf("aborted: no content provided"))
			}
			oneliner, body = parseEditorContent(content)
		}

		runPluginHooks(plugin.HookBeforeCreate, map[string]interface{}{
			"type":     "pattern",
			"oneliner": oneliner,
			"domain":   domain,
			"refs":     refs,
		})

		id, err := s.Pattern(oneliner, domain, body, refs)
		CheckError(err)

		runPluginHooks(plugin.HookAfterCreate, map[string]interface{}{
			"id":       id,
			"type":     "pattern",
			"oneliner": oneliner,
			"domain":   domain,
			"refs":     refs,
		})

		result := map[string]string{
			"id":       id,
			"status":   "created",
			"oneliner": oneliner,
			"domain":   domain,
		}

		OutputResult(cmd, result, func() {
			fmt.Printf("Created %s\n", id)
		})
	},
}

func init() {
	patternCmd.Flags().StringP("domain", "d", "", "Domain for the pattern (required)")
	patternCmd.Flags().StringP("body", "b", "", "Full markdown body content (use - to read from stdin)")
	patternCmd.Flags().StringSliceP("refs", "R", []string{}, "References to other knowledge items (e.g., L-01JX...,D-01JY...)")
	RootCmd.AddCommand(patternCmd)
}
