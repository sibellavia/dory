package commands

import (
	"strings"
	"testing"
)

func TestParseFrontmatterInvalidYAML(t *testing.T) {
	content := strings.Join([]string{
		"---",
		"type: lesson",
		"topic: [bad",
		"---",
		"# Title",
	}, "\n")

	_, _, err := parseFrontmatter(content)
	if err == nil {
		t.Fatal("expected invalid frontmatter to return error")
	}
}

func TestParseFrontmatterValidYAML(t *testing.T) {
	content := strings.Join([]string{
		"---",
		"type: lesson",
		"topic: api",
		"severity: high",
		"---",
		"# Title",
		"",
		"Body",
	}, "\n")

	fm, body, err := parseFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm["type"] != "lesson" {
		t.Fatalf("expected type=lesson, got %#v", fm["type"])
	}
	if fm["topic"] != "api" {
		t.Fatalf("expected topic=api, got %#v", fm["topic"])
	}
	if !strings.HasPrefix(body, "# Title") {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestSplitNumberedItems(t *testing.T) {
	content := strings.Join([]string{
		"1) First item",
		"first body",
		"",
		"2. Second item",
		"second body line 1",
		"second body line 2",
	}, "\n")

	items, err := splitNumberedItems(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].title != "First item" {
		t.Fatalf("unexpected first title: %q", items[0].title)
	}
	if items[1].title != "Second item" {
		t.Fatalf("unexpected second title: %q", items[1].title)
	}
}

func TestImportItemValidationErrors(t *testing.T) {
	if _, err := importItem(nil, "lesson", "a", "b", "", "", "", nil); err == nil {
		t.Fatal("expected lesson without topic to fail")
	}
	if _, err := importItem(nil, "decision", "a", "b", "", "", "", nil); err == nil {
		t.Fatal("expected decision without topic to fail")
	}
	if _, err := importItem(nil, "pattern", "a", "b", "", "", "", nil); err == nil {
		t.Fatal("expected pattern without domain/topic to fail")
	}
	if _, err := importItem(nil, "unknown", "a", "b", "", "", "", nil); err == nil {
		t.Fatal("expected unknown type to fail")
	}
}
