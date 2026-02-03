package commands

import (
	"fmt"
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

		CheckError(validateItemType(itemType))
		CheckError(validateSeverityFlag(severity))

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

		s := store.New("")
		defer s.Close()
		items, err := s.List(topic, itemType, severity, since, until)
		CheckError(err)

		OutputResult(cmd, items, func() {
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
		})
	},
}

func init() {
	listCmd.Flags().StringP("topic", "t", "", "Filter by topic")
	listCmd.Flags().String("type", "", "Filter by type (e.g. lesson, decision, pattern, or plugin custom type)")
	listCmd.Flags().StringP("severity", "s", "", "Filter by severity: critical, high, normal, low")
	listCmd.Flags().String("since", "", "Show items created on or after date (YYYY-MM-DD)")
	listCmd.Flags().String("until", "", "Show items created on or before date (YYYY-MM-DD)")
	RootCmd.AddCommand(listCmd)
}
