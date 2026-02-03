package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func makePluginSource(t *testing.T, root, name, version string) string {
	t.Helper()

	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir plugin source: %v", err)
	}

	manifest := "name: " + name + "\napi_version: v1\nversion: " + version + "\ncommand: [\"./run.sh\"]\ncapabilities:\n  commands: [\"sync\"]\n"
	if err := os.WriteFile(ManifestPath(dir), []byte(manifest), 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "run.sh"), []byte("#!/bin/sh\necho ok\n"), 0755); err != nil {
		t.Fatalf("write run.sh: %v", err)
	}

	return dir
}

func TestInstallFromDirectoryAndManifestPath(t *testing.T) {
	tmp := t.TempDir()
	srcRoot := filepath.Join(tmp, "src")
	doryRoot := filepath.Join(tmp, ".dory")
	srcDir := makePluginSource(t, srcRoot, "demo", "1.0.0")

	info, err := Install(doryRoot, srcDir, InstallOptions{})
	if err != nil {
		t.Fatalf("install from dir: %v", err)
	}
	if info.Name != "demo" {
		t.Fatalf("unexpected plugin name: %q", info.Name)
	}

	scriptPath := filepath.Join(PluginsDirPath(doryRoot), "demo", "run.sh")
	st, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("stat copied script: %v", err)
	}
	if st.Mode()&0111 == 0 {
		t.Fatalf("expected copied script to be executable, mode=%v", st.Mode())
	}

	manifestPath := filepath.Join(srcDir, ManifestFileName)
	info, err = Install(doryRoot, manifestPath, InstallOptions{Force: true})
	if err != nil {
		t.Fatalf("install from manifest path: %v", err)
	}
	if info.Name != "demo" {
		t.Fatalf("unexpected plugin name after manifest install: %q", info.Name)
	}
}

func TestInstallForceOverwrite(t *testing.T) {
	tmp := t.TempDir()
	srcRoot := filepath.Join(tmp, "src")
	doryRoot := filepath.Join(tmp, ".dory")

	srcDirV1 := makePluginSource(t, srcRoot, "demo", "1.0.0")
	if _, err := Install(doryRoot, srcDirV1, InstallOptions{}); err != nil {
		t.Fatalf("initial install: %v", err)
	}

	srcDirV2 := makePluginSource(t, srcRoot, "demo", "2.0.0")
	if _, err := Install(doryRoot, srcDirV2, InstallOptions{}); err == nil {
		t.Fatal("expected install without force to fail when plugin exists")
	}

	info, err := Install(doryRoot, srcDirV2, InstallOptions{Force: true})
	if err != nil {
		t.Fatalf("force install: %v", err)
	}
	if info.Version != "2.0.0" {
		t.Fatalf("expected overwritten version 2.0.0, got %q", info.Version)
	}
}
