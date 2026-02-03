package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var graphDepth int

var graphCmd = &cobra.Command{
	Use:   "graph [id]",
	Short: "Visualize knowledge graph",
	Long: `Visualize the knowledge graph in the terminal.

Without an ID, shows an overview of all items and connections.
With an ID, shows a visual graph centered on that item.

Examples:
  dory graph              # Overview of all items
  dory graph D001         # Graph centered on D001
  dory graph D001 --depth 3  # Include items up to 3 hops away`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		s := store.New("")
		defer s.Close()

		if len(args) == 0 {
			output, err := generateFullTerminalGraph(s)
			CheckError(err)
			fmt.Print(output)
		} else {
			output, err := generateTerminalGraph(s, args[0])
			CheckError(err)
			fmt.Print(output)
		}
	},
}

func generateTerminalGraph(s *store.Store, id string) (string, error) {
	refInfo, err := s.Refs(id)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("\n")

	// Layer 0: Items that root references (upstream)
	if len(refInfo.RefsTo) > 0 {
		sb.WriteString(renderNodeRow(refInfo.RefsTo, ""))
		sb.WriteString(renderUpwardConnectors(len(refInfo.RefsTo)))
		sb.WriteString("\n")
	}

	// Layer 1: Root node (centered, highlighted)
	sb.WriteString(renderRootNode(id, refInfo.Oneliner))

	// Layer 2: Items that reference root (downstream)
	if len(refInfo.ReferencedBy) > 0 {
		sb.WriteString(renderDownwardConnectors(len(refInfo.ReferencedBy)))
		sb.WriteString(renderNodeRow(refInfo.ReferencedBy, ""))
	}

	sb.WriteString("\n")
	sb.WriteString("  Legend: ═══ root node   ─── connected nodes   │ reference\n")

	return sb.String(), nil
}

func renderRootNode(id, oneliner string) string {
	var sb strings.Builder

	if len(oneliner) > 50 {
		oneliner = oneliner[:47] + "..."
	}

	idLine := fmt.Sprintf("  %s  ", id)
	width := len(idLine)
	if len(oneliner)+4 > width {
		width = len(oneliner) + 4
	}

	padding := (60 - width) / 2
	pad := strings.Repeat(" ", padding)

	sb.WriteString(pad + "╔" + strings.Repeat("═", width) + "╗\n")
	sb.WriteString(pad + "║" + centerText(id, width) + "║\n")
	sb.WriteString(pad + "║" + centerText(oneliner, width) + "║\n")
	sb.WriteString(pad + "╚" + strings.Repeat("═", width) + "╝\n")

	return sb.String()
}

func renderNodeRow(items []store.RefItem, highlight string) string {
	if len(items) == 0 {
		return ""
	}

	var sb strings.Builder
	const nodeWidth = 14

	totalWidth := len(items)*nodeWidth + (len(items)-1)*2
	padding := (60 - totalWidth) / 2
	if padding < 0 {
		padding = 0
	}
	pad := strings.Repeat(" ", padding)

	// Top border
	sb.WriteString(pad)
	for i, item := range items {
		if i > 0 {
			sb.WriteString("  ")
		}
		if item.ID == highlight {
			sb.WriteString("╔" + strings.Repeat("═", nodeWidth-2) + "╗")
		} else {
			sb.WriteString("┌" + strings.Repeat("─", nodeWidth-2) + "┐")
		}
	}
	sb.WriteString("\n")

	// ID line
	sb.WriteString(pad)
	for i, item := range items {
		if i > 0 {
			sb.WriteString("  ")
		}
		border := "│"
		if item.ID == highlight {
			border = "║"
		}
		sb.WriteString(border + centerText(item.ID, nodeWidth-2) + border)
	}
	sb.WriteString("\n")

	// Bottom border
	sb.WriteString(pad)
	for i, item := range items {
		if i > 0 {
			sb.WriteString("  ")
		}
		if item.ID == highlight {
			sb.WriteString("╚" + strings.Repeat("═", nodeWidth-2) + "╝")
		} else {
			sb.WriteString("└" + strings.Repeat("─", nodeWidth-2) + "┘")
		}
	}
	sb.WriteString("\n")

	return sb.String()
}

func renderUpwardConnectors(count int) string {
	if count == 0 {
		return ""
	}

	var sb strings.Builder
	const nodeWidth = 14

	totalWidth := count*nodeWidth + (count-1)*2
	padding := (60 - totalWidth) / 2
	if padding < 0 {
		padding = 0
	}

	// Vertical bars from each node
	sb.WriteString(strings.Repeat(" ", padding))
	for i := 0; i < count; i++ {
		if i > 0 {
			sb.WriteString("  ")
		}
		left := (nodeWidth - 2) / 2
		right := nodeWidth - 2 - left - 1
		sb.WriteString(strings.Repeat(" ", left) + "│" + strings.Repeat(" ", right+1))
	}
	sb.WriteString("\n")

	// Merge line
	sb.WriteString(strings.Repeat(" ", padding))
	for i := 0; i < count; i++ {
		left := (nodeWidth - 2) / 2
		if i == 0 {
			sb.WriteString(strings.Repeat(" ", left) + "└")
			sb.WriteString(strings.Repeat("─", nodeWidth-left-2+1))
		} else if i == count-1 {
			sb.WriteString("─" + strings.Repeat("─", left) + "┘")
			sb.WriteString(strings.Repeat(" ", nodeWidth-left-3))
		} else {
			sb.WriteString("─" + strings.Repeat("─", left) + "┴")
			sb.WriteString(strings.Repeat("─", nodeWidth-left-3+1))
		}
	}
	sb.WriteString("\n")

	// Single line down to root
	center := padding + (totalWidth / 2)
	sb.WriteString(strings.Repeat(" ", center) + "│\n")
	sb.WriteString(strings.Repeat(" ", center) + "▼\n")

	return sb.String()
}

func renderDownwardConnectors(count int) string {
	if count == 0 {
		return ""
	}

	var sb strings.Builder
	const nodeWidth = 14

	totalWidth := count*nodeWidth + (count-1)*2
	padding := (60 - totalWidth) / 2
	if padding < 0 {
		padding = 0
	}

	// Single line down from root
	center := padding + (totalWidth / 2)
	sb.WriteString(strings.Repeat(" ", center) + "│\n")

	// Split line
	sb.WriteString(strings.Repeat(" ", padding))
	for i := 0; i < count; i++ {
		left := (nodeWidth - 2) / 2
		if i == 0 {
			sb.WriteString(strings.Repeat(" ", left) + "┌")
			sb.WriteString(strings.Repeat("─", nodeWidth-left-2+1))
		} else if i == count-1 {
			sb.WriteString("─" + strings.Repeat("─", left) + "┐")
			sb.WriteString(strings.Repeat(" ", nodeWidth-left-3))
		} else {
			sb.WriteString("─" + strings.Repeat("─", left) + "┬")
			sb.WriteString(strings.Repeat("─", nodeWidth-left-3+1))
		}
	}
	sb.WriteString("\n")

	// Arrows down to each node
	sb.WriteString(strings.Repeat(" ", padding))
	for i := 0; i < count; i++ {
		if i > 0 {
			sb.WriteString("  ")
		}
		left := (nodeWidth - 2) / 2
		right := nodeWidth - 2 - left - 1
		sb.WriteString(strings.Repeat(" ", left) + "▼" + strings.Repeat(" ", right+1))
	}
	sb.WriteString("\n")

	return sb.String()
}

func centerText(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	padding := (width - len(s)) / 2
	right := width - len(s) - padding
	return strings.Repeat(" ", padding) + s + strings.Repeat(" ", right)
}

func generateFullTerminalGraph(s *store.Store) (string, error) {
	items, err := s.List("", "", "", time.Time{}, time.Time{})
	if err != nil {
		return "", err
	}

	if len(items) == 0 {
		return "(empty knowledge base)\n", nil
	}

	var sb strings.Builder

	// Build edge list
	type edge struct{ from, to string }
	var edges []edge
	hasRefs := make(map[string]bool)
	isReferenced := make(map[string]bool)

	for _, item := range items {
		refInfo, err := s.Refs(item.ID)
		if err != nil {
			continue
		}
		for _, ref := range refInfo.RefsTo {
			edges = append(edges, edge{item.ID, ref.ID})
			hasRefs[item.ID] = true
			isReferenced[ref.ID] = true
		}
	}

	sb.WriteString("╭─────────────────────────────────────────╮\n")
	sb.WriteString("│         KNOWLEDGE GRAPH                 │\n")
	sb.WriteString("├─────────────────────────────────────────┤\n")
	sb.WriteString(fmt.Sprintf("│  Items: %-6d  Connections: %-6d    │\n", len(items), len(edges)))
	sb.WriteString("╰─────────────────────────────────────────╯\n\n")

	// Group by type
	byType := make(map[string][]store.ListItem)
	for _, item := range items {
		byType[item.Type] = append(byType[item.Type], item)
	}

	typeOrder := []string{"lesson", "decision", "pattern"}
	typeSymbol := map[string]string{
		"lesson":   "◆",
		"decision": "◼",
		"pattern":  "●",
	}

	// Add other types
	for t := range byType {
		found := false
		for _, to := range typeOrder {
			if t == to {
				found = true
				break
			}
		}
		if !found {
			typeOrder = append(typeOrder, t)
			typeSymbol[t] = "○"
		}
	}

	for _, itemType := range typeOrder {
		typeItems, ok := byType[itemType]
		if !ok {
			continue
		}

		symbol := typeSymbol[itemType]
		sb.WriteString(fmt.Sprintf("%s %s (%d)\n", symbol, strings.ToUpper(itemType), len(typeItems)))
		sb.WriteString(strings.Repeat("─", 50) + "\n")

		for _, item := range typeItems {
			oneliner := item.Oneliner
			if len(oneliner) > 40 {
				oneliner = oneliner[:37] + "..."
			}

			indicator := "  "
			if hasRefs[item.ID] && isReferenced[item.ID] {
				indicator = "⇄ "
			} else if hasRefs[item.ID] {
				indicator = "→ "
			} else if isReferenced[item.ID] {
				indicator = "← "
			}

			sb.WriteString(fmt.Sprintf("  %s%-5s %s\n", indicator, item.ID, oneliner))
		}
		sb.WriteString("\n")
	}

	if len(edges) > 0 {
		sb.WriteString("CONNECTIONS\n")
		sb.WriteString(strings.Repeat("─", 50) + "\n")
		for _, e := range edges {
			sb.WriteString(fmt.Sprintf("  %s → %s\n", e.from, e.to))
		}
	}

	sb.WriteString("\nLegend: → has refs  ← is referenced  ⇄ both\n")

	return sb.String(), nil
}

func init() {
	graphCmd.Flags().IntVar(&graphDepth, "depth", 2, "Depth for graph traversal (default: 2)")
	RootCmd.AddCommand(graphCmd)
}
