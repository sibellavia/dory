package commands

import (
	"testing"

	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

func TestResolveListSort(t *testing.T) {
	originalAgentMode := agentMode
	t.Cleanup(func() {
		agentMode = originalAgentMode
	})

	cmd := &cobra.Command{}
	cmd.Flags().String("sort", "id", "")

	sortKey, _ := cmd.Flags().GetString("sort")

	agentMode = false
	if got := resolveListSort(cmd, sortKey); got != "id" {
		t.Fatalf("expected id sort in normal mode, got %q", got)
	}

	agentMode = true
	if got := resolveListSort(cmd, sortKey); got != "created" {
		t.Fatalf("expected created sort in agent mode default, got %q", got)
	}

	if err := cmd.Flags().Set("sort", "id"); err != nil {
		t.Fatalf("unexpected flag set error: %v", err)
	}
	sortKey, _ = cmd.Flags().GetString("sort")
	if got := resolveListSort(cmd, sortKey); got != "id" {
		t.Fatalf("expected explicit sort to win in agent mode, got %q", got)
	}
}

func TestSortListItemsByCreated(t *testing.T) {
	items := []store.ListItem{
		{ID: "D-02", CreatedAt: "2026-02-03T10:00:00Z"},
		{ID: "L-01", CreatedAt: "2026-02-03T09:00:00Z"},
		{ID: "P-03", CreatedAt: "2026-02-03T10:00:00Z"},
	}

	sortListItems(items, "created", false)
	if items[0].ID != "L-01" || items[1].ID != "D-02" || items[2].ID != "P-03" {
		t.Fatalf("unexpected ascending created order: %#v", items)
	}

	sortListItems(items, "created", true)
	if items[0].ID != "P-03" || items[1].ID != "D-02" || items[2].ID != "L-01" {
		t.Fatalf("unexpected descending created order: %#v", items)
	}
}
