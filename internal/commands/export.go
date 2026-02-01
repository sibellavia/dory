package commands

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [ids...]",
	Short: "Export knowledge as markdown",
	Long: `Export knowledge items as markdown for inclusion in CLAUDE.md or AGENTS.md.

Examples:
  dory export                      # Export all knowledge
  dory export --topic architecture # Export by topic
  dory export D001 D002 L001       # Export specific items
  dory export --append CLAUDE.md   # Append to file`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		topic, _ := cmd.Flags().GetString("topic")
		appendFile, _ := cmd.Flags().GetString("append")

		s := store.New("")

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
			f, err := os.OpenFile(appendFile, os.O_APPEND|os.O_WRONLY, 0644)
			CheckError(err)
			defer f.Close()

			_, err = f.WriteString("\n" + output)
			CheckError(err)

			fmt.Printf("Appended to %s\n", appendFile)
		} else {
			fmt.Print(output)
		}
	},
}

func exportAll(s *store.Store) (string, error) {
	index, err := s.LoadIndex()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString("## Project Knowledge\n\n")

	// Collect all IDs
	var ids []string
	for id := range index.Lessons {
		ids = append(ids, id)
	}
	for id := range index.Decisions {
		ids = append(ids, id)
	}
	for id := range index.Patterns {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	// Group by type
	var lessons, decisions, patterns []string
	for _, id := range ids {
		if strings.HasPrefix(id, "L") {
			lessons = append(lessons, id)
		} else if strings.HasPrefix(id, "D") {
			decisions = append(decisions, id)
		} else if strings.HasPrefix(id, "P") {
			patterns = append(patterns, id)
		}
	}

	if len(lessons) > 0 {
		buf.WriteString("### Lessons\n\n")
		for _, id := range lessons {
			entry := index.Lessons[id]
			buf.WriteString(fmt.Sprintf("- **%s** [%s]: %s\n", id, entry.Severity, entry.Oneliner))
		}
		buf.WriteString("\n")
	}

	if len(decisions) > 0 {
		buf.WriteString("### Decisions\n\n")
		for _, id := range decisions {
			entry := index.Decisions[id]
			buf.WriteString(fmt.Sprintf("- **%s**: %s\n", id, entry.Oneliner))
		}
		buf.WriteString("\n")
	}

	if len(patterns) > 0 {
		buf.WriteString("### Patterns\n\n")
		for _, id := range patterns {
			entry := index.Patterns[id]
			buf.WriteString(fmt.Sprintf("- **%s**: %s\n", id, entry.Oneliner))
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

func exportByTopic(s *store.Store, topic string) (string, error) {
	index, err := s.LoadIndex()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("## Knowledge: %s\n\n", topic))

	// Lessons for topic
	var lessons []string
	for id, entry := range index.Lessons {
		if entry.Topic == topic {
			lessons = append(lessons, id)
		}
	}
	sort.Strings(lessons)
	if len(lessons) > 0 {
		buf.WriteString("### Lessons\n\n")
		for _, id := range lessons {
			entry := index.Lessons[id]
			buf.WriteString(fmt.Sprintf("- **%s** [%s]: %s\n", id, entry.Severity, entry.Oneliner))
		}
		buf.WriteString("\n")
	}

	// Decisions for topic
	var decisions []string
	for id, entry := range index.Decisions {
		if entry.Topic == topic {
			decisions = append(decisions, id)
		}
	}
	sort.Strings(decisions)
	if len(decisions) > 0 {
		buf.WriteString("### Decisions\n\n")
		for _, id := range decisions {
			entry := index.Decisions[id]
			buf.WriteString(fmt.Sprintf("- **%s**: %s\n", id, entry.Oneliner))
		}
		buf.WriteString("\n")
	}

	// Patterns for topic (domain)
	var patterns []string
	for id, entry := range index.Patterns {
		if entry.Domain == topic {
			patterns = append(patterns, id)
		}
	}
	sort.Strings(patterns)
	if len(patterns) > 0 {
		buf.WriteString("### Patterns\n\n")
		for _, id := range patterns {
			entry := index.Patterns[id]
			buf.WriteString(fmt.Sprintf("- **%s**: %s\n", id, entry.Oneliner))
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

func exportItems(s *store.Store, ids []string) (string, error) {
	index, err := s.LoadIndex()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString("## Project Knowledge\n\n")

	for _, id := range ids {
		if entry, ok := index.Lessons[id]; ok {
			buf.WriteString(fmt.Sprintf("- **%s** [%s]: %s\n", id, entry.Severity, entry.Oneliner))
		} else if entry, ok := index.Decisions[id]; ok {
			buf.WriteString(fmt.Sprintf("- **%s**: %s\n", id, entry.Oneliner))
		} else if entry, ok := index.Patterns[id]; ok {
			buf.WriteString(fmt.Sprintf("- **%s**: %s\n", id, entry.Oneliner))
		} else {
			buf.WriteString(fmt.Sprintf("- **%s**: (not found)\n", id))
		}
	}
	buf.WriteString("\n")

	return buf.String(), nil
}

func init() {
	exportCmd.Flags().StringP("topic", "t", "", "Export items for a specific topic")
	exportCmd.Flags().StringP("append", "a", "", "Append output to file")
	RootCmd.AddCommand(exportCmd)
}
