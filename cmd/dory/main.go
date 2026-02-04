package main

import (
	"os"
	"runtime/debug"

	"github.com/sibellavia/dory/internal/commands"
)

var version = "" // set via -ldflags "-X main.version=..."

func main() {
	commands.RootCmd.Version = getVersion()
	if err := commands.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func getVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return ""
}
