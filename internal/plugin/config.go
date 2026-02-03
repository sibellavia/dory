package plugin

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ProjectPluginsDirName  = "plugins"
	ProjectPluginsFileName = "plugins.yaml"
)

// DefaultProjectConfig returns an empty v1 project plugin config.
func DefaultProjectConfig() *ProjectConfig {
	return &ProjectConfig{
		Version: 1,
		Enabled: make(map[string]bool),
	}
}

// ConfigPath returns the plugin config path under .dory.
func ConfigPath(doryRoot string) string {
	return filepath.Join(doryRoot, ProjectPluginsFileName)
}

// PluginsDirPath returns the plugin directory path under .dory.
func PluginsDirPath(doryRoot string) string {
	return filepath.Join(doryRoot, ProjectPluginsDirName)
}

// LoadProjectConfig reads project plugin config or returns defaults if missing.
func LoadProjectConfig(doryRoot string) (*ProjectConfig, error) {
	path := ConfigPath(doryRoot)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultProjectConfig(), nil
		}
		return nil, err
	}

	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	if cfg.Enabled == nil {
		cfg.Enabled = make(map[string]bool)
	}
	return &cfg, nil
}

// SaveProjectConfig writes project plugin config under .dory.
func SaveProjectConfig(doryRoot string, cfg *ProjectConfig) error {
	if cfg == nil {
		cfg = DefaultProjectConfig()
	}
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	if cfg.Enabled == nil {
		cfg.Enabled = make(map[string]bool)
	}

	if err := os.MkdirAll(doryRoot, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(doryRoot), data, 0644)
}

// SetPluginEnabled enables or disables a plugin by name in project config.
func SetPluginEnabled(doryRoot, pluginName string, enabled bool) error {
	cfg, err := LoadProjectConfig(doryRoot)
	if err != nil {
		return err
	}
	if enabled {
		cfg.Enabled[pluginName] = true
	} else {
		delete(cfg.Enabled, pluginName)
	}
	return SaveProjectConfig(doryRoot, cfg)
}
