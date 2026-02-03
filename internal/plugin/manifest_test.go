package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ManifestFileName)
	content := []byte("name: demo\napi_version: v1\ncommand: [\"demo-plugin\", \"serve\"]\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	manifest, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if manifest.Name != "demo" {
		t.Fatalf("unexpected name: %q", manifest.Name)
	}
	if len(manifest.Command) != 2 {
		t.Fatalf("unexpected command: %v", manifest.Command)
	}
}

func TestLoadManifestValidationError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ManifestFileName)
	content := []byte("name: demo\napi_version: v1\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	if _, err := LoadManifest(path); err == nil {
		t.Fatal("expected validation error for missing command")
	}
}

func TestLoadManifestUnsupportedAPIVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ManifestFileName)
	content := []byte("name: demo\napi_version: v2\ncommand: [\"demo-plugin\"]\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	_, err := LoadManifest(path)
	if err == nil {
		t.Fatal("expected validation error for unsupported api version")
	}
	if !strings.Contains(err.Error(), "unsupported api_version") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadManifestInvalidName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ManifestFileName)
	content := []byte("name: ../bad\napi_version: v1\ncommand: [\"demo-plugin\"]\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	_, err := LoadManifest(path)
	if err == nil {
		t.Fatal("expected validation error for invalid name")
	}
	if !strings.Contains(err.Error(), "must match") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadManifestRejectsStoreCapability(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ManifestFileName)
	content := []byte("name: demo\napi_version: v1\ncommand: [\"demo-plugin\"]\ncapabilities:\n  store: true\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	_, err := LoadManifest(path)
	if err == nil {
		t.Fatal("expected validation error for store capability")
	}
	if !strings.Contains(err.Error(), "capabilities.store is not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}
