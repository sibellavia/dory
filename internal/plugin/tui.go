package plugin

import "sort"

// TUIExtensionInfo describes a TUI extension point exposed by a plugin.
type TUIExtensionInfo struct {
	Name    string `json:"name" yaml:"name"`
	Plugin  string `json:"plugin" yaml:"plugin"`
	Enabled bool   `json:"enabled" yaml:"enabled"`
}

// DiscoverTUIExtensions collects declared TUI extension points from plugins.
func DiscoverTUIExtensions(doryRoot string) ([]TUIExtensionInfo, []DiscoveryIssue, error) {
	plugins, issues, err := Discover(doryRoot)
	if err != nil {
		return nil, nil, err
	}

	extensions := make([]TUIExtensionInfo, 0)
	for _, p := range plugins {
		for _, name := range p.Capabilities.TUI {
			if name == "" {
				continue
			}
			extensions = append(extensions, TUIExtensionInfo{
				Name:    name,
				Plugin:  p.Name,
				Enabled: p.Enabled,
			})
		}
	}

	sort.Slice(extensions, func(i, j int) bool {
		if extensions[i].Name == extensions[j].Name {
			return extensions[i].Plugin < extensions[j].Plugin
		}
		return extensions[i].Name < extensions[j].Name
	})

	return extensions, issues, nil
}
