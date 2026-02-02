package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sibellavia/dory/internal/store"
	"github.com/sibellavia/dory/templates"
	"github.com/spf13/cobra"
)

var decideCmd = &cobra.Command{
	Use:   "decide [oneliner]",
	Short: "Record a decision",
	Long: `Record an architectural or technical decision.

If oneliner is provided, creates a quick decision.
If no oneliner is provided, opens your editor for full content.

Use --body to provide the full markdown content directly:
  dory decide "Decision title" --topic mytopic --body "# Full markdown content..."

Use --body - to read markdown content from stdin:
  cat decision.md | dory decide "Decision title" --topic mytopic --body -`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		topic, _ := cmd.Flags().GetString("topic")
		rationale, _ := cmd.Flags().GetString("rationale")
		bodyFlag, _ := cmd.Flags().GetString("body")
		summaryFlag, _ := cmd.Flags().GetString("summary")
		refs, _ := cmd.Flags().GetStringSlice("refs")

		if topic == "" {
			fmt.Fprintln(os.Stderr, "Error: --topic is required")
			os.Exit(1)
		}

		s := store.NewSingle("")
		defer s.Close()

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
			content, err := openEditor(templates.DecisionTemplate)
			CheckError(err)
			if content == "" || content == templates.DecisionTemplate {
				fmt.Fprintln(os.Stderr, "Aborted: no content provided")
				os.Exit(1)
			}
			oneliner, summary, body = parseEditorContent(content)
		}

		id, err := s.Decide(oneliner, topic, rationale, summary, body, refs)
		CheckError(err)

		result := map[string]string{
			"id":       id,
			"status":   "created",
			"oneliner": oneliner,
			"topic":    topic,
		}

		OutputResult(cmd, result, func() {
			fmt.Printf("Created %s\n", id)
		})
	},
}

func init() {
	decideCmd.Flags().StringP("topic", "t", "", "Topic for the decision (required)")
	decideCmd.Flags().StringP("rationale", "r", "", "Rationale for the decision")
	decideCmd.Flags().StringP("body", "b", "", "Full markdown body content (use - to read from stdin)")
	decideCmd.Flags().String("summary", "", "Short summary for the decision")
	decideCmd.Flags().StringSliceP("refs", "R", []string{}, "References to other knowledge items (e.g., L001,D002)")
	RootCmd.AddCommand(decideCmd)
}
