package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/sibellavia/dory/internal/fileio"
	"github.com/sibellavia/dory/internal/models"
	"github.com/sibellavia/dory/internal/store"
	"github.com/sibellavia/dory/templates"
	"github.com/spf13/cobra"
)

var learnCmd = &cobra.Command{
	Use:   "learn [oneliner]",
	Short: "Add a new lesson",
	Long: `Record something you learned (often the hard way).

If oneliner is provided, creates a quick lesson.
If no oneliner is provided, opens your editor for full content.

Use --body to provide the full markdown content directly:
  dory learn "Lesson title" --topic mytopic --body "# Full markdown content..."

Use --body - to read markdown content from stdin:
  cat lesson.md | dory learn "Lesson title" --topic mytopic --body -`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		topic, _ := cmd.Flags().GetString("topic")
		severityStr, _ := cmd.Flags().GetString("severity")
		severity := models.Severity(severityStr)
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
			content, err := openEditor(templates.LessonTemplate)
			CheckError(err)
			if content == "" || content == templates.LessonTemplate {
				fmt.Fprintln(os.Stderr, "Aborted: no content provided")
				os.Exit(1)
			}
			oneliner, summary, body = parseEditorContent(content)
		}

		id, err := s.Learn(oneliner, topic, severity, summary, body, refs)
		CheckError(err)

		result := map[string]string{
			"id":       id,
			"status":   "created",
			"oneliner": oneliner,
			"topic":    topic,
			"severity": string(severity),
		}

		OutputResult(cmd, result, func() {
			fmt.Printf("Created %s\n", id)
		})
	},
}

func init() {
	learnCmd.Flags().StringP("topic", "t", "", "Topic for the lesson (required)")
	learnCmd.Flags().StringP("severity", "s", "normal", "Severity level: critical, high, normal, low")
	learnCmd.Flags().StringP("body", "b", "", "Full markdown body content (use - to read from stdin)")
	learnCmd.Flags().String("summary", "", "Short summary for the lesson")
	learnCmd.Flags().StringSliceP("refs", "R", []string{}, "References to other knowledge items (e.g., L001,D002)")
	RootCmd.AddCommand(learnCmd)
}

// openEditor opens the user's preferred editor with initial content
func openEditor(initialContent string) (string, error) {
	// Create temp file
	tmpfile, err := os.CreateTemp("", "dory-*.md")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	// Write initial content
	if _, err := tmpfile.WriteString(initialContent); err != nil {
		return "", err
	}
	tmpfile.Close()

	// Open editor
	editor := fileio.GetEditor()
	editorCmd := exec.Command(editor, tmpfile.Name())
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return "", err
	}

	// Read content back
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// parseEditorContent extracts oneliner, summary, and body from editor content
func parseEditorContent(content string) (oneliner, summary, body string) {
	lines := strings.Split(content, "\n")

	// Find the first heading for oneliner
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			oneliner = strings.TrimPrefix(line, "# ")
			break
		}
	}

	// The body is the full content
	body = content

	return oneliner, "", body
}
