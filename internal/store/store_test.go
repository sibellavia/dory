package store

import (
	"path/filepath"
	"testing"

	"github.com/sibellavia/dory/internal/doryfile"
)

func TestStoreInitPersistsDescription(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	s := New(root)

	if err := s.Init("my-project", "my-description"); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	df, err := doryfile.Open(root)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer df.Close()

	if df.Index.Project != "my-project" {
		t.Fatalf("project mismatch: got %q", df.Index.Project)
	}
	if df.Index.Description != "my-description" {
		t.Fatalf("description mismatch: got %q", df.Index.Description)
	}
}
