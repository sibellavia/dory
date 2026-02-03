package store

import (
	"path/filepath"
	"testing"
	"time"
)

func TestCreateCustomAndList(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	s := New(root)
	if err := s.Init("proj", "desc"); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer s.Close()

	id, err := s.CreateCustom("incident", "DB outage", "ops", "", []string{"L001"})
	if err != nil {
		t.Fatalf("create custom: %v", err)
	}
	if id == "" {
		t.Fatal("expected id")
	}

	items, err := s.List("", "incident", "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Type != "incident" {
		t.Fatalf("unexpected type: %+v", items[0])
	}
	if items[0].Topic != "ops" {
		t.Fatalf("unexpected topic: %+v", items[0])
	}

	topics, err := s.Topics()
	if err != nil {
		t.Fatalf("topics: %v", err)
	}
	if len(topics) != 1 || topics[0].Name != "ops" {
		t.Fatalf("unexpected topics: %+v", topics)
	}
}
