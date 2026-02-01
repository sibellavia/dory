package main

import (
	"os"
	"runtime/debug"

	"github.com/sibellavia/dory/internal/commands"
)

var version = "0.2.3"

func main() {
	commands.RootCmd.Version = getVersion()
	if err := commands.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return version
}
