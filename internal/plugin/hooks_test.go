package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeHookPlugin(t *testing.T, doryRoot, dirName, pluginName, hookEvent, scriptBody string, enabled bool) {
	t.Helper()

	pluginDir := filepath.Join(PluginsDirPath(doryRoot), dirName)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("mkdir plugin dir: %v", err)
	}

	manifest := "name: " + pluginName + "\napi_version: v1\ncommand: [\"./hook.sh\"]\ncapabilities:\n  hooks: [\"" + hookEvent + "\"]\n"
	if err := os.WriteFile(ManifestPath(pluginDir), []byte(manifest), 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	script := "#!/bin/sh\nset -eu\nread line\n" + scriptBody + "\n"
	scriptPath := filepath.Join(pluginDir, "hook.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	if err := os.Chmod(scriptPath, 0755); err != nil {
		t.Fatalf("chmod script: %v", err)
	}

	if enabled {
		if err := SetPluginEnabled(doryRoot, pluginName, true); err != nil {
			t.Fatalf("enable plugin: %v", err)
		}
	}
}

func TestRunHooksBeforeCreateCanBlock(t *testing.T) {
	doryRoot := filepath.Join(t.TempDir(), ".dory")
	writeHookPlugin(t, doryRoot, "gatekeeper", "gatekeeper", string(HookBeforeCreate), `echo '{"id":"req-1","result":{"allow":false,"message":"blocked by policy"}}'`, true)

	results, err := RunHooks(doryRoot, HookBeforeCreate, map[string]interface{}{"type": "lesson"}, 2*time.Second)
	if err == nil {
		t.Fatal("expected hook block error")
	}
	if !strings.Contains(err.Error(), "blocked") {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one result, got %d", len(results))
	}
	if results[0].Status != "blocked" {
		t.Fatalf("unexpected status: %+v", results[0])
	}
}

func TestRunHooksAfterCreateIsFailSoft(t *testing.T) {
	doryRoot := filepath.Join(t.TempDir(), ".dory")
	writeHookPlugin(t, doryRoot, "broken", "broken", string(HookAfterCreate), `exit 1`, true)

	results, err := RunHooks(doryRoot, HookAfterCreate, map[string]interface{}{"id": "L001"}, 2*time.Second)
	if err != nil {
		t.Fatalf("expected fail-soft behavior, got error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one result, got %d", len(results))
	}
	if results[0].Status != "error" {
		t.Fatalf("expected error status, got %+v", results[0])
	}
}

func TestRunHooksOnlyUsesEnabledMatchingPlugins(t *testing.T) {
	doryRoot := filepath.Join(t.TempDir(), ".dory")
	writeHookPlugin(t, doryRoot, "enabled", "enabled", string(HookBeforeRemove), `echo '{"id":"req-1","result":{"allow":true}}'`, true)
	writeHookPlugin(t, doryRoot, "disabled", "disabled", string(HookBeforeRemove), `echo '{"id":"req-1","result":{"allow":false}}'`, false)
	writeHookPlugin(t, doryRoot, "other-event", "other", string(HookAfterCreate), `echo '{"id":"req-1","result":{"allow":false}}'`, true)

	results, err := RunHooks(doryRoot, HookBeforeRemove, map[string]interface{}{"id": "L001"}, 2*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one matching result, got %d", len(results))
	}
	if results[0].Plugin != "enabled" {
		t.Fatalf("unexpected plugin executed: %+v", results[0])
	}
}
