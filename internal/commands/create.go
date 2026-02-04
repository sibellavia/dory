package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/sibellavia/dory/internal/fileio"
	"github.com/sibellavia/dory/internal/models"
	"github.com/sibellavia/dory/internal/plugin"
	"github.com/sibellavia/dory/internal/store"
	"github.com/sibellavia/dory/templates"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a knowledge item",
	Long: `Create a new knowledge item (lesson, decision, or convention).

Examples:
  dory create "Pool exhausts under load" --tag database --severity critical
  dory create "Use Redis for sessions" --kind decision --tag backend
  dory create "All handlers return {data,error}" --kind convention --tag api
  dory create "Title" --tag api --body "# Details..."
  cat notes.md | dory create "Title" --tag api --body -

Kinds:
  lesson      Something learned (default) - supports --severity
  decision    Architectural/technical choice
  convention  Established standard or pattern`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		kind, _ := cmd.Flags().GetString("kind")
		tag := resolveTag(cmd, "topic")
		severityStr, _ := cmd.Flags().GetString("severity")
		severity := models.Severity(severityStr)
		bodyFlag, _ := cmd.Flags().GetString("body")
		refs, _ := cmd.Flags().GetStringSlice("refs")

		if tag == "" {
			CheckError(fmt.Errorf("--tag is required"))
		}

		// Validate kind
		switch kind {
		case "lesson", "decision", "convention":
			// valid
		default:
			CheckError(fmt.Errorf("invalid --kind %q (use: lesson, decision, convention)", kind))
		}

		// Severity only applies to lessons
		if kind == "lesson" {
			CheckError(validateSeverityFlag(severity))
		} else if severityStr != "normal" && severityStr != "" {
			CheckError(fmt.Errorf("--severity only applies to lessons"))
		}

		s := store.New(doryRoot)
		defer s.Close()

		var oneliner, body string

		if len(args) > 0 {
			// Quick mode: oneliner provided
			oneliner = strings.Join(args, " ")

			// Check for body flag
			if bodyFlag == "-" {
				content, err := io.ReadAll(os.Stdin)
				CheckError(err)
				body = string(content)
			} else if bodyFlag != "" {
				body = bodyFlag
			}
		} else {
			// Editor mode
			template := getTemplateForKind(kind)
			content, err := openEditor(template)
			CheckError(err)
			if content == "" || content == template {
				CheckError(fmt.Errorf("aborted: no content provided"))
			}
			oneliner, body = parseEditorContent(content)
		}

		runPluginHooks(plugin.HookBeforeCreate, map[string]interface{}{
			"type":     kind,
			"oneliner": oneliner,
			"topic":    tag,
			"severity": string(severity),
			"refs":     refs,
		})

		var id string
		var err error

		switch kind {
		case "lesson":
			id, err = s.Learn(oneliner, tag, severity, body, refs)
		case "decision":
			id, err = s.Decide(oneliner, tag, "", body, refs)
		case "convention":
			id, err = s.Convention(oneliner, tag, body, refs)
		}
		CheckError(err)

		runPluginHooks(plugin.HookAfterCreate, map[string]interface{}{
			"id":       id,
			"type":     kind,
			"oneliner": oneliner,
			"topic":    tag,
			"severity": string(severity),
			"refs":     refs,
		})

		result := map[string]interface{}{
			"id":       id,
			"kind":     kind,
			"status":   "created",
			"oneliner": oneliner,
			"tag":      tag,
		}
		if kind == "lesson" {
			result["severity"] = string(severity)
		}

		OutputResult(cmd, result, func() {
			fmt.Printf("Created %s\n", id)
		})
	},
}

func getTemplateForKind(kind string) string {
	switch kind {
	case "decision":
		return templates.DecisionTemplate
	case "convention":
		return templates.ConventionTemplate
	default:
		return templates.LessonTemplate
	}
}

// openEditor opens the user's preferred editor with initial content
func openEditor(initialContent string) (string, error) {
	tmpfile, err := os.CreateTemp("", "dory-*.md")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(initialContent); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}

	editor := fileio.GetEditor()
	editorCmd := exec.Command(editor, tmpfile.Name())
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return "", err
	}

	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// parseEditorContent extracts oneliner and body from editor content
func parseEditorContent(content string) (oneliner, body string) {
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			oneliner = strings.TrimPrefix(line, "# ")
			break
		}
	}

	body = content
	return oneliner, body
}

func init() {
	createCmd.Flags().StringP("kind", "k", "lesson", "Kind: lesson, decision, convention")
	createCmd.Flags().StringP("tag", "T", "", "Tag/category (required)")
	createCmd.Flags().StringP("severity", "S", "normal", "Severity: critical, high, normal, low (lessons only)")
	createCmd.Flags().StringP("body", "b", "", "Full markdown body (use - for stdin)")
	createCmd.Flags().StringSliceP("refs", "R", []string{}, "References (comma-separated, e.g., L-abc123,D-def456)")
	RootCmd.AddCommand(createCmd)
}
