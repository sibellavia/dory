package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sibellavia/dory/content"
	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

const doryInstructions = `
## Dory - Project Memory

This project uses dory for persistent knowledge. At session start:
` + "```bash" + `
cat .dory/index.yaml
` + "```" + `

Record lessons (` + "`dory learn`" + `), decisions (` + "`dory decide`" + `), and patterns (` + "`dory pattern`" + `).
Update status before ending (` + "`dory status`" + `).
`


var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize dory in current project",
	Long:  `Creates the .dory directory structure with index.yaml.`,
	Run: func(cmd *cobra.Command, args []string) {
		project, _ := cmd.Flags().GetString("project")
		description, _ := cmd.Flags().GetString("description")

		// Default project name to current directory name
		if project == "" {
			cwd, err := filepath.Abs(".")
			CheckError(err)
			project = filepath.Base(cwd)
		}

		s := store.NewSingle("")
		err := s.Init(project, description)
		CheckError(err)
		s.Close()

		// Create DORY.md if it doesn't exist
		doryMdCreated := createDoryMd()

		// Auto-append to CLAUDE.md and/or AGENTS.md if they exist
		agentFiles := []string{"CLAUDE.md", "AGENTS.md"}
		var appendedTo []string
		for _, file := range agentFiles {
			if appendDoryInstructions(file) {
				appendedTo = append(appendedTo, file)
			}
		}

		result := map[string]interface{}{
			"status":  "initialized",
			"project": project,
			"path":    ".dory",
		}
		if doryMdCreated {
			result["created"] = "DORY.md"
		}
		if len(appendedTo) > 0 {
			result["appended_to"] = appendedTo
		}

		OutputResult(cmd, result, func() {
			fmt.Printf("Initialized dory for '%s' in .dory/\n", project)
			if doryMdCreated {
				fmt.Println("Created DORY.md")
			}
			for _, file := range appendedTo {
				fmt.Printf("Added dory instructions to %s\n", file)
			}
		})
	},
}

// createDoryMd creates DORY.md if it doesn't exist. Returns true if created.
func createDoryMd() bool {
	filename := "DORY.md"

	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		return false // File exists
	}

	// Create the file with embedded content
	err := os.WriteFile(filename, []byte(content.DoryMd), 0644)
	return err == nil
}

// appendDoryInstructions appends dory instructions to a file if it exists
// and doesn't already contain dory instructions. Returns true if appended.
func appendDoryInstructions(filename string) bool {
	// Check if file exists
	content, err := os.ReadFile(filename)
	if err != nil {
		return false // File doesn't exist
	}

	// Check if already contains dory instructions
	if strings.Contains(string(content), "## Dory - Project Memory") {
		return false // Already has instructions
	}

	// Append instructions
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.WriteString(doryInstructions)
	return err == nil
}

func init() {
	initCmd.Flags().StringP("project", "p", "", "Project name (default: current directory name)")
	initCmd.Flags().StringP("description", "d", "", "Project description")
	RootCmd.AddCommand(initCmd)
}
