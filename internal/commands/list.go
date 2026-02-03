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
	Short: "List all items",
	Long:  `List all knowledge items (core and custom types) with optional filters.`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		topic, _ := cmd.Flags().GetString("topic")
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

		s := store.New(doryRoot)
		defer s.Close()
		items, err := s.List(topic, itemType, severity, since, until)
		CheckError(err)
		sortListItems(items, sortKey, desc)

		OutputResult(cmd, items, func() { renderListHuman(items) })
	},
}

func init() {
	listCmd.Flags().StringP("topic", "t", "", "Filter by topic")
	listCmd.Flags().String("type", "", "Filter by type (e.g. lesson, decision, pattern, or plugin custom type)")
	listCmd.Flags().StringP("severity", "s", "", "Filter by severity: critical, high, normal, low")
	listCmd.Flags().String("since", "", "Show items created on or after date (YYYY-MM-DD)")
	listCmd.Flags().String("until", "", "Show items created on or before date (YYYY-MM-DD)")
	listCmd.Flags().String("sort", "id", "Sort by: id, created")
	listCmd.Flags().Bool("desc", false, "Sort in descending order")
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
