package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	ManifestFileName = "plugin.yaml"
	APIVersionV1     = "v1"
)

var pluginNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// LoadManifest reads and validates a plugin manifest from disk.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest %s: %w", path, err)
	}

	if err := validateManifest(&manifest, path); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func validateManifest(m *Manifest, path string) error {
	if m.Name == "" {
		return fmt.Errorf("invalid manifest %s: name is required", path)
	}
	if strings.Contains(m.Name, " ") {
		return fmt.Errorf("invalid manifest %s: name must not contain spaces", path)
	}
	if !pluginNamePattern.MatchString(m.Name) {
		return fmt.Errorf("invalid manifest %s: name %q must match %q", path, m.Name, pluginNamePattern.String())
	}
	if err := ValidateAPIVersion(m.APIVersion); err != nil {
		return fmt.Errorf("invalid manifest %s: %w", path, err)
	}
	if len(m.Command) == 0 || strings.TrimSpace(m.Command[0]) == "" {
		return fmt.Errorf("invalid manifest %s: command is required", path)
	}
	if m.Capabilities.Store {
		return fmt.Errorf("invalid manifest %s: capabilities.store is not supported in this release", path)
	}
	return nil
}

// ManifestPath returns a conventional manifest path for a plugin directory.
func ManifestPath(pluginDir string) string {
	return filepath.Join(pluginDir, ManifestFileName)
}
