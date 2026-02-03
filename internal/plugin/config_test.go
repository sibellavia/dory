package plugin

import (
	"path/filepath"
	"testing"
)

func TestSetPluginEnabledRoundTrip(t *testing.T) {
	doryRoot := filepath.Join(t.TempDir(), ".dory")

	cfg, err := LoadProjectConfig(doryRoot)
	if err != nil {
		t.Fatalf("load default config: %v", err)
	}
	if cfg.Version != 1 {
		t.Fatalf("unexpected default version: %d", cfg.Version)
	}

	if err := SetPluginEnabled(doryRoot, "demo", true); err != nil {
		t.Fatalf("enable plugin: %v", err)
	}
	cfg, err = LoadProjectConfig(doryRoot)
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}
	if !cfg.Enabled["demo"] {
		t.Fatal("expected plugin to be enabled")
	}

	if err := SetPluginEnabled(doryRoot, "demo", false); err != nil {
		t.Fatalf("disable plugin: %v", err)
	}
	cfg, err = LoadProjectConfig(doryRoot)
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}
	if cfg.Enabled["demo"] {
		t.Fatal("expected plugin to be disabled")
	}
}
