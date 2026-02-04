package commands

import (
	"fmt"
	"sort"
	"time"

	"github.com/sibellavia/dory/internal/models"
	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List items or tags",
	Long: `List all knowledge items (core and custom types) with optional filters.

Use --tags to list all tags with item counts instead of items.

Examples:
  dory list                      # List all items
  dory list --tag database       # Filter by tag
  dory list --type lesson        # Filter by type
  dory list --tags               # List all tags with counts`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		showTags, _ := cmd.Flags().GetBool("tags")

		s := store.New(doryRoot)
		defer s.Close()

		// --tags mode: show tags with counts
		if showTags {
			tags, err := s.Topics()
			CheckError(err)
			OutputResult(cmd, tags, func() { renderTagsHuman(tags) })
			return
		}

		// Default: list items
		topic := resolveTag(cmd, "topic")
		itemType, _ := cmd.Flags().GetString("type")
		severityStr, _ := cmd.Flags().GetString("severity")
		severity := models.Severity(severityStr)
		sinceStr, _ := cmd.Flags().GetString("since")
		untilStr, _ := cmd.Flags().GetString("until")
		sortKey, _ := cmd.Flags().GetString("sort")
		sortKey = resolveListSort(cmd, sortKey)
		desc, _ := cmd.Flags().GetBool("desc")

		CheckError(validateItemType(itemType))
		CheckError(validateSeverityFlag(severity))
		CheckError(validateListSort(sortKey))

		since, err := parseDateFlag(sinceStr, "--since")
		CheckError(err)
		until, err := parseDateFlag(untilStr, "--until")
		CheckError(err)
		if !until.IsZero() {
			// Include the full day for date-only filters.
			until = until.Add(24*time.Hour - time.Nanosecond)
		}
		if !since.IsZero() && !until.IsZero() && since.After(until) {
			CheckError(fmt.Errorf("--since must be earlier than or equal to --until"))
		}

		items, err := s.List(topic, itemType, severity, since, until)
		CheckError(err)
		sortListItems(items, sortKey, desc)

		OutputResult(cmd, items, func() { renderListHuman(items) })
	},
}

func init() {
	listCmd.Flags().Bool("tags", false, "List all tags with item counts")
	listCmd.Flags().StringP("tag", "T", "", "Filter by tag/category")
	listCmd.Flags().StringP("topic", "t", "", "Alias for --tag (deprecated)")
	listCmd.Flags().String("type", "", "Filter by type (e.g. lesson, decision, pattern, or plugin custom type)")
	listCmd.Flags().StringP("severity", "S", "", "Filter by severity: critical, high, normal, low")
	listCmd.Flags().String("since", "", "Show items created on or after date (YYYY-MM-DD)")
	listCmd.Flags().String("until", "", "Show items created on or before date (YYYY-MM-DD)")
	listCmd.Flags().StringP("sort", "s", "id", "Sort by: id, created")
	listCmd.Flags().Bool("desc", false, "Sort in descending order")
	listCmd.Flags().MarkHidden("topic")
	RootCmd.AddCommand(listCmd)
}

func validateListSort(sortKey string) error {
	switch sortKey {
	case "", "id", "created":
		return nil
	default:
		return fmt.Errorf("invalid --sort value %q (expected: id, created)", sortKey)
	}
}

func resolveListSort(cmd *cobra.Command, sortKey string) string {
	if sortKey == "" {
		sortKey = "id"
	}
	if agentMode && !cmd.Flags().Changed("sort") {
		// Agent mode defaults to chronological ordering for easier replay.
		return "created"
	}
	return sortKey
}

func renderListHuman(items []store.ListItem) {
	if len(items) == 0 {
		fmt.Println("No items found")
		return
	}

	for _, item := range items {
		topicStr := item.Topic
		if topicStr == "" {
			topicStr = item.Domain
		}

		severityIndicator := ""
		if item.Severity != "" {
			switch item.Severity {
			case models.SeverityCritical:
				severityIndicator = " [CRITICAL]"
			case models.SeverityHigh:
				severityIndicator = " [HIGH]"
			}
		}

		fmt.Printf("%s  %-8s  %-15s  %s%s\n",
			item.ID,
			item.Type,
			topicStr,
			item.Oneliner,
			severityIndicator)
	}
}

func renderTagsHuman(tags []store.TopicInfo) {
	if len(tags) == 0 {
		fmt.Println("No tags found")
		return
	}

	for _, t := range tags {
		fmt.Printf("%-20s  %d items\n", t.Name, t.Count)
	}
}

func sortListItems(items []store.ListItem, sortKey string, desc bool) {
	compare := func(a, b store.ListItem) int {
		switch sortKey {
		case "created":
			if a.CreatedAt < b.CreatedAt {
				return -1
			}
			if a.CreatedAt > b.CreatedAt {
				return 1
			}
		}
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	}

	sort.SliceStable(items, func(i, j int) bool {
		cmp := compare(items[i], items[j])
		if desc {
			cmp = -cmp
		}
		return cmp < 0
	})
}
