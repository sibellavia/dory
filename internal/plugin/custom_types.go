package plugin

import "sort"

// CustomTypeInfo describes a custom type exposed by a plugin.
type CustomTypeInfo struct {
	Name    string `json:"name" yaml:"name"`
	Plugin  string `json:"plugin" yaml:"plugin"`
	Enabled bool   `json:"enabled" yaml:"enabled"`
}

// DiscoverCustomTypes collects custom knowledge types from discovered plugins.
func DiscoverCustomTypes(doryRoot string) ([]CustomTypeInfo, []DiscoveryIssue, error) {
	plugins, issues, err := Discover(doryRoot)
	if err != nil {
		return nil, nil, err
	}

	types := make([]CustomTypeInfo, 0)
	for _, p := range plugins {
		for _, name := range p.Capabilities.Types {
			if name == "" {
				continue
			}
			types = append(types, CustomTypeInfo{
				Name:    name,
				Plugin:  p.Name,
				Enabled: p.Enabled,
			})
		}
	}

	sort.Slice(types, func(i, j int) bool {
		if types[i].Name == types[j].Name {
			return types[i].Plugin < types[j].Plugin
		}
		return types[i].Name < types[j].Name
	})

	return types, issues, nil
}
