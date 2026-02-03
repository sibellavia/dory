package doryfile

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestGetPreservesBodyLinesThatLookLikeDelimiter(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	df, err := Create(root, "test", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	entry1 := &Entry{
		ID:       "L001",
		Type:     "lesson",
		Topic:    "api",
		Severity: "normal",
		Oneliner: "Delimiter body line",
		Created:  time.Now(),
		Body:     "# Title\n\nLine before\n===\nLine after\n",
	}
	entry2 := &Entry{
		ID:       "L002",
		Type:     "lesson",
		Topic:    "api",
		Severity: "normal",
		Oneliner: "Second",
		Created:  time.Now(),
		Body:     "# Second",
	}
	if err := df.Append(entry1); err != nil {
		t.Fatalf("append entry1: %v", err)
	}
	if err := df.Append(entry2); err != nil {
		t.Fatalf("append entry2: %v", err)
	}
	if err := df.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	df2, err := Open(root)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer df2.Close()

	got, err := df2.Get("L001")
	if err != nil {
		t.Fatalf("get L001: %v", err)
	}
	if !strings.Contains(got.Body, "===\nLine after") {
		t.Fatalf("expected full body to be preserved, got:\n%s", got.Body)
	}
}

func TestAppendSameIDKeepsLatestVersion(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	df, err := Create(root, "test", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer df.Close()

	oldEntry := &Entry{
		ID:       "L001",
		Type:     "lesson",
		Topic:    "api",
		Severity: "normal",
		Oneliner: "Old",
		Created:  time.Now(),
		Body:     "old",
	}
	newEntry := &Entry{
		ID:       "L001",
		Type:     "lesson",
		Topic:    "api",
		Severity: "high",
		Oneliner: "New",
		Created:  time.Now(),
		Body:     "new",
	}

	if err := df.Append(oldEntry); err != nil {
		t.Fatalf("append old: %v", err)
	}
	if err := df.Append(newEntry); err != nil {
		t.Fatalf("append new: %v", err)
	}

	got, err := df.Get("L001")
	if err != nil {
		t.Fatalf("get L001: %v", err)
	}
	if got.Oneliner != "New" || got.Body != "new" || got.Severity != "high" {
		t.Fatalf("expected latest version, got %+v", got)
	}
}

func TestCompactWritesDeterministicOrder(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	df, err := Create(root, "test", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	entries := []*Entry{
		{ID: "L002", Type: "lesson", Topic: "api", Severity: "normal", Oneliner: "L2", Created: time.Now(), Body: "a"},
		{ID: "D001", Type: "decision", Topic: "api", Oneliner: "D1", Created: time.Now(), Body: "b"},
		{ID: "L001", Type: "lesson", Topic: "api", Severity: "normal", Oneliner: "L1", Created: time.Now(), Body: "c"},
	}
	for _, entry := range entries {
		if err := df.Append(entry); err != nil {
			t.Fatalf("append %s: %v", entry.ID, err)
		}
	}
	if err := df.Compact(); err != nil {
		t.Fatalf("compact: %v", err)
	}

	raw, err := df.DumpKnowledge()
	if err != nil {
		t.Fatalf("dump knowledge: %v", err)
	}

	idPattern := regexp.MustCompile(`(?m)^id: (.+)$`)
	matches := idPattern.FindAllStringSubmatch(raw, -1)
	var ids []string
	for _, m := range matches {
		ids = append(ids, m[1])
	}

	want := []string{"D001", "L001", "L002"}
	if len(ids) != len(want) {
		t.Fatalf("expected %d ids, got %d (%v)", len(want), len(ids), ids)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("unexpected order: got %v, want %v", ids, want)
		}
	}
}

func TestDeleteThenReuseIDSurvivesReopen(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	df, err := Create(root, "test", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	now := time.Now()
	if err := df.Append(&Entry{
		ID:       "L001",
		Type:     "lesson",
		Topic:    "api",
		Severity: "normal",
		Oneliner: "First",
		Created:  now,
		Body:     "body-1",
	}); err != nil {
		t.Fatalf("append L001: %v", err)
	}
	if err := df.Append(&Entry{
		ID:       "L002",
		Type:     "lesson",
		Topic:    "api",
		Severity: "normal",
		Oneliner: "Second",
		Created:  now,
		Body:     "body-2",
	}); err != nil {
		t.Fatalf("append first L002: %v", err)
	}

	if err := df.Delete("L002"); err != nil {
		t.Fatalf("delete L002: %v", err)
	}

	if err := df.Append(&Entry{
		ID:       "L002",
		Type:     "lesson",
		Topic:    "api",
		Severity: "high",
		Oneliner: "Reused",
		Created:  now.Add(time.Second),
		Body:     "body-reused",
	}); err != nil {
		t.Fatalf("append reused L002: %v", err)
	}
	if err := df.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	df2, err := Open(root)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer df2.Close()

	got, err := df2.Get("L002")
	if err != nil {
		t.Fatalf("get reused L002 after reopen: %v", err)
	}
	if got.Oneliner != "Reused" || got.Severity != "high" || got.Body != "body-reused" {
		t.Fatalf("unexpected reused entry: %+v", got)
	}
	if len(df2.Index.Deleted) != 0 {
		t.Fatalf("expected deleted list to be cleared for reused ID, got %v", df2.Index.Deleted)
	}
}
