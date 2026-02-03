package commands

import (
	"fmt"
	"strings"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var (
	contextTopic      string
	contextRecentDays int
	contextFull       bool
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Smart context for session start",
	Long: `Returns essential context for starting an agent session.

Combines session state, critical knowledge, and recent items in one call.
Designed to give agents everything they need to understand project context.

Includes:
  - Current session state (goal, progress, blockers, next steps)
  - Critical and high severity lessons (always included)
  - Recent items (last 7 days by default)
  - Topic-filtered items (if --topic specified)

Examples:
  dory context                  # Default context
  dory context --topic auth     # Include all auth-related items
  dory context --recent 14      # Recent items from last 14 days
  dory context --full           # Include all items
  dory context --json           # Output as JSON for programmatic use`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		s := store.New("")
		defer s.Close()

		result, err := s.Context(contextTopic, contextRecentDays, contextFull)
		CheckError(err)

		OutputResult(cmd, result, func() {
			printContext(result)
		})
	},
}

func printContext(ctx *store.ContextResult) {
	// Header
	fmt.Printf("╭─────────────────────────────────────────╮\n")
	fmt.Printf("│  CONTEXT: %-29s │\n", ctx.Project)
	fmt.Printf("╰─────────────────────────────────────────╯\n\n")

	// Session State
	if ctx.State != nil && (ctx.State.Goal != "" || ctx.State.Progress != "" || len(ctx.State.Next) > 0) {
		fmt.Println("SESSION STATE")
		fmt.Println(strings.Repeat("─", 50))

		if ctx.State.Goal != "" {
			fmt.Printf("  Goal: %s\n", ctx.State.Goal)
		}
		if ctx.State.Progress != "" {
			fmt.Printf("  Progress: %s\n", ctx.State.Progress)
		}
		if ctx.State.Blocker != "" {
			fmt.Printf("  ⚠ Blocker: %s\n", ctx.State.Blocker)
		}
		if len(ctx.State.Next) > 0 {
			fmt.Println("  Next:")
			for _, n := range ctx.State.Next {
				fmt.Printf("    • %s\n", n)
			}
		}
		if ctx.State.LastUpdated != "" {
			fmt.Printf("  (updated: %s)\n", ctx.State.LastUpdated)
		}
		fmt.Println()
	}

	// Critical lessons
	if len(ctx.Critical) > 0 {
		fmt.Printf("CRITICAL/HIGH LESSONS (%d)\n", len(ctx.Critical))
		fmt.Println(strings.Repeat("─", 50))
		for _, item := range ctx.Critical {
			sev := ""
			if item.Severity != "" {
				sev = fmt.Sprintf("[%s] ", item.Severity)
			}
			fmt.Printf("  %s: %s%s\n", item.ID, sev, truncateOneliner(item.Oneliner, 40))
		}
		fmt.Println()
	}

	// Topic items
	if len(ctx.Topic) > 0 {
		fmt.Printf("TOPIC ITEMS (%d)\n", len(ctx.Topic))
		fmt.Println(strings.Repeat("─", 50))
		for _, item := range ctx.Topic {
			fmt.Printf("  %s [%s]: %s\n", item.ID, item.Type, truncateOneliner(item.Oneliner, 35))
		}
		fmt.Println()
	}

	// Recent items
	if len(ctx.Recent) > 0 {
		fmt.Printf("RECENT ITEMS (%d)\n", len(ctx.Recent))
		fmt.Println(strings.Repeat("─", 50))
		for _, item := range ctx.Recent {
			fmt.Printf("  %s [%s]: %s\n", item.ID, item.Type, truncateOneliner(item.Oneliner, 35))
		}
		fmt.Println()
	}

	// Summary
	total := len(ctx.Critical) + len(ctx.Recent)
	if len(ctx.Topic) > 0 {
		total += len(ctx.Topic)
	}
	fmt.Printf("Total: %d items loaded\n", total)
	fmt.Println("Use 'dory show <id>' for full content, 'dory expand <id>' for related items")
}

func truncateOneliner(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

func init() {
	contextCmd.Flags().StringVar(&contextTopic, "topic", "", "Include all items for this topic")
	contextCmd.Flags().IntVar(&contextRecentDays, "recent", 7, "Include items from last N days")
	contextCmd.Flags().BoolVar(&contextFull, "full", false, "Include all items (for small knowledge bases)")
	RootCmd.AddCommand(contextCmd)
}
