package plugin

// Manifest represents a plugin manifest file.
type Manifest struct {
	Name         string       `yaml:"name" json:"name"`
	Version      string       `yaml:"version,omitempty" json:"version,omitempty"`
	APIVersion   string       `yaml:"api_version" json:"api_version"`
	Description  string       `yaml:"description,omitempty" json:"description,omitempty"`
	Command      []string     `yaml:"command" json:"command"`
	Capabilities Capabilities `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
}

// Capabilities describes what a plugin exposes to Dory.
type Capabilities struct {
	Commands []string `yaml:"commands,omitempty" json:"commands,omitempty"`
	Hooks    []string `yaml:"hooks,omitempty" json:"hooks,omitempty"`
	Types    []string `yaml:"types,omitempty" json:"types,omitempty"`
	// Store remains parsed so we can explicitly reject it at validation time.
	Store bool `yaml:"store,omitempty" json:"store,omitempty"`
}

// ProjectConfig stores plugin enablement for a single project.
type ProjectConfig struct {
	Version int             `yaml:"version" json:"version"`
	Enabled map[string]bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

// PluginInfo is a discovered plugin plus project state.
type PluginInfo struct {
	Name         string       `json:"name" yaml:"name"`
	Version      string       `json:"version,omitempty" yaml:"version,omitempty"`
	APIVersion   string       `json:"api_version" yaml:"api_version"`
	Description  string       `json:"description,omitempty" yaml:"description,omitempty"`
	Command      []string     `json:"command" yaml:"command"`
	Capabilities Capabilities `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
	Enabled      bool         `json:"enabled" yaml:"enabled"`
	Dir          string       `json:"dir" yaml:"dir"`
	ManifestPath string       `json:"manifest_path" yaml:"manifest_path"`
}

// DiscoveryIssue is a non-fatal discovery warning.
type DiscoveryIssue struct {
	Path  string `json:"path" yaml:"path"`
	Error string `json:"error" yaml:"error"`
}

// HealthStatus represents a plugin health-check result.
type HealthStatus struct {
	Name       string `json:"name" yaml:"name"`
	Reachable  bool   `json:"reachable" yaml:"reachable"`
	Status     string `json:"status" yaml:"status"` // ok, error, warning
	Message    string `json:"message,omitempty" yaml:"message,omitempty"`
	Error      string `json:"error,omitempty" yaml:"error,omitempty"`
	DurationMS int64  `json:"duration_ms,omitempty" yaml:"duration_ms,omitempty"`
}
