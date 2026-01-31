package main

import (
	"os"

	"github.com/simonebellavia/dory/internal/commands"
)

func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
