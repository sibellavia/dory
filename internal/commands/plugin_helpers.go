package commands

import (
	"path/filepath"
	"strings"

	"github.com/sibellavia/dory/internal/plugin"
)

const doryRoot = ".dory"

func capabilitySummary(c plugin.Capabilities) string {
	var parts []string
	if len(c.Commands) > 0 {
		parts = append(parts, "commands")
	}
	if len(c.Hooks) > 0 {
		parts = append(parts, "hooks")
	}
	if len(c.Types) > 0 {
		parts = append(parts, "types")
	}
	if len(c.TUI) > 0 {
		parts = append(parts, "tui")
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, ",")
}

func pluginPathDisplay(path string) string {
	if rel, err := filepath.Rel(".", path); err == nil {
		return rel
	}
	return path
}

func findPluginByName(plugins []plugin.PluginInfo, name string) *plugin.PluginInfo {
	for i := range plugins {
		if plugins[i].Name == name {
			return &plugins[i]
		}
	}
	return nil
}
