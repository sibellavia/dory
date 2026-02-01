package main

import (
	"os"

	"github.com/sibellavia/dory/internal/commands"
)

func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
