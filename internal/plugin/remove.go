package plugin

import (
	"fmt"
	"os"
	"path/filepath"
)

// Remove deletes an installed plugin directory and clears enabled state.
// It first tries manifest-name lookup through discovery, then falls back to
// directory-name lookup for broken plugin folders.
func Remove(doryRoot, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("plugin name is required")
	}
	if filepath.Base(name) != name || name == "." || name == ".." {
		return "", fmt.Errorf("invalid plugin name %q", name)
	}

	plugins, _, err := Discover(doryRoot)
	if err != nil {
		return "", err
	}

	var targetDir string
	resolvedName := name
	for _, p := range plugins {
		if p.Name == name {
			targetDir = p.Dir
			resolvedName = p.Name
			break
		}
	}

	if targetDir == "" {
		targetDir = filepath.Join(PluginsDirPath(doryRoot), name)
		if _, err := os.Stat(targetDir); err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("plugin %q not found", name)
			}
			return "", err
		}
	}

	if err := os.RemoveAll(targetDir); err != nil {
		return "", err
	}
	_ = SetPluginEnabled(doryRoot, resolvedName, false)
	return targetDir, nil
}
