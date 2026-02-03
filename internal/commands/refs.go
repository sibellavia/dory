package commands

import (
	"fmt"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var refsCmd = &cobra.Command{
	Use:   "refs <id>",
	Short: "Show relationships for an item",
	Long: `Shows what an item references and what references it.

Useful for understanding how knowledge items are connected:
- refs_to: items this one references (was caused by, builds on)
- referenced_by: items that reference this one (led to, influenced)

Examples:
  dory refs D001           # Show refs for decision D001
  dory refs L003 --json    # Output as JSON for programmatic use`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		id := args[0]
		s := store.New("")
		defer s.Close()

		refInfo, err := s.Refs(id)
		CheckError(err)

		OutputResult(cmd, refInfo, func() {
			printRefInfo(refInfo)
		})
	},
}

func printRefInfo(info *store.RefInfo) {
	fmt.Printf("%s: %s\n", info.ID, info.Oneliner)

	if len(info.RefsTo) == 0 && len(info.ReferencedBy) == 0 {
		fmt.Println("  (no references)")
		return
	}

	for _, ref := range info.RefsTo {
		fmt.Printf("  ← refs: %s (%s)\n", ref.ID, ref.Oneliner)
	}

	for _, ref := range info.ReferencedBy {
		fmt.Printf("  → referenced by: %s (%s)\n", ref.ID, ref.Oneliner)
	}
}

func init() {
	RootCmd.AddCommand(refsCmd)
}
