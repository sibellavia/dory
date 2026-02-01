package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove an item",
	Long:  `Delete an item from the knowledge store. Requires confirmation unless --force is used.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		id := args[0]
		force, _ := cmd.Flags().GetBool("force")

		s := store.New("")

		// Show item before removal
		content, err := s.Show(id)
		CheckError(err)

		if !force {
			fmt.Printf("About to remove %s:\n\n", id)
			// Show first few lines
			lines := strings.Split(content, "\n")
			maxLines := 10
			if len(lines) < maxLines {
				maxLines = len(lines)
			}
			for _, line := range lines[:maxLines] {
				fmt.Println(line)
			}
			if len(lines) > 10 {
				fmt.Println("...")
			}

			fmt.Print("\nConfirm removal? [y/N] ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Aborted")
				return
			}
		}

		err = s.Remove(id)
		CheckError(err)

		result := map[string]string{
			"id":     id,
			"status": "removed",
		}

		OutputResult(cmd, result, func() {
			fmt.Printf("Removed %s\n", id)
		})
	},
}

func init() {
	removeCmd.Flags().BoolP("force", "f", false, "Remove without confirmation")
	RootCmd.AddCommand(removeCmd)
}
