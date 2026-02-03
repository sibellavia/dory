package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/sibellavia/dory/internal/doryfile"
	"github.com/sibellavia/dory/internal/models"
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

func TestStoreDeleteAndReuseIDPersistsAcrossReopen(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	s := New(root)
	if err := s.Init("project", ""); err != nil {
		t.Fatalf("init: %v", err)
	}

	if _, err := s.Learn("one", "topic", models.SeverityNormal, "", nil); err != nil {
		t.Fatalf("learn one: %v", err)
	}
	if _, err := s.Learn("two", "topic", models.SeverityNormal, "", nil); err != nil {
		t.Fatalf("learn two: %v", err)
	}

	if err := s.Remove("L002"); err != nil {
		t.Fatalf("remove L002: %v", err)
	}

	id, err := s.Learn("three", "topic", models.SeverityNormal, "", nil)
	if err != nil {
		t.Fatalf("learn three: %v", err)
	}
	if id != "L002" {
		t.Fatalf("expected reused id L002, got %s", id)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	s2 := New(root)
	defer s2.Close()

	items, err := s2.List("", "lesson", "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 lessons after reopen, got %d", len(items))
	}
	if items[1].ID != "L002" || items[1].Oneliner != "three" {
		t.Fatalf("expected reused L002 to persist, got %+v", items[1])
	}
}
