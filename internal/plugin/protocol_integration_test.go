package plugin

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestProtocolFixtureHealthAndRun(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is POSIX-only")
	}

	fixture := filepath.Join("testdata", "fixture-plugin.sh")
	fixtureAbs, err := filepath.Abs(fixture)
	if err != nil {
		t.Fatalf("abs fixture path: %v", err)
	}
	if err := os.Chmod(fixtureAbs, 0755); err != nil {
		t.Fatalf("chmod fixture: %v", err)
	}
	fixtureDir := filepath.Dir(fixtureAbs)

	info := PluginInfo{
		Name:       "fixture",
		APIVersion: APIVersionV1,
		Command:    []string{"./fixture-plugin.sh"},
		Dir:        fixtureDir,
	}

	health, stderr, _, err := Invoke(info, healthMethod, map[string]interface{}{
		"api_version": APIVersionV1,
	}, 2*time.Second)
	if err != nil {
		t.Fatalf("health invoke: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if health["status"] != "ok" {
		t.Fatalf("unexpected health result: %#v", health)
	}

	run, stderr, _, err := Invoke(info, "dory.command.run", map[string]interface{}{
		"api_version": APIVersionV1,
		"plugin":      "fixture",
		"command":     "sync",
		"args":        []string{"--dry-run"},
	}, 2*time.Second)
	if err != nil {
		t.Fatalf("run invoke: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if run["message"] != "fixture done" {
		t.Fatalf("unexpected run result: %#v", run)
	}
}
