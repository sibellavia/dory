package doryfile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestGetPreservesBodyLinesThatLookLikeDelimiterOnFullReplay(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	df, err := Create(root, "test", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	entry1 := &Entry{
		ID:       "L-01KGTEST000000000000000001",
		Type:     "lesson",
		Topic:    "api",
		Severity: "normal",
		Oneliner: "Delimiter body line",
		Created:  time.Now(),
		Body:     "# Title\n\nLine before\n---\nLine after\n",
	}
	entry2 := &Entry{
		ID:       "L-01KGTEST000000000000000002",
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

	indexPath := filepath.Join(root, IndexFile)
	indexRaw, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	var idx Index
	if err := yaml.Unmarshal(indexRaw, &idx); err != nil {
		t.Fatalf("unmarshal index: %v", err)
	}
	idx.LogOffset = 0 // Force full replay from file start.
	indexRaw, err = yaml.Marshal(&idx)
	if err != nil {
		t.Fatalf("marshal index: %v", err)
	}
	if err := os.WriteFile(indexPath, indexRaw, 0644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	df2, err := Open(root)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer df2.Close()

	got, err := df2.Get("L-01KGTEST000000000000000001")
	if err != nil {
		t.Fatalf("get entry: %v", err)
	}
	if !strings.Contains(got.Body, "---\nLine after") {
		t.Fatalf("expected full body to be preserved, got:\n%s", got.Body)
	}
}

func TestOpenFailsOnCorruptEvent(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	df, err := Create(root, "test", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := df.Append(&Entry{
		ID:       "L-01KGTEST000000000000000003",
		Type:     "lesson",
		Topic:    "api",
		Severity: "normal",
		Oneliner: "seed",
		Created:  time.Now(),
		Body:     "seed",
	}); err != nil {
		t.Fatalf("append seed: %v", err)
	}
	if err := df.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	knowledgePath := filepath.Join(root, KnowledgeFile)
	f, err := os.OpenFile(knowledgePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("open knowledge: %v", err)
	}
	if _, err := f.WriteString("---\nop: unknown\n"); err != nil {
		f.Close()
		t.Fatalf("append corrupt event: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close knowledge: %v", err)
	}

	_, err = Open(root)
	if err == nil {
		t.Fatal("expected open to fail for corrupt event")
	}
	var cErr *CorruptionError
	if !errors.As(err, &cErr) {
		t.Fatalf("expected CorruptionError, got: %v", err)
	}
}

func TestOpenRejectsUnsupportedHeader(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	df, err := Create(root, "test", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := df.Append(&Entry{
		ID:       "L001",
		Type:     "lesson",
		Topic:    "api",
		Severity: "normal",
		Oneliner: "Header validation",
		Created:  time.Now(),
		Body:     "ok",
	}); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := df.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	knowledgePath := filepath.Join(root, KnowledgeFile)
	raw, err := os.ReadFile(knowledgePath)
	if err != nil {
		t.Fatalf("read knowledge: %v", err)
	}
	parts := strings.SplitN(string(raw), "\n", 2)
	if len(parts) != 2 {
		t.Fatalf("unexpected knowledge format")
	}
	parts[0] = "UNSUPPORTED:v0"
	if err := os.WriteFile(knowledgePath, []byte(parts[0]+"\n"+parts[1]), 0644); err != nil {
		t.Fatalf("write unsupported header: %v", err)
	}

	df2, err := Open(root)
	if err == nil {
		df2.Close()
		t.Fatal("expected open to fail for unsupported header")
	}
	if !strings.Contains(err.Error(), "invalid dory file header") {
		t.Fatalf("expected invalid header error, got: %v", err)
	}
}

func TestOpenRejectsUnsupportedIndexFormat(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	df, err := Create(root, "test", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := df.Append(&Entry{
		ID:       "L001",
		Type:     "lesson",
		Topic:    "api",
		Severity: "normal",
		Oneliner: "seed",
		Created:  time.Now(),
		Body:     "seed",
	}); err != nil {
		t.Fatalf("append seed: %v", err)
	}
	if err := df.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	indexPath := filepath.Join(root, IndexFile)
	indexRaw, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	var idx Index
	if err := yaml.Unmarshal(indexRaw, &idx); err != nil {
		t.Fatalf("unmarshal index: %v", err)
	}
	idx.Format = "invalid-format"
	indexRaw, err = yaml.Marshal(&idx)
	if err != nil {
		t.Fatalf("marshal index: %v", err)
	}
	if err := os.WriteFile(indexPath, indexRaw, 0644); err != nil {
		t.Fatalf("write unsupported index: %v", err)
	}

	df2, err := Open(root)
	if err == nil {
		df2.Close()
		t.Fatal("expected open to fail for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Fatalf("expected unsupported format error, got: %v", err)
	}
}

func TestWritesMinimalEventEnvelope(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	df, err := Create(root, "test", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer df.Close()

	if err := df.Append(&Entry{
		ID:       "L001",
		Type:     "lesson",
		Topic:    "api",
		Severity: "normal",
		Oneliner: "Minimal envelope",
		Created:  time.Now(),
		Body:     "body",
	}); err != nil {
		t.Fatalf("append: %v", err)
	}

	raw, err := df.DumpKnowledge()
	if err != nil {
		t.Fatalf("dump knowledge: %v", err)
	}

	if !strings.Contains(raw, "op: item.create") {
		t.Fatalf("expected item.create event, got:\n%s", raw)
	}
	if strings.Contains(raw, "\nevent_id:") || strings.Contains(raw, "\nts:") {
		t.Fatalf("expected minimal envelope without event_id/ts, got:\n%s", raw)
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

	first := strings.Index(raw, "id: D001")
	second := strings.Index(raw, "id: L001")
	third := strings.Index(raw, "id: L002")
	if first == -1 || second == -1 || third == -1 {
		t.Fatalf("expected compacted output to include D001, L001, L002; got:\n%s", raw)
	}
	if !(first < second && second < third) {
		t.Fatalf("expected deterministic ID order D001, L001, L002 in compacted output")
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
