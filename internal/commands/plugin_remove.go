package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sibellavia/dory/internal/plugin"
	"github.com/spf13/cobra"
)

var pluginRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an installed plugin from this project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		name := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Remove plugin %s from this project? [y/N] ", name)
			reader := bufio.NewReader(os.Stdin)
			answer, err := reader.ReadString('\n')
			CheckError(err)
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted")
				return
			}
		}

		path, err := plugin.Remove(doryRoot, name)
		CheckError(err)

		OutputResult(cmd, map[string]interface{}{
			"status": "removed",
			"name":   name,
			"path":   path,
		}, func() {
			fmt.Printf("Removed plugin %s\n", name)
		})
	},
}

func init() {
	pluginRemoveCmd.Flags().BoolP("force", "f", false, "Remove without confirmation")
	pluginCmd.AddCommand(pluginRemoveCmd)
}
