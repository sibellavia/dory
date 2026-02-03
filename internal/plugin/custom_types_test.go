package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverCustomTypes(t *testing.T) {
	doryRoot := filepath.Join(t.TempDir(), ".dory")
	pluginsDir := PluginsDirPath(doryRoot)

	enabledDir := filepath.Join(pluginsDir, "enabled")
	if err := os.MkdirAll(enabledDir, 0755); err != nil {
		t.Fatalf("mkdir enabled plugin dir: %v", err)
	}
	enabledManifest := []byte("name: alpha\napi_version: v1\ncommand: [\"alpha\"]\ncapabilities:\n  types: [\"incident\", \"postmortem\"]\n")
	if err := os.WriteFile(ManifestPath(enabledDir), enabledManifest, 0644); err != nil {
		t.Fatalf("write enabled manifest: %v", err)
	}

	disabledDir := filepath.Join(pluginsDir, "disabled")
	if err := os.MkdirAll(disabledDir, 0755); err != nil {
		t.Fatalf("mkdir disabled plugin dir: %v", err)
	}
	disabledManifest := []byte("name: beta\napi_version: v1\ncommand: [\"beta\"]\ncapabilities:\n  types: [\"runbook\"]\n")
	if err := os.WriteFile(ManifestPath(disabledDir), disabledManifest, 0644); err != nil {
		t.Fatalf("write disabled manifest: %v", err)
	}

	if err := SetPluginEnabled(doryRoot, "alpha", true); err != nil {
		t.Fatalf("enable alpha: %v", err)
	}

	customTypes, issues, err := DiscoverCustomTypes(doryRoot)
	if err != nil {
		t.Fatalf("discover custom types: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}
	if len(customTypes) != 3 {
		t.Fatalf("expected 3 custom types, got %d", len(customTypes))
	}

	foundEnabled := false
	foundDisabled := false
	for _, info := range customTypes {
		if info.Name == "incident" && info.Plugin == "alpha" && info.Enabled {
			foundEnabled = true
		}
		if info.Name == "runbook" && info.Plugin == "beta" && !info.Enabled {
			foundDisabled = true
		}
	}
	if !foundEnabled {
		t.Fatal("expected enabled custom type from alpha")
	}
	if !foundDisabled {
		t.Fatal("expected disabled custom type from beta")
	}
}
