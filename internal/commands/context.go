package commands

import (
	"fmt"
	"strings"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Get or update session context",
	Long: `Get session context, or update session state and return context.

READ MODE (no state flags):
  Returns essential context for starting an agent session:
  - Current session state (goal, progress, blockers, next steps)
  - Critical and high severity lessons
  - Recent items (last 7 days by default)

WRITE MODE (with state flags):
  Updates session state, then returns full context.
  Use at end of session to save progress for next agent.

Examples:
  dory context                              # Get context (read)
  dory context --tag auth                   # Include auth-related items
  dory context --goal "Add auth" --progress "50%" --next "Add logout"  # Update state
  dory context --goal "Add auth" --next "Step 1" --next "Step 2"       # Multiple next steps`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		// State flags (write mode)
		goal, _ := cmd.Flags().GetString("goal")
		progress, _ := cmd.Flags().GetString("progress")
		blocker, _ := cmd.Flags().GetString("blocker")
		next, _ := cmd.Flags().GetStringSlice("next")
		workingFiles, _ := cmd.Flags().GetStringSlice("working-file")
		openQuestions, _ := cmd.Flags().GetStringSlice("question")

		// Context flags (read mode)
		tag, _ := cmd.Flags().GetString("tag")
		recentDays, _ := cmd.Flags().GetInt("recent")
		full, _ := cmd.Flags().GetBool("full")

		s := store.New(doryRoot)
		defer s.Close()

		// Check if any state flags provided (write mode)
		hasStateFlags := goal != "" || progress != "" || blocker != "" ||
			len(next) > 0 || len(workingFiles) > 0 || len(openQuestions) > 0

		if hasStateFlags {
			// Write mode: update state first
			_, err := s.UpdateStatus(goal, progress, blocker, next, workingFiles, openQuestions)
			CheckError(err)
		}

		// Always return full context
		result, err := s.Context(tag, recentDays, full)
		CheckError(err)

		OutputResult(cmd, result, func() {
			printContext(result, hasStateFlags)
		})
	},
}

func printContext(ctx *store.ContextResult, updated bool) {
	// Header
	fmt.Printf("╭─────────────────────────────────────────╮\n")
	fmt.Printf("│  CONTEXT: %-29s │\n", ctx.Project)
	fmt.Printf("╰─────────────────────────────────────────╯\n\n")

	if updated {
		fmt.Println("Session state updated")
		fmt.Println()
	}

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
			fmt.Printf("  Blocker: %s\n", ctx.State.Blocker)
		}
		if len(ctx.State.Next) > 0 {
			fmt.Println("  Next:")
			for _, n := range ctx.State.Next {
				fmt.Printf("    - %s\n", n)
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
		fmt.Printf("TAG ITEMS (%d)\n", len(ctx.Topic))
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
	// Read mode flags
	contextCmd.Flags().StringP("tag", "T", "", "Include all items for this tag")
	contextCmd.Flags().Int("recent", 7, "Include items from last N days")
	contextCmd.Flags().Bool("full", false, "Include all items")

	// Write mode flags (state)
	contextCmd.Flags().StringP("goal", "g", "", "Set current goal")
	contextCmd.Flags().StringP("progress", "p", "", "Set current progress")
	contextCmd.Flags().StringP("blocker", "b", "", "Set current blocker")
	contextCmd.Flags().StringSliceP("next", "n", nil, "Set next steps (repeatable)")
	contextCmd.Flags().StringSlice("working-file", nil, "Set working files (repeatable)")
	contextCmd.Flags().StringSliceP("question", "q", nil, "Set open questions (repeatable)")

	RootCmd.AddCommand(contextCmd)
}
