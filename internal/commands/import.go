package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sibellavia/dory/internal/models"
	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var importCmd = &cobra.Command{
	Use:   "import <file.md>",
	Short: "Import a markdown file as knowledge",
	Long: `Import an existing markdown file into dory.

If the file has YAML frontmatter, dory will extract type, topic, domain, and severity from it.
CLI flags override frontmatter values.

Use --split to parse numbered items (e.g., "1) Title" or "1. Title") as separate entries.

Examples:
  dory import docs/api-gotchas.md --type lesson --topic api
  dory import docs/why-redis.md --type decision --topic caching
  dory import notes.md  # uses frontmatter if present
  dory import lessons.md --type lesson --topic infra --split`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		filePath := args[0]
		content, err := os.ReadFile(filePath)
		CheckError(err)

		frontmatter, body := parseFrontmatter(string(content))

		itemType, _ := cmd.Flags().GetString("type")
		topic, _ := cmd.Flags().GetString("topic")
		domain, _ := cmd.Flags().GetString("domain")
		severityStr, _ := cmd.Flags().GetString("severity")
		refs, _ := cmd.Flags().GetStringSlice("refs")
		split, _ := cmd.Flags().GetBool("split")

		if itemType == "" {
			if v, ok := frontmatter["type"].(string); ok {
				itemType = v
			}
		}
		if topic == "" {
			if v, ok := frontmatter["topic"].(string); ok {
				topic = v
			}
		}
		if domain == "" {
			if v, ok := frontmatter["domain"].(string); ok {
				domain = v
			}
		}
		if severityStr == "" {
			if v, ok := frontmatter["severity"].(string); ok {
				severityStr = v
			}
		}

		if itemType == "" {
			fmt.Fprintln(os.Stderr, "Error: --type is required (lesson, decision, pattern) or specify in frontmatter")
			os.Exit(1)
		}

		severity := models.Severity(severityStr)
		if severity == "" {
			severity = models.SeverityNormal
		}

		s := store.NewSingle("")
		defer s.Close()

		if split {
			items := splitNumberedItems(body)
			if len(items) == 0 {
				fmt.Fprintln(os.Stderr, "Error: no numbered items found (expected patterns like '1) Title' or '1. Title')")
				os.Exit(1)
			}
			for _, item := range items {
				id, err := importItem(s, itemType, item.title, item.body, topic, domain, severity, refs)
				CheckError(err)
				fmt.Printf("Imported %s: %s\n", id, item.title)
			}
			fmt.Printf("\nImported %d items\n", len(items))
		} else {
			oneliner := extractOneliner(body, filePath)
			id, err := importItem(s, itemType, oneliner, body, topic, domain, severity, refs)
			CheckError(err)
			fmt.Printf("Imported %s: %s\n", id, oneliner)
		}
	},
}

func init() {
	importCmd.Flags().String("type", "", "Item type: lesson, decision, pattern")
	importCmd.Flags().StringP("topic", "t", "", "Topic (for lessons and decisions)")
	importCmd.Flags().StringP("domain", "d", "", "Domain (for patterns)")
	importCmd.Flags().StringP("severity", "s", "", "Severity: critical, high, normal, low")
	importCmd.Flags().StringSliceP("refs", "R", nil, "References to other items")
	importCmd.Flags().Bool("split", false, "Split numbered items into separate entries")
	RootCmd.AddCommand(importCmd)
}

func importItem(s *store.SingleStore, itemType, oneliner, body, topic, domain string, severity models.Severity, refs []string) (string, error) {
	switch itemType {
	case "lesson":
		if topic == "" {
			fmt.Fprintln(os.Stderr, "Error: --topic is required for lessons")
			os.Exit(1)
		}
		return s.Learn(oneliner, topic, severity, "", body, refs)
	case "decision":
		if topic == "" {
			fmt.Fprintln(os.Stderr, "Error: --topic is required for decisions")
			os.Exit(1)
		}
		return s.Decide(oneliner, topic, "", "", body, refs)
	case "pattern":
		if domain == "" {
			domain = topic
		}
		if domain == "" {
			fmt.Fprintln(os.Stderr, "Error: --domain is required for patterns")
			os.Exit(1)
		}
		return s.Pattern(oneliner, domain, "", body, refs)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown type %q (use lesson, decision, or pattern)\n", itemType)
		os.Exit(1)
		return "", nil
	}
}

type numberedItem struct {
	title string
	body  string
}

var numberedItemPattern = regexp.MustCompile(`^(\d+)[)\.]\s+(.+)$`)

func splitNumberedItems(content string) []numberedItem {
	var items []numberedItem
	var currentTitle string
	var currentBody strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if match := numberedItemPattern.FindStringSubmatch(line); match != nil {
			if currentTitle != "" {
				items = append(items, numberedItem{
					title: currentTitle,
					body:  strings.TrimSpace(currentBody.String()),
				})
			}
			currentTitle = match[2]
			currentBody.Reset()
		} else if currentTitle != "" {
			currentBody.WriteString(line)
			currentBody.WriteString("\n")
		}
	}

	if currentTitle != "" {
		items = append(items, numberedItem{
			title: currentTitle,
			body:  strings.TrimSpace(currentBody.String()),
		})
	}

	return items
}

func parseFrontmatter(content string) (map[string]interface{}, string) {
	frontmatter := make(map[string]interface{})

	if !strings.HasPrefix(content, "---\n") {
		return frontmatter, content
	}

	rest := content[4:]
	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		return frontmatter, content
	}

	yamlContent := rest[:endIdx]
	body := strings.TrimPrefix(rest[endIdx+4:], "\n")

	yaml.Unmarshal([]byte(yamlContent), &frontmatter)
	return frontmatter, body
}

func extractOneliner(body, filePath string) string {
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	base := filepath.Base(filePath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
