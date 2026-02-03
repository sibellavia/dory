package plugin

import (
	"os"
	"path/filepath"
	"sort"
)

// Discover scans .dory/plugins and returns discovered plugin info plus issues.
func Discover(doryRoot string) ([]PluginInfo, []DiscoveryIssue, error) {
	cfg, err := LoadProjectConfig(doryRoot)
	if err != nil {
		return nil, nil, err
	}

	pluginsDir := PluginsDirPath(doryRoot)
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []PluginInfo{}, []DiscoveryIssue{}, nil
		}
		return nil, nil, err
	}

	var plugins []PluginInfo
	var issues []DiscoveryIssue
	seenNames := make(map[string]string)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join(pluginsDir, entry.Name())
		manifestPath := ManifestPath(dir)
		manifest, err := LoadManifest(manifestPath)
		if err != nil {
			issues = append(issues, DiscoveryIssue{
				Path:  manifestPath,
				Error: err.Error(),
			})
			continue
		}
		if prevDir, exists := seenNames[manifest.Name]; exists {
			issues = append(issues, DiscoveryIssue{
				Path:  manifestPath,
				Error: "duplicate plugin name " + manifest.Name + " (already discovered at " + prevDir + ")",
			})
			continue
		}
		seenNames[manifest.Name] = dir

		plugins = append(plugins, PluginInfo{
			Name:         manifest.Name,
			Version:      manifest.Version,
			APIVersion:   manifest.APIVersion,
			Description:  manifest.Description,
			Command:      append([]string(nil), manifest.Command...),
			Capabilities: manifest.Capabilities,
			Enabled:      cfg.Enabled[manifest.Name],
			Dir:          dir,
			ManifestPath: manifestPath,
		})
	}

	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].Path < issues[j].Path
	})

	return plugins, issues, nil
}
