package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func makeScript(t *testing.T, body string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "plugin.sh")
	script := "#!/bin/sh\nset -eu\n" + body + "\n"
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	if err := os.Chmod(path, 0755); err != nil {
		t.Fatalf("chmod script: %v", err)
	}
	return path
}

func TestInvokeSuccess(t *testing.T) {
	script := makeScript(t, `read line
echo '{"id":"req-1","result":{"status":"ok","message":"pong","output":"hello"}}'`)
	info := PluginInfo{Name: "demo", APIVersion: APIVersionV1, Command: []string{script}}

	result, stderr, duration, err := Invoke(info, "dory.health", map[string]interface{}{"api_version": APIVersionV1}, 2*time.Second)
	if err != nil {
		t.Fatalf("invoke: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if duration < 0 {
		t.Fatalf("unexpected duration: %d", duration)
	}
	if result["status"] != "ok" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestInvokeResponseError(t *testing.T) {
	script := makeScript(t, `read line
echo '{"id":"req-1","error":{"code":400,"message":"bad request"}}'`)
	info := PluginInfo{Name: "demo", APIVersion: APIVersionV1, Command: []string{script}}

	_, _, _, err := Invoke(info, "dory.command.run", map[string]interface{}{}, 2*time.Second)
	if err == nil {
		t.Fatal("expected invoke error")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHealthCheckOK(t *testing.T) {
	script := makeScript(t, `read line
echo '{"id":"req-1","result":{"status":"ok","message":"healthy"}}'`)
	info := PluginInfo{Name: "demo", APIVersion: APIVersionV1, Command: []string{script}}

	status := HealthCheck(info, 2*time.Second)
	if !status.Reachable {
		t.Fatalf("expected reachable status, got %+v", status)
	}
	if status.Status != "ok" {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestInvokeUnsupportedAPIVersion(t *testing.T) {
	info := PluginInfo{
		Name:       "demo",
		APIVersion: "v9",
		Command:    []string{"echo"},
	}

	_, _, _, err := Invoke(info, "dory.health", map[string]interface{}{"api_version": APIVersionV1}, 2*time.Second)
	if err == nil {
		t.Fatal("expected invoke error")
	}
	if !strings.Contains(err.Error(), "unsupported api_version") {
		t.Fatalf("unexpected error: %v", err)
	}
}
