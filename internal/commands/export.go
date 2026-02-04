package commands

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [ids...]",
	Short: "Export knowledge as markdown",
	Long: `Export knowledge items as markdown for inclusion in CLAUDE.md or AGENTS.md.

Examples:
  dory export                      # Export all knowledge
  dory export --tag architecture   # Export by tag
  dory export D-01JX... D-01JY... L-01JX...  # Export specific items
  dory export --append CLAUDE.md   # Append to file`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		topic := resolveTag(cmd, "topic")
		appendFile, _ := cmd.Flags().GetString("append")

		s := store.New(doryRoot)
		defer s.Close()

		var output string
		var err error

		if len(args) > 0 {
			// Export specific items
			output, err = exportItems(s, args)
		} else if topic != "" {
			// Export by topic
			output, err = exportByTopic(s, topic)
		} else {
			// Export all
			output, err = exportAll(s)
		}
		CheckError(err)

		if appendFile != "" {
			// Append to file
			f, err := os.OpenFile(appendFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			CheckError(err)
			defer f.Close()

			_, err = f.WriteString("\n" + output)
			CheckError(err)

			result := map[string]interface{}{
				"status": "appended",
				"file":   appendFile,
			}
			OutputResult(cmd, result, func() {
				fmt.Printf("Appended to %s\n", appendFile)
			})
			return
		}

		if GetOutputFormat(cmd) == "human" {
			fmt.Print(output)
			return
		}

		OutputResult(cmd, map[string]string{"content": output}, func() {})
	},
}

func exportAll(s *store.Store) (string, error) {
	items, err := s.List("", "", "", time.Time{}, time.Time{})
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString("## Project Knowledge\n\n")

	// Group by type
	var lessons, decisions, conventions []store.ListItem
	for _, item := range items {
		switch item.Type {
		case "lesson":
			lessons = append(lessons, item)
		case "decision":
			decisions = append(decisions, item)
		case "convention":
			conventions = append(conventions, item)
		}
	}

	if len(lessons) > 0 {
		buf.WriteString("### Lessons\n\n")
		sort.Slice(lessons, func(i, j int) bool { return lessons[i].ID < lessons[j].ID })
		for _, item := range lessons {
			buf.WriteString(fmt.Sprintf("- **%s** [%s]: %s\n", item.ID, item.Severity, item.Oneliner))
		}
		buf.WriteString("\n")
	}

	if len(decisions) > 0 {
		buf.WriteString("### Decisions\n\n")
		sort.Slice(decisions, func(i, j int) bool { return decisions[i].ID < decisions[j].ID })
		for _, item := range decisions {
			buf.WriteString(fmt.Sprintf("- **%s**: %s\n", item.ID, item.Oneliner))
		}
		buf.WriteString("\n")
	}

	if len(conventions) > 0 {
		buf.WriteString("### Conventions\n\n")
		sort.Slice(conventions, func(i, j int) bool { return conventions[i].ID < conventions[j].ID })
		for _, item := range conventions {
			buf.WriteString(fmt.Sprintf("- **%s**: %s\n", item.ID, item.Oneliner))
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

func exportByTopic(s *store.Store, topic string) (string, error) {
	items, err := s.List(topic, "", "", time.Time{}, time.Time{})
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("## Knowledge: %s\n\n", topic))

	// Group by type
	var lessons, decisions, conventions []store.ListItem
	for _, item := range items {
		switch item.Type {
		case "lesson":
			lessons = append(lessons, item)
		case "decision":
			decisions = append(decisions, item)
		case "convention":
			conventions = append(conventions, item)
		}
	}

	if len(lessons) > 0 {
		buf.WriteString("### Lessons\n\n")
		sort.Slice(lessons, func(i, j int) bool { return lessons[i].ID < lessons[j].ID })
		for _, item := range lessons {
			buf.WriteString(fmt.Sprintf("- **%s** [%s]: %s\n", item.ID, item.Severity, item.Oneliner))
		}
		buf.WriteString("\n")
	}

	if len(decisions) > 0 {
		buf.WriteString("### Decisions\n\n")
		sort.Slice(decisions, func(i, j int) bool { return decisions[i].ID < decisions[j].ID })
		for _, item := range decisions {
			buf.WriteString(fmt.Sprintf("- **%s**: %s\n", item.ID, item.Oneliner))
		}
		buf.WriteString("\n")
	}

	if len(conventions) > 0 {
		buf.WriteString("### Conventions\n\n")
		sort.Slice(conventions, func(i, j int) bool { return conventions[i].ID < conventions[j].ID })
		for _, item := range conventions {
			buf.WriteString(fmt.Sprintf("- **%s**: %s\n", item.ID, item.Oneliner))
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

func exportItems(s *store.Store, ids []string) (string, error) {
	// Get all items and filter
	allItems, err := s.List("", "", "", time.Time{}, time.Time{})
	if err != nil {
		return "", err
	}

	// Build lookup map
	itemMap := make(map[string]store.ListItem)
	for _, item := range allItems {
		itemMap[item.ID] = item
	}

	var buf bytes.Buffer
	buf.WriteString("## Project Knowledge\n\n")

	for _, id := range ids {
		if item, ok := itemMap[id]; ok {
			switch item.Type {
			case "lesson":
				buf.WriteString(fmt.Sprintf("- **%s** [%s]: %s\n", id, item.Severity, item.Oneliner))
			default:
				buf.WriteString(fmt.Sprintf("- **%s**: %s\n", id, item.Oneliner))
			}
		} else {
			buf.WriteString(fmt.Sprintf("- **%s**: (not found)\n", id))
		}
	}
	buf.WriteString("\n")

	return buf.String(), nil
}

func init() {
	exportCmd.Flags().StringP("tag", "T", "", "Export items for a specific tag/category")
	exportCmd.Flags().StringP("topic", "t", "", "Alias for --tag (deprecated)")
	exportCmd.Flags().StringP("append", "a", "", "Append output to file")
	exportCmd.Flags().MarkHidden("topic")
	RootCmd.AddCommand(exportCmd)
}
