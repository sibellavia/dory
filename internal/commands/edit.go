package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sibellavia/dory/internal/doryfile"
	"github.com/sibellavia/dory/internal/fileio"
	"github.com/sibellavia/dory/internal/models"
	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit an item using apply, patch, or inline flags",
	Long: `Update a knowledge item. Supports multiple modes for different workflows:

AGENT-FRIENDLY MODES:

  --apply (like kubectl apply -f -)
    Read full YAML from stdin to update fields:
      echo 'tag: networking
      severity: critical
      oneliner: New title' | dory edit L-abc123 --apply -

  --patch (like kubectl patch)
    Partial JSON update:
      dory edit L-abc123 --patch '{"tag":"networking","severity":"critical"}'

  Inline flags:
      dory edit L-abc123 --tag networking --severity critical

HUMAN MODE:

  No flags opens $EDITOR (not recommended for agents)

APPLY/PATCH FIELDS:
  tag, severity, oneliner, body, refs (array)`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		id := args[0]

		applyFlag, _ := cmd.Flags().GetString("apply")
		patchFlag, _ := cmd.Flags().GetString("patch")

		// Priority 1: Apply mode (full YAML)
		if applyFlag != "" {
			editApply(cmd, id, applyFlag)
			return
		}

		// Priority 2: Patch mode (JSON)
		if patchFlag != "" {
			editPatch(cmd, id, patchFlag)
			return
		}

		// Priority 3: Inline flags
		severity, _ := cmd.Flags().GetString("severity")
		tag := resolveTag(cmd, "topic")
		if tag == "" {
			tag = resolveTag(cmd, "domain")
		}
		topic := tag
		domain := tag
		oneliner, _ := cmd.Flags().GetString("oneliner")
		refs, _ := cmd.Flags().GetStringSlice("refs")

		hasInlineFlags := severity != "" || tag != "" || oneliner != "" || len(refs) > 0

		if hasInlineFlags {
			editInline(cmd, id, severity, topic, domain, oneliner, refs)
			return
		}

		// Priority 4: Editor mode (for humans)
		editWithEditor(cmd, id)
	},
}

// editPatch represents the fields that can be patched
type editPatchData struct {
	Tag      string   `json:"tag" yaml:"tag"`
	Severity string   `json:"severity" yaml:"severity"`
	Oneliner string   `json:"oneliner" yaml:"oneliner"`
	Body     string   `json:"body" yaml:"body"`
	Refs     []string `json:"refs" yaml:"refs"`
}

func editApply(cmd *cobra.Command, id, applyFlag string) {
	var data []byte
	var err error

	if applyFlag == "-" {
		data, err = io.ReadAll(os.Stdin)
		CheckError(err)
	} else {
		data, err = os.ReadFile(applyFlag)
		CheckError(err)
	}

	var patch editPatchData
	CheckError(yaml.Unmarshal(data, &patch))

	applyPatch(cmd, id, &patch)
}

func editPatch(cmd *cobra.Command, id, patchJSON string) {
	var patch editPatchData
	CheckError(json.Unmarshal([]byte(patchJSON), &patch))

	applyPatch(cmd, id, &patch)
}

func applyPatch(cmd *cobra.Command, id string, patch *editPatchData) {
	s := store.New(doryRoot)
	defer s.Close()

	entry, err := s.GetEntry(id)
	CheckError(err)

	var updated []string

	if patch.Tag != "" {
		entry.Topic = patch.Tag
		entry.Domain = patch.Tag
		updated = append(updated, "tag")
	}
	if patch.Severity != "" {
		CheckError(validateSeverityFlag(models.Severity(patch.Severity)))
		entry.Severity = patch.Severity
		updated = append(updated, "severity")
	}
	if patch.Oneliner != "" {
		entry.Oneliner = patch.Oneliner
		updated = append(updated, "oneliner")
	}
	if patch.Body != "" {
		entry.Body = patch.Body
		updated = append(updated, "body")
	}
	if len(patch.Refs) > 0 {
		entry.Refs = patch.Refs
		updated = append(updated, "refs")
	}

	if len(updated) == 0 {
		CheckError(fmt.Errorf("no fields to update in patch"))
	}

	CheckError(s.UpdateEntry(entry))

	result := map[string]interface{}{
		"id":      id,
		"status":  "patched",
		"updated": updated,
	}

	OutputResult(cmd, result, func() {
		fmt.Printf("Patched %s: %s\n", id, strings.Join(updated, ", "))
	})
}

func editInline(cmd *cobra.Command, id, severity, topic, domain, oneliner string, refs []string) {
	s := store.New(doryRoot)
	defer s.Close()

	entry, err := s.GetEntry(id)
	CheckError(err)

	// Update fields if provided
	var updated []string
	if severity != "" {
		CheckError(validateSeverityFlag(models.Severity(severity)))
		entry.Severity = severity
		updated = append(updated, "severity")
	}
	if topic != "" {
		entry.Topic = topic
		updated = append(updated, "topic")
	}
	if domain != "" {
		entry.Domain = domain
		updated = append(updated, "domain")
	}
	if oneliner != "" {
		entry.Oneliner = oneliner
		updated = append(updated, "oneliner")
	}
	if len(refs) > 0 {
		entry.Refs = refs
		updated = append(updated, "refs")
	}

	CheckError(s.UpdateEntry(entry))

	result := map[string]interface{}{
		"id":      id,
		"status":  "updated",
		"updated": updated,
	}

	OutputResult(cmd, result, func() {
		fmt.Printf("Updated %s: %s\n", id, strings.Join(updated, ", "))
	})
}

func editWithEditor(cmd *cobra.Command, id string) {
	s := store.New(doryRoot)
	defer s.Close()

	// Get current content
	content, err := s.Show(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Write to temp file
	tmpfile, err := os.CreateTemp("", "dory-edit-*.md")
	if err != nil {
		CheckError(err)
	}
	tmpPath := tmpfile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpfile.WriteString(content); err != nil {
		tmpfile.Close()
		CheckError(err)
	}
	if err := tmpfile.Close(); err != nil {
		CheckError(err)
	}

	// Open editor
	editor := fileio.GetEditor()
	editorCmd := exec.Command(editor, tmpPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		CheckError(err)
	}

	// Read edited content
	newContent, err := os.ReadFile(tmpPath)
	if err != nil {
		CheckError(err)
	}

	// Check if content changed
	if string(newContent) == content {
		fmt.Println("No changes made")
		return
	}

	// Parse the edited content
	entry, err := parseEditedContent(string(newContent))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing edited content: %v\n", err)
		os.Exit(1)
	}

	// Preserve original ID
	entry.ID = id

	CheckError(s.UpdateEntry(entry))

	result := map[string]string{
		"id":     id,
		"status": "edited",
	}

	OutputResult(cmd, result, func() {
		fmt.Printf("Edited %s\n", id)
	})
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
	if entry.Type == "" {
		return nil, fmt.Errorf("frontmatter must include type")
	}
	if entry.Created.IsZero() {
		return nil, fmt.Errorf("frontmatter must include a valid created timestamp")
	}
	if err := validateItemType(entry.Type); err != nil {
		return nil, err
	}
	if err := validateSeverityFlag(models.Severity(entry.Severity)); err != nil {
		return nil, err
	}

	return entry, nil
}

func init() {
	// Agent-friendly modes
	editCmd.Flags().StringP("apply", "a", "", "Apply YAML from file or stdin (-)")
	editCmd.Flags().StringP("patch", "p", "", "Patch with JSON (e.g., '{\"tag\":\"net\"}')")

	// Inline flags
	editCmd.Flags().StringP("severity", "S", "", "Update severity: critical, high, normal, low")
	editCmd.Flags().StringP("tag", "T", "", "Update tag/category")
	editCmd.Flags().StringP("topic", "t", "", "Alias for --tag (deprecated)")
	editCmd.Flags().StringP("domain", "d", "", "Alias for --tag (deprecated)")
	editCmd.Flags().StringP("oneliner", "o", "", "Update oneliner/title")
	editCmd.Flags().StringSliceP("refs", "R", nil, "Update references (comma-separated, replaces existing)")
	editCmd.Flags().MarkHidden("topic")
	editCmd.Flags().MarkHidden("domain")
	RootCmd.AddCommand(editCmd)
}
