package commands

import (
	"fmt"
	"path/filepath"

	"github.com/simonebellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize dory in current project",
	Long:  `Creates the .dory directory structure with index.yaml and state.yaml.`,
	Run: func(cmd *cobra.Command, args []string) {
		project, _ := cmd.Flags().GetString("project")
		description, _ := cmd.Flags().GetString("description")

		// Default project name to current directory name
		if project == "" {
			cwd, err := filepath.Abs(".")
			CheckError(err)
			project = filepath.Base(cwd)
		}

		s := store.New("")
		err := s.Init(project, description)
		CheckError(err)

		result := map[string]string{
			"status":  "initialized",
			"project": project,
			"path":    ".dory",
		}

		OutputResult(cmd, result, func() {
			fmt.Printf("Initialized dory for '%s' in .dory/\n", project)
		})
	},
}

func init() {
	initCmd.Flags().StringP("project", "p", "", "Project name (default: current directory name)")
	initCmd.Flags().StringP("description", "d", "", "Project description")
	RootCmd.AddCommand(initCmd)
}
