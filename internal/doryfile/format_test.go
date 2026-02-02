package doryfile

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateAndAppend(t *testing.T) {
	dir := t.TempDir()

	df, err := Create(dir, "test-project")
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	entry := &Entry{
		ID:       "L001",
		Type:     "lesson",
		Topic:    "testing",
		Severity: "high",
		Oneliner: "Always write tests",
		Created:  time.Now(),
		Body:     "Testing is important for reliability.",
	}

	if err := df.Append(entry); err != nil {
		t.Fatalf("failed to append: %v", err)
	}
	df.Close()

	// Check files exist
	if _, err := os.Stat(filepath.Join(dir, KnowledgeFile)); err != nil {
		t.Errorf("knowledge file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, IndexFile)); err != nil {
		t.Errorf("index file missing: %v", err)
	}

	// Read knowledge file
	content, _ := os.ReadFile(filepath.Join(dir, KnowledgeFile))
	t.Logf("Knowledge file:\n%s", string(content))

	// Read index file
	index, _ := os.ReadFile(filepath.Join(dir, IndexFile))
	t.Logf("Index file:\n%s", string(index))

	// Reopen and verify
	df2, err := Open(dir)
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}
	defer df2.Close()

	if len(df2.Lessons()) != 1 {
		t.Errorf("expected 1 lesson, got %d", len(df2.Lessons()))
	}

	got, err := df2.Get("L001")
	if err != nil {
		t.Fatalf("failed to get L001: %v", err)
	}
	if got.Oneliner != entry.Oneliner {
		t.Errorf("oneliner mismatch: got %q, want %q", got.Oneliner, entry.Oneliner)
	}
}

func TestMultipleEntries(t *testing.T) {
	dir := t.TempDir()

	df, err := Create(dir, "test-project")
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	entries := []*Entry{
		{ID: "L001", Type: "lesson", Topic: "api", Oneliner: "Lesson 1", Created: time.Now()},
		{ID: "L002", Type: "lesson", Topic: "db", Oneliner: "Lesson 2", Created: time.Now()},
		{ID: "D001", Type: "decision", Topic: "arch", Oneliner: "Decision 1", Created: time.Now()},
	}

	for _, e := range entries {
		if err := df.Append(e); err != nil {
			t.Fatalf("failed to append %s: %v", e.ID, err)
		}
	}
	df.Close()

	// Reopen and verify
	df2, err := Open(dir)
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}
	defer df2.Close()

	if len(df2.Lessons()) != 2 {
		t.Errorf("expected 2 lessons, got %d", len(df2.Lessons()))
	}
	if len(df2.Decisions()) != 1 {
		t.Errorf("expected 1 decision, got %d", len(df2.Decisions()))
	}

	for _, e := range entries {
		got, err := df2.Get(e.ID)
		if err != nil {
			t.Errorf("failed to get %s: %v", e.ID, err)
			continue
		}
		if got.Oneliner != e.Oneliner {
			t.Errorf("%s oneliner mismatch: got %q, want %q", e.ID, got.Oneliner, e.Oneliner)
		}
	}
}

func TestDeleteAndCompact(t *testing.T) {
	dir := t.TempDir()

	df, err := Create(dir, "test-project")
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	entries := []*Entry{
		{ID: "L001", Type: "lesson", Topic: "api", Oneliner: "Lesson 1", Created: time.Now()},
		{ID: "L002", Type: "lesson", Topic: "db", Oneliner: "Lesson 2", Created: time.Now()},
		{ID: "L003", Type: "lesson", Topic: "api", Oneliner: "Lesson 3", Created: time.Now()},
	}

	for _, e := range entries {
		if err := df.Append(e); err != nil {
			t.Fatalf("failed to append %s: %v", e.ID, err)
		}
	}

	// Delete L002
	if err := df.Delete("L002"); err != nil {
		t.Fatalf("failed to delete L002: %v", err)
	}

	// Verify L002 is gone from in-memory index
	if _, ok := df.Lessons()["L002"]; ok {
		t.Error("L002 should be deleted from index")
	}

	// Knowledge file size before compact
	stat1, _ := os.Stat(filepath.Join(dir, KnowledgeFile))
	t.Logf("Knowledge file size before compact: %d", stat1.Size())

	// Compact
	if err := df.Compact(); err != nil {
		t.Fatalf("failed to compact: %v", err)
	}

	// Knowledge file size after compact
	stat2, _ := os.Stat(filepath.Join(dir, KnowledgeFile))
	t.Logf("Knowledge file size after compact: %d", stat2.Size())

	if stat2.Size() >= stat1.Size() {
		t.Logf("Warning: knowledge file did not shrink")
	}

	// Verify entries
	if _, err := df.Get("L001"); err != nil {
		t.Errorf("L001 should exist: %v", err)
	}
	if _, err := df.Get("L003"); err != nil {
		t.Errorf("L003 should exist: %v", err)
	}
	if _, err := df.Get("L002"); err == nil {
		t.Error("L002 should not exist")
	}

	df.Close()
}

func TestState(t *testing.T) {
	dir := t.TempDir()

	df, err := Create(dir, "test-project")
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	state := &State{
		Goal:     "Build the API",
		Progress: "Endpoints done",
		Next:     []string{"Add tests", "Deploy"},
	}

	if err := df.UpdateState(state); err != nil {
		t.Fatalf("failed to update state: %v", err)
	}
	df.Close()

	// Reopen and verify
	df2, err := Open(dir)
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}
	defer df2.Close()

	if df2.Index.State.Goal != "Build the API" {
		t.Errorf("goal mismatch: got %q", df2.Index.State.Goal)
	}
	if len(df2.Index.State.Next) != 2 {
		t.Errorf("next steps count mismatch: got %d", len(df2.Index.State.Next))
	}
}

func TestAppendOnlyContent(t *testing.T) {
	dir := t.TempDir()

	df, err := Create(dir, "test-project")
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	// Add first entry
	df.Append(&Entry{ID: "L001", Type: "lesson", Topic: "api", Oneliner: "First", Created: time.Now()})
	stat1, _ := os.Stat(filepath.Join(dir, KnowledgeFile))
	size1 := stat1.Size()

	// Add second entry
	df.Append(&Entry{ID: "L002", Type: "lesson", Topic: "api", Oneliner: "Second", Created: time.Now()})
	stat2, _ := os.Stat(filepath.Join(dir, KnowledgeFile))
	size2 := stat2.Size()

	// Knowledge file should only grow, never shrink
	if size2 <= size1 {
		t.Errorf("knowledge file should grow: %d -> %d", size1, size2)
	}

	// Add third entry
	df.Append(&Entry{ID: "L003", Type: "lesson", Topic: "api", Oneliner: "Third", Created: time.Now()})
	stat3, _ := os.Stat(filepath.Join(dir, KnowledgeFile))
	size3 := stat3.Size()

	if size3 <= size2 {
		t.Errorf("knowledge file should grow: %d -> %d", size2, size3)
	}

	t.Logf("Knowledge file sizes: %d -> %d -> %d (append-only confirmed)", size1, size2, size3)
	df.Close()
}
