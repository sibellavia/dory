package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverTUIExtensions(t *testing.T) {
	doryRoot := filepath.Join(t.TempDir(), ".dory")
	pluginsDir := PluginsDirPath(doryRoot)

	pluginDir := filepath.Join(pluginsDir, "ui")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("mkdir plugin dir: %v", err)
	}
	manifest := []byte("name: ui\napi_version: v1\ncommand: [\"ui\"]\ncapabilities:\n  tui: [\"dashboard\", \"panel.incident\"]\n")
	if err := os.WriteFile(ManifestPath(pluginDir), manifest, 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	extensions, issues, err := DiscoverTUIExtensions(doryRoot)
	if err != nil {
		t.Fatalf("discover tui extensions: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}
	if len(extensions) != 2 {
		t.Fatalf("expected 2 extensions, got %d", len(extensions))
	}
	if extensions[0].Plugin != "ui" {
		t.Fatalf("unexpected plugin: %+v", extensions[0])
	}
}
