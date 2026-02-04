package idgen

import (
	"strings"
	"testing"
	"time"
)

func TestNewTypedProducesValidV2ID(t *testing.T) {
	id, err := NewTyped(PrefixLesson)
	if err != nil {
		t.Fatalf("NewTyped failed: %v", err)
	}

	if !IsValidV2ID(id) {
		t.Fatalf("expected valid v2 ID, got %q", id)
	}
	if !IsValidItemID(id) {
		t.Fatalf("expected valid item ID, got %q", id)
	}
	if !strings.HasPrefix(id, "L-") {
		t.Fatalf("expected L- prefix, got %q", id)
	}
}

func TestNewTypedAtSortsByTimestamp(t *testing.T) {
	t1 := time.UnixMilli(1_700_000_000_000).UTC()
	t2 := t1.Add(time.Millisecond)

	id1, err := NewTypedAt(PrefixDecision, t1)
	if err != nil {
		t.Fatalf("NewTypedAt(t1) failed: %v", err)
	}
	id2, err := NewTypedAt(PrefixDecision, t2)
	if err != nil {
		t.Fatalf("NewTypedAt(t2) failed: %v", err)
	}

	if !(id1 < id2) {
		t.Fatalf("expected lexical ordering by time, got %q then %q", id1, id2)
	}
}

func TestNewTypedRejectsInvalidPrefix(t *testing.T) {
	if _, err := NewTyped("X"); err == nil {
		t.Fatal("expected error for invalid prefix")
	}
	if _, err := NewTyped("LL"); err == nil {
		t.Fatal("expected error for multi-char prefix")
	}
}

func TestPrefixForType(t *testing.T) {
	tests := map[string]string{
		"lesson":     PrefixLesson,
		"decision":   PrefixDecision,
		"convention": PrefixConvention,
		"incident":   PrefixCustom,
	}
	for itemType, want := range tests {
		got, err := PrefixForType(itemType)
		if err != nil {
			t.Fatalf("PrefixForType(%q) failed: %v", itemType, err)
		}
		if got != want {
			t.Fatalf("PrefixForType(%q): got %q want %q", itemType, got, want)
		}
	}
}

func TestIsValidItemIDRejectsEventID(t *testing.T) {
	eventID, err := NewTyped(PrefixEvent)
	if err != nil {
		t.Fatalf("NewTyped event failed: %v", err)
	}
	if !IsValidV2ID(eventID) {
		t.Fatalf("expected valid v2 event ID, got %q", eventID)
	}
	if IsValidItemID(eventID) {
		t.Fatalf("expected event ID to be rejected as item ID: %q", eventID)
	}
}
