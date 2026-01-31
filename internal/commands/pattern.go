package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/simonebellavia/dory/internal/store"
	"github.com/simonebellavia/dory/templates"
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
		summaryFlag, _ := cmd.Flags().GetString("summary")

		if domain == "" {
			fmt.Fprintln(os.Stderr, "Error: --domain is required")
			os.Exit(1)
		}

		s := store.New("")

		var oneliner, summary, body string

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

			if summaryFlag != "" {
				summary = summaryFlag
			}
		} else {
			// Editor mode
			content, err := openEditor(templates.PatternTemplate)
			CheckError(err)
			if content == "" || content == templates.PatternTemplate {
				fmt.Fprintln(os.Stderr, "Aborted: no content provided")
				os.Exit(1)
			}
			oneliner, summary, body = parseEditorContent(content)
		}

		id, err := s.Pattern(oneliner, domain, summary, body)
		CheckError(err)

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
	patternCmd.Flags().String("summary", "", "Short summary for the pattern")
	RootCmd.AddCommand(patternCmd)
}
