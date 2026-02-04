package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

// Graph types for JSON/YAML output
type GraphNode struct {
	ID       string `json:"id" yaml:"id"`
	Type     string `json:"type" yaml:"type"`
	Oneliner string `json:"oneliner" yaml:"oneliner"`
}

type GraphEdge struct {
	From string `json:"from" yaml:"from"`
	To   string `json:"to" yaml:"to"`
}

type GraphResult struct {
	Center string      `json:"center,omitempty" yaml:"center,omitempty"`
	Depth  int         `json:"depth" yaml:"depth"`
	Nodes  []GraphNode `json:"nodes" yaml:"nodes"`
	Edges  []GraphEdge `json:"edges" yaml:"edges"`
}

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show full content for an item",
	Long: `Returns the complete content for a specific item.

Use --refs to include relationships (what it references, what references it).
Use --expand to include full content of connected items.
Use --graph to visualize the item's connections.

Examples:
  dory show D-01JX...                     # Content only
  dory show D-01JX... --refs              # Content + relationships
  dory show D-01JX... --expand            # Content + connected items
  dory show D-01JX... --expand --depth 2  # Include items 2 hops away
  dory show D-01JX... --graph             # Visual graph centered on item`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		id := args[0]
		showRefs, _ := cmd.Flags().GetBool("refs")
		showExpand, _ := cmd.Flags().GetBool("expand")
		showGraph, _ := cmd.Flags().GetBool("graph")
		depth, _ := cmd.Flags().GetInt("depth")

		s := store.New(doryRoot)
		defer s.Close()

		// --graph mode: visual graph centered on item
		if showGraph {
			format := GetOutputFormat(cmd)
			if format == "json" || format == "yaml" {
				result, err := buildGraphData(s, id, depth)
				CheckError(err)
				OutputResult(cmd, result, func() {})
				return
			}
			output, err := generateTerminalGraph(s, id, depth)
			CheckError(err)
			fmt.Print(output)
			return
		}

		// --expand mode: show item + connected items
		if showExpand {
			result, err := s.Expand(id, depth)
			CheckError(err)

			OutputResult(cmd, result, func() {
				printExpandResult(result)
			})
			return
		}

		// --refs mode: show content + relationships
		if showRefs {
			content, err := s.Show(id)
			CheckError(err)
			refInfo, err := s.Refs(id)
			CheckError(err)

			result := map[string]interface{}{
				"id":            id,
				"content":       content,
				"refs_to":       refInfo.RefsTo,
				"referenced_by": refInfo.ReferencedBy,
			}

			OutputResult(cmd, result, func() {
				fmt.Print(content)
				fmt.Println()
				printRefInfo(refInfo)
			})
			return
		}

		// Default: content only
		content, err := s.Show(id)
		CheckError(err)

		format := GetOutputFormat(cmd)
		if format == "human" {
			fmt.Print(content)
			return
		}

		result := map[string]string{
			"id":      id,
			"content": content,
		}
		if format == "json" || format == "yaml" {
			OutputResult(cmd, result, func() {})
			return
		}

		OutputResult(cmd, result, func() {
			fmt.Print(content)
		})
	},
}

func printRefInfo(info *store.RefInfo) {
	fmt.Printf("Refs for %s:\n", info.ID)

	if len(info.RefsTo) == 0 && len(info.ReferencedBy) == 0 {
		fmt.Println("  (no references)")
		return
	}

	for _, ref := range info.RefsTo {
		fmt.Printf("  <- refs: %s (%s)\n", ref.ID, ref.Oneliner)
	}

	for _, ref := range info.ReferencedBy {
		fmt.Printf("  -> referenced by: %s (%s)\n", ref.ID, ref.Oneliner)
	}
}

func printExpandResult(result *store.ExpandResult) {
	printExpandedItem(&result.Root, true)

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

// Graph rendering functions

func generateTerminalGraph(s *store.Store, id string, depth int) (string, error) {
	if depth < 1 {
		depth = 1
	}
	if depth > 1 {
		result, err := buildGraphData(s, id, depth)
		if err != nil {
			return "", err
		}
		return renderDepthGraph(result), nil
	}

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

func renderDepthGraph(result *GraphResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Graph center: %s (depth=%d)\n\n", result.Center, result.Depth))
	sb.WriteString(fmt.Sprintf("Nodes (%d)\n", len(result.Nodes)))
	sb.WriteString(strings.Repeat("─", 50) + "\n")
	for _, node := range result.Nodes {
		oneliner := node.Oneliner
		if len(oneliner) > 60 {
			oneliner = oneliner[:57] + "..."
		}
		sb.WriteString(fmt.Sprintf("  %s [%s] %s\n", node.ID, node.Type, oneliner))
	}

	sb.WriteString(fmt.Sprintf("\nEdges (%d)\n", len(result.Edges)))
	sb.WriteString(strings.Repeat("─", 50) + "\n")
	for _, edge := range result.Edges {
		sb.WriteString(fmt.Sprintf("  %s -> %s\n", edge.From, edge.To))
	}
	sb.WriteString("\n")
	return sb.String()
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

func buildGraphData(s *store.Store, center string, depth int) (*GraphResult, error) {
	if depth < 1 {
		depth = 1
	}

	result := &GraphResult{
		Center: center,
		Depth:  depth,
		Nodes:  make([]GraphNode, 0),
		Edges:  make([]GraphEdge, 0),
	}

	nodeSet := make(map[string]GraphNode)
	if center == "" {
		items, err := s.List("", "", "", time.Time{}, time.Time{})
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			nodeSet[item.ID] = GraphNode{
				ID:       item.ID,
				Type:     item.Type,
				Oneliner: item.Oneliner,
			}
		}
	} else {
		expanded, err := s.Expand(center, depth)
		if err != nil {
			return nil, err
		}
		nodeSet[expanded.Root.ID] = GraphNode{
			ID:       expanded.Root.ID,
			Type:     expanded.Root.Type,
			Oneliner: expanded.Root.Oneliner,
		}
		for _, item := range expanded.Connected {
			nodeSet[item.ID] = GraphNode{
				ID:       item.ID,
				Type:     item.Type,
				Oneliner: item.Oneliner,
			}
		}
	}

	nodeIDs := make([]string, 0, len(nodeSet))
	for id := range nodeSet {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Strings(nodeIDs)
	for _, id := range nodeIDs {
		result.Nodes = append(result.Nodes, nodeSet[id])
	}

	edgeSet := make(map[string]GraphEdge)
	for _, id := range nodeIDs {
		refInfo, err := s.Refs(id)
		if err != nil {
			continue
		}
		for _, ref := range refInfo.RefsTo {
			if _, ok := nodeSet[ref.ID]; !ok {
				continue
			}
			key := id + "->" + ref.ID
			edgeSet[key] = GraphEdge{From: id, To: ref.ID}
		}
	}

	edgeKeys := make([]string, 0, len(edgeSet))
	for key := range edgeSet {
		edgeKeys = append(edgeKeys, key)
	}
	sort.Strings(edgeKeys)
	for _, key := range edgeKeys {
		result.Edges = append(result.Edges, edgeSet[key])
	}

	return result, nil
}

func init() {
	showCmd.Flags().Bool("refs", false, "Include relationships (refs_to, referenced_by)")
	showCmd.Flags().Bool("expand", false, "Include full content of connected items")
	showCmd.Flags().Bool("graph", false, "Visualize connections as a graph")
	showCmd.Flags().Int("depth", 1, "Depth for --expand/--graph traversal (default: 1)")
	RootCmd.AddCommand(showCmd)
}
