package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/sibellavia/dory/internal/fileio"
	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit an item in your editor",
	Long:  `Opens the .eng file for the specified item in your $EDITOR.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		id := args[0]
		s := store.New("")

		filePath := s.GetFilePath(id)
		if !fileio.FileExists(filePath) {
			fmt.Fprintf(os.Stderr, "Error: item %s not found\n", id)
			os.Exit(1)
		}

		editor := fileio.GetEditor()
		editorCmd := exec.Command(editor, filePath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		err := editorCmd.Run()
		CheckError(err)

		// Rebuild index after edit to capture any changes
		err = s.Rebuild()
		CheckError(err)

		result := map[string]string{
			"id":     id,
			"status": "edited",
		}

		OutputResult(cmd, result, func() {
			fmt.Printf("Edited %s\n", id)
		})
	},
}

func init() {
	RootCmd.AddCommand(editCmd)
}
