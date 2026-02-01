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
	Long:  `List all lessons, decisions, and patterns with optional filters.`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		topic, _ := cmd.Flags().GetString("topic")
		itemType, _ := cmd.Flags().GetString("type")
		severityStr, _ := cmd.Flags().GetString("severity")
		severity := models.Severity(severityStr)
		sinceStr, _ := cmd.Flags().GetString("since")
		untilStr, _ := cmd.Flags().GetString("until")

		var since, until time.Time
		if sinceStr != "" {
			since, _ = time.Parse("2006-01-02", sinceStr)
		}
		if untilStr != "" {
			until, _ = time.Parse("2006-01-02", untilStr)
		}

		s := store.New("")
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
	listCmd.Flags().String("type", "", "Filter by type: lesson, decision, pattern")
	listCmd.Flags().StringP("severity", "s", "", "Filter by severity: critical, high, normal, low")
	listCmd.Flags().String("since", "", "Show items created on or after date (YYYY-MM-DD)")
	listCmd.Flags().String("until", "", "Show items created on or before date (YYYY-MM-DD)")
	RootCmd.AddCommand(listCmd)
}
