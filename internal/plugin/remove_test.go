package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoveInstalledPluginAndDisable(t *testing.T) {
	tmp := t.TempDir()
	srcRoot := filepath.Join(tmp, "src")
	doryRoot := filepath.Join(tmp, ".dory")

	srcDir := makePluginSource(t, srcRoot, "demo", "1.0.0")
	if _, err := Install(doryRoot, srcDir, InstallOptions{}); err != nil {
		t.Fatalf("install: %v", err)
	}
	if err := SetPluginEnabled(doryRoot, "demo", true); err != nil {
		t.Fatalf("enable: %v", err)
	}

	removedPath, err := Remove(doryRoot, "demo")
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, err := os.Stat(removedPath); !os.IsNotExist(err) {
		t.Fatalf("expected plugin directory removed, stat err=%v", err)
	}

	cfg, err := LoadProjectConfig(doryRoot)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Enabled["demo"] {
		t.Fatal("expected plugin to be disabled after remove")
	}
}

func TestRemoveBrokenPluginByDirectoryName(t *testing.T) {
	tmp := t.TempDir()
	doryRoot := filepath.Join(tmp, ".dory")
	brokenDir := filepath.Join(PluginsDirPath(doryRoot), "broken")
	if err := os.MkdirAll(brokenDir, 0755); err != nil {
		t.Fatalf("mkdir broken dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(brokenDir, "junk.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("write broken file: %v", err)
	}

	removedPath, err := Remove(doryRoot, "broken")
	if err != nil {
		t.Fatalf("remove broken plugin: %v", err)
	}
	if removedPath == "" {
		t.Fatal("expected removed path")
	}
	if _, err := os.Stat(brokenDir); !os.IsNotExist(err) {
		t.Fatalf("expected broken dir removed, stat err=%v", err)
	}
}

func TestRemoveRejectsInvalidName(t *testing.T) {
	_, err := Remove(t.TempDir(), "../demo")
	if err == nil {
		t.Fatal("expected invalid name error")
	}
	if !strings.Contains(err.Error(), "invalid plugin name") {
		t.Fatalf("unexpected error: %v", err)
	}
}
