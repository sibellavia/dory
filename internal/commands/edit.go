package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sibellavia/dory/internal/doryfile"
	"github.com/sibellavia/dory/internal/fileio"
	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit an item in your editor",
	Long:  `Opens the item in your $EDITOR. After editing, the item is updated in the knowledge store.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		id := args[0]
		s := store.NewSingle("")

		// Get current content
		content, err := s.Show(id)
		if err != nil {
			s.Close()
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Write to temp file
		tmpfile, err := os.CreateTemp("", "dory-edit-*.md")
		if err != nil {
			s.Close()
			CheckError(err)
		}
		tmpPath := tmpfile.Name()
		defer os.Remove(tmpPath)

		if _, err := tmpfile.WriteString(content); err != nil {
			tmpfile.Close()
			s.Close()
			CheckError(err)
		}
		tmpfile.Close()

		// Open editor
		editor := fileio.GetEditor()
		editorCmd := exec.Command(editor, tmpPath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		if err := editorCmd.Run(); err != nil {
			s.Close()
			CheckError(err)
		}

		// Read edited content
		newContent, err := os.ReadFile(tmpPath)
		if err != nil {
			s.Close()
			CheckError(err)
		}

		// Check if content changed
		if string(newContent) == content {
			s.Close()
			fmt.Println("No changes made")
			return
		}

		// Parse the edited content
		entry, err := parseEditedContent(string(newContent))
		if err != nil {
			s.Close()
			fmt.Fprintf(os.Stderr, "Error parsing edited content: %v\n", err)
			os.Exit(1)
		}

		// Preserve original ID
		entry.ID = id

		// Remove old entry and add new one
		s.Remove(id)

		// Re-add with the same ID (need direct access to doryfile)
		// For now, close and reopen to add
		s.Close()

		// Reopen and add the entry directly
		df, err := doryfile.Open(".dory")
		if err != nil {
			CheckError(err)
		}
		defer df.Close()

		if err := df.Append(entry); err != nil {
			CheckError(err)
		}

		result := map[string]string{
			"id":     id,
			"status": "edited",
		}

		OutputResult(cmd, result, func() {
			fmt.Printf("Edited %s\n", id)
		})
	},
}

func parseEditedContent(content string) (*doryfile.Entry, error) {
	// Parse YAML frontmatter and body
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("missing YAML frontmatter")
	}

	rest := content[4:]
	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		return nil, fmt.Errorf("invalid frontmatter format")
	}

	yamlContent := rest[:endIdx]
	body := strings.TrimPrefix(rest[endIdx+4:], "\n")

	var frontmatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &frontmatter); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	entry := &doryfile.Entry{
		Body: body,
	}

	if v, ok := frontmatter["id"].(string); ok {
		entry.ID = v
	}
	if v, ok := frontmatter["type"].(string); ok {
		entry.Type = v
	}
	if v, ok := frontmatter["oneliner"].(string); ok {
		entry.Oneliner = v
	}
	if v, ok := frontmatter["topic"].(string); ok {
		entry.Topic = v
	}
	if v, ok := frontmatter["domain"].(string); ok {
		entry.Domain = v
	}
	if v, ok := frontmatter["severity"].(string); ok {
		entry.Severity = v
	}
	if v, ok := frontmatter["refs"].([]interface{}); ok {
		for _, r := range v {
			if s, ok := r.(string); ok {
				entry.Refs = append(entry.Refs, s)
			}
		}
	}
	if v, ok := frontmatter["created"].(string); ok {
		// Try RFC3339 first, then other common formats
		for _, format := range []string{time.RFC3339, "2006-01-02T15:04:05-07:00", "2006-01-02"} {
			if t, err := time.Parse(format, v); err == nil {
				entry.Created = t
				break
			}
		}
	}

	return entry, nil
}

func init() {
	RootCmd.AddCommand(editCmd)
}
