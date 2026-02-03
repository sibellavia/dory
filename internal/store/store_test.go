package store

import (
	"path/filepath"
	"strings"
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

func TestStoreDeleteThenCreateDoesNotReuseIDAcrossReopen(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	s := New(root)
	if err := s.Init("project", ""); err != nil {
		t.Fatalf("init: %v", err)
	}

	id1, err := s.Learn("one", "topic", models.SeverityNormal, "", nil)
	if err != nil {
		t.Fatalf("learn one: %v", err)
	}
	id2, err := s.Learn("two", "topic", models.SeverityNormal, "", nil)
	if err != nil {
		t.Fatalf("learn two: %v", err)
	}
	if !strings.HasPrefix(id1, "L-") || !strings.HasPrefix(id2, "L-") {
		t.Fatalf("expected typed lesson IDs, got %q and %q", id1, id2)
	}

	if err := s.Remove(id2); err != nil {
		t.Fatalf("remove %s: %v", id2, err)
	}

	id3, err := s.Learn("three", "topic", models.SeverityNormal, "", nil)
	if err != nil {
		t.Fatalf("learn three: %v", err)
	}
	if id3 == id2 {
		t.Fatalf("expected a new ID after delete, got reused %s", id3)
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

	ids := map[string]bool{}
	for _, item := range items {
		ids[item.ID] = true
	}
	if !ids[id1] || !ids[id3] {
		t.Fatalf("expected IDs %s and %s after reopen, got %+v", id1, id3, items)
	}
	if ids[id2] {
		t.Fatalf("did not expect deleted ID %s in reopen list", id2)
	}
}

func TestStoreRefs(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	s := New(root)
	if err := s.Init("project", ""); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer s.Close()

	// Create items with refs: lesson <- decision <- pattern.
	lessonID, err := s.Learn("base lesson", "topic", models.SeverityNormal, "", nil)
	if err != nil {
		t.Fatalf("learn: %v", err)
	}

	decisionID, err := s.Decide("decision based on lesson", "topic", "", "", []string{lessonID})
	if err != nil {
		t.Fatalf("decide: %v", err)
	}

	patternID, err := s.Pattern("pattern from decision", "domain", "", []string{decisionID})
	if err != nil {
		t.Fatalf("pattern: %v", err)
	}

	refInfo, err := s.Refs(decisionID)
	if err != nil {
		t.Fatalf("refs %s: %v", decisionID, err)
	}

	if refInfo.ID != decisionID {
		t.Fatalf("expected ID %s, got %s", decisionID, refInfo.ID)
	}

	if len(refInfo.RefsTo) != 1 || refInfo.RefsTo[0].ID != lessonID {
		t.Fatalf("expected %s to ref %s, got %+v", decisionID, lessonID, refInfo.RefsTo)
	}

	if len(refInfo.ReferencedBy) != 1 || refInfo.ReferencedBy[0].ID != patternID {
		t.Fatalf("expected %s to be referenced by %s, got %+v", decisionID, patternID, refInfo.ReferencedBy)
	}

	refInfo, err = s.Refs(lessonID)
	if err != nil {
		t.Fatalf("refs %s: %v", lessonID, err)
	}

	if len(refInfo.RefsTo) != 0 {
		t.Fatalf("expected %s to have no refs_to, got %+v", lessonID, refInfo.RefsTo)
	}

	if len(refInfo.ReferencedBy) != 1 || refInfo.ReferencedBy[0].ID != decisionID {
		t.Fatalf("expected %s to be referenced by %s, got %+v", lessonID, decisionID, refInfo.ReferencedBy)
	}

	refInfo, err = s.Refs(patternID)
	if err != nil {
		t.Fatalf("refs %s: %v", patternID, err)
	}

	if len(refInfo.RefsTo) != 1 || refInfo.RefsTo[0].ID != decisionID {
		t.Fatalf("expected %s to ref %s, got %+v", patternID, decisionID, refInfo.RefsTo)
	}

	if len(refInfo.ReferencedBy) != 0 {
		t.Fatalf("expected %s to have no referenced_by, got %+v", patternID, refInfo.ReferencedBy)
	}

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

	lessonID, err := s.Learn("base lesson", "topic", models.SeverityNormal, "Lesson body", nil)
	if err != nil {
		t.Fatalf("learn: %v", err)
	}

	decisionID, err := s.Decide("decision", "topic", "", "Decision body", []string{lessonID})
	if err != nil {
		t.Fatalf("decide: %v", err)
	}

	patternID, err := s.Pattern("pattern", "domain", "Pattern body", []string{decisionID})
	if err != nil {
		t.Fatalf("pattern: %v", err)
	}

	result, err := s.Expand(decisionID, 1)
	if err != nil {
		t.Fatalf("expand %s: %v", decisionID, err)
	}

	if result.Root.ID != decisionID {
		t.Fatalf("expected root %s, got %s", decisionID, result.Root.ID)
	}

	if len(result.Connected) != 2 {
		t.Fatalf("expected 2 connected items, got %d", len(result.Connected))
	}

	connectedIDs := make(map[string]bool)
	for _, item := range result.Connected {
		connectedIDs[item.ID] = true
	}
	if !connectedIDs[lessonID] || !connectedIDs[patternID] {
		t.Fatalf("expected %s and %s in connected, got %v", lessonID, patternID, result.Connected)
	}

	result, err = s.Expand(lessonID, 1)
	if err != nil {
		t.Fatalf("expand %s: %v", lessonID, err)
	}

	if len(result.Connected) != 1 || result.Connected[0].ID != decisionID {
		t.Fatalf("expected %s connected to %s, got %v", decisionID, lessonID, result.Connected)
	}

	result, err = s.Expand(lessonID, 2)
	if err != nil {
		t.Fatalf("expand %s depth 2: %v", lessonID, err)
	}

	if len(result.Connected) != 2 {
		t.Fatalf("expected 2 connected at depth 2, got %d", len(result.Connected))
	}

	_, err = s.Expand("X999", 1)
	if err == nil {
		t.Fatal("expected error for non-existent item")
	}
}

func TestStoreReadsSeeLatestWritesAcrossHandles(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	writer := New(root)
	if err := writer.Init("project", ""); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer writer.Close()

	reader := New(root)
	defer reader.Close()

	if _, err := writer.Learn("first", "topic", models.SeverityNormal, "", nil); err != nil {
		t.Fatalf("learn first: %v", err)
	}

	items, err := reader.List("", "lesson", "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("reader initial list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if _, err := writer.Learn("second", "topic", models.SeverityNormal, "", nil); err != nil {
		t.Fatalf("learn second: %v", err)
	}

	items, err = reader.List("", "lesson", "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("reader refreshed list: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected reader to see latest writes, got %d items", len(items))
	}
}
