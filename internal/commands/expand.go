package commands

import (
	"fmt"
	"strings"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var expandDepth int

var expandCmd = &cobra.Command{
	Use:   "expand <id>",
	Short: "Get an item and all connected items",
	Long: `Returns full content of an item plus all items connected to it.

Traverses the reference graph to collect related knowledge in one call.
Use --depth to control how many hops to traverse (default: 1).

Examples:
  dory expand D-01JX...             # Item + directly connected items
  dory expand D-01JX... --depth 2   # Include items 2 hops away
  dory expand D-01JX... --json      # Output as JSON`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		id := args[0]
		s := store.New(doryRoot)
		defer s.Close()

		result, err := s.Expand(id, expandDepth)
		CheckError(err)

		OutputResult(cmd, result, func() {
			printExpandResult(result)
		})
	},
}

func printExpandResult(result *store.ExpandResult) {
	// Print root item
	printExpandedItem(&result.Root, true)

	// Print connected items
	if len(result.Connected) > 0 {
		fmt.Printf("\n--- %d connected item(s) ---\n", len(result.Connected))
		for i := range result.Connected {
			fmt.Println()
			printExpandedItem(&result.Connected[i], false)
		}
	}
}

func printExpandedItem(item *store.ExpandedItem, isRoot bool) {
	if isRoot {
		fmt.Printf("=== %s [%s] ===\n", item.ID, item.Type)
	} else {
		fmt.Printf("--- %s [%s] ---\n", item.ID, item.Type)
	}
	fmt.Printf("%s\n", item.Oneliner)

	if item.Topic != "" {
		fmt.Printf("topic: %s\n", item.Topic)
	}
	if item.Domain != "" {
		fmt.Printf("domain: %s\n", item.Domain)
	}
	if len(item.Refs) > 0 {
		fmt.Printf("refs: %s\n", strings.Join(item.Refs, ", "))
	}

	fmt.Println()
	fmt.Println(item.Body)
}

func init() {
	expandCmd.Flags().IntVar(&expandDepth, "depth", 1, "Number of hops to traverse (default: 1)")
	RootCmd.AddCommand(expandCmd)
}
