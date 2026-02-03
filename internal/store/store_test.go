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

func TestStoreRefs(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	s := New(root)
	if err := s.Init("project", ""); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer s.Close()

	// Create items with refs: L001 <- D001 <- P001
	_, err := s.Learn("base lesson", "topic", models.SeverityNormal, "", nil)
	if err != nil {
		t.Fatalf("learn: %v", err)
	}

	_, err = s.Decide("decision based on lesson", "topic", "", "", []string{"L001"})
	if err != nil {
		t.Fatalf("decide: %v", err)
	}

	_, err = s.Pattern("pattern from decision", "domain", "", []string{"D001"})
	if err != nil {
		t.Fatalf("pattern: %v", err)
	}

	// Test refs for D001 (middle of chain)
	refInfo, err := s.Refs("D001")
	if err != nil {
		t.Fatalf("refs D001: %v", err)
	}

	if refInfo.ID != "D001" {
		t.Fatalf("expected ID D001, got %s", refInfo.ID)
	}

	// D001 refs L001
	if len(refInfo.RefsTo) != 1 || refInfo.RefsTo[0].ID != "L001" {
		t.Fatalf("expected D001 to ref L001, got %+v", refInfo.RefsTo)
	}

	// P001 refs D001
	if len(refInfo.ReferencedBy) != 1 || refInfo.ReferencedBy[0].ID != "P001" {
		t.Fatalf("expected D001 to be referenced by P001, got %+v", refInfo.ReferencedBy)
	}

	// Test refs for L001 (start of chain - no refs_to)
	refInfo, err = s.Refs("L001")
	if err != nil {
		t.Fatalf("refs L001: %v", err)
	}

	if len(refInfo.RefsTo) != 0 {
		t.Fatalf("expected L001 to have no refs_to, got %+v", refInfo.RefsTo)
	}

	if len(refInfo.ReferencedBy) != 1 || refInfo.ReferencedBy[0].ID != "D001" {
		t.Fatalf("expected L001 to be referenced by D001, got %+v", refInfo.ReferencedBy)
	}

	// Test refs for P001 (end of chain - no referenced_by)
	refInfo, err = s.Refs("P001")
	if err != nil {
		t.Fatalf("refs P001: %v", err)
	}

	if len(refInfo.RefsTo) != 1 || refInfo.RefsTo[0].ID != "D001" {
		t.Fatalf("expected P001 to ref D001, got %+v", refInfo.RefsTo)
	}

	if len(refInfo.ReferencedBy) != 0 {
		t.Fatalf("expected P001 to have no referenced_by, got %+v", refInfo.ReferencedBy)
	}

	// Test refs for non-existent item
	_, err = s.Refs("X999")
	if err == nil {
		t.Fatal("expected error for non-existent item")
	}
}

func TestStoreExpand(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	s := New(root)
	if err := s.Init("project", ""); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer s.Close()

	// Create chain: L001 <- D001 <- P001
	_, err := s.Learn("base lesson", "topic", models.SeverityNormal, "Lesson body", nil)
	if err != nil {
		t.Fatalf("learn: %v", err)
	}

	_, err = s.Decide("decision", "topic", "", "Decision body", []string{"L001"})
	if err != nil {
		t.Fatalf("decide: %v", err)
	}

	_, err = s.Pattern("pattern", "domain", "Pattern body", []string{"D001"})
	if err != nil {
		t.Fatalf("pattern: %v", err)
	}

	// Test expand D001 with depth 1 (should get L001 and P001)
	result, err := s.Expand("D001", 1)
	if err != nil {
		t.Fatalf("expand D001: %v", err)
	}

	if result.Root.ID != "D001" {
		t.Fatalf("expected root D001, got %s", result.Root.ID)
	}

	if len(result.Connected) != 2 {
		t.Fatalf("expected 2 connected items, got %d", len(result.Connected))
	}

	// Check connected items (sorted: L001, P001)
	connectedIDs := make(map[string]bool)
	for _, item := range result.Connected {
		connectedIDs[item.ID] = true
	}
	if !connectedIDs["L001"] || !connectedIDs["P001"] {
		t.Fatalf("expected L001 and P001 in connected, got %v", result.Connected)
	}

	// Test expand L001 with depth 1 (should only get D001)
	result, err = s.Expand("L001", 1)
	if err != nil {
		t.Fatalf("expand L001: %v", err)
	}

	if len(result.Connected) != 1 || result.Connected[0].ID != "D001" {
		t.Fatalf("expected D001 connected to L001, got %v", result.Connected)
	}

	// Test expand L001 with depth 2 (should get D001 and P001)
	result, err = s.Expand("L001", 2)
	if err != nil {
		t.Fatalf("expand L001 depth 2: %v", err)
	}

	if len(result.Connected) != 2 {
		t.Fatalf("expected 2 connected at depth 2, got %d", len(result.Connected))
	}

	// Test expand non-existent item
	_, err = s.Expand("X999", 1)
	if err == nil {
		t.Fatal("expected error for non-existent item")
	}
}
