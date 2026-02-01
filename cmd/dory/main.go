package main

import (
	"os"

	"github.com/sibellavia/dory/internal/commands"
)

var version = "dev"

func main() {
	commands.RootCmd.Version = version
	if err := commands.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
