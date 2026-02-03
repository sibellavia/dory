package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverIncludesEnabledStateAndIssues(t *testing.T) {
	doryRoot := filepath.Join(t.TempDir(), ".dory")
	pluginsDir := PluginsDirPath(doryRoot)

	validDir := filepath.Join(pluginsDir, "valid")
	if err := os.MkdirAll(validDir, 0755); err != nil {
		t.Fatalf("mkdir valid dir: %v", err)
	}
	validManifest := []byte("name: demo\napi_version: v1\ncommand: [\"demo-plugin\"]\n")
	if err := os.WriteFile(ManifestPath(validDir), validManifest, 0644); err != nil {
		t.Fatalf("write valid manifest: %v", err)
	}

	invalidDir := filepath.Join(pluginsDir, "invalid")
	if err := os.MkdirAll(invalidDir, 0755); err != nil {
		t.Fatalf("mkdir invalid dir: %v", err)
	}
	invalidManifest := []byte("name: broken\napi_version: v1\n")
	if err := os.WriteFile(ManifestPath(invalidDir), invalidManifest, 0644); err != nil {
		t.Fatalf("write invalid manifest: %v", err)
	}

	if err := SetPluginEnabled(doryRoot, "demo", true); err != nil {
		t.Fatalf("set enabled: %v", err)
	}

	plugins, issues, err := Discover(doryRoot)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 valid plugin, got %d", len(plugins))
	}
	if plugins[0].Name != "demo" {
		t.Fatalf("unexpected plugin name: %q", plugins[0].Name)
	}
	if !plugins[0].Enabled {
		t.Fatal("expected plugin demo to be enabled")
	}
	if len(issues) == 0 {
		t.Fatal("expected at least one discovery issue")
	}
}

func TestDiscoverReportsDuplicateNames(t *testing.T) {
	doryRoot := filepath.Join(t.TempDir(), ".dory")
	pluginsDir := PluginsDirPath(doryRoot)

	firstDir := filepath.Join(pluginsDir, "one")
	secondDir := filepath.Join(pluginsDir, "two")
	if err := os.MkdirAll(firstDir, 0755); err != nil {
		t.Fatalf("mkdir first dir: %v", err)
	}
	if err := os.MkdirAll(secondDir, 0755); err != nil {
		t.Fatalf("mkdir second dir: %v", err)
	}

	manifest := []byte("name: demo\napi_version: v1\ncommand: [\"demo-plugin\"]\n")
	if err := os.WriteFile(ManifestPath(firstDir), manifest, 0644); err != nil {
		t.Fatalf("write first manifest: %v", err)
	}
	if err := os.WriteFile(ManifestPath(secondDir), manifest, 0644); err != nil {
		t.Fatalf("write second manifest: %v", err)
	}

	plugins, issues, err := Discover(doryRoot)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin after duplicate filtering, got %d", len(plugins))
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 duplicate issue, got %d", len(issues))
	}
}
