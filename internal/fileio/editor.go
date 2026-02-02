package fileio

import "os"

// GetEditor returns the user's preferred editor from environment variables.
func GetEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	return "vi"
}
