package commands

import (
	"testing"
	"time"

	"github.com/sibellavia/dory/internal/models"
)

func TestValidateSeverityFlag(t *testing.T) {
	valid := []models.Severity{
		models.SeverityCritical,
		models.SeverityHigh,
		models.SeverityNormal,
		models.SeverityLow,
	}
	for _, severity := range valid {
		if err := validateSeverityFlag(severity); err != nil {
			t.Fatalf("expected %q to be valid, got error: %v", severity, err)
		}
	}

	if err := validateSeverityFlag("banana"); err == nil {
		t.Fatal("expected invalid severity to error")
	}
}

func TestValidateItemType(t *testing.T) {
	for _, itemType := range []string{"", "lesson", "decision", "pattern", "incident", "postmortem_v2"} {
		if err := validateItemType(itemType); err != nil {
			t.Fatalf("expected type %q to be valid, got error: %v", itemType, err)
		}
	}
	for _, itemType := range []string{"UPPERCASE", "has space", "contains/slash"} {
		if err := validateItemType(itemType); err == nil {
			t.Fatalf("expected invalid type %q to error", itemType)
		}
	}
}

func TestIsCoreItemType(t *testing.T) {
	for _, itemType := range []string{"lesson", "decision", "pattern"} {
		if !isCoreItemType(itemType) {
			t.Fatalf("expected %q to be core type", itemType)
		}
	}
	if isCoreItemType("incident") {
		t.Fatal("expected custom type not to be core")
	}
}

func TestParseDateFlag(t *testing.T) {
	got, err := parseDateFlag("2026-02-02", "--since")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if got.Year() != 2026 || got.Month() != time.February || got.Day() != 2 {
		t.Fatalf("unexpected parsed date: %v", got)
	}

	if _, err := parseDateFlag("bad-date", "--since"); err == nil {
		t.Fatal("expected invalid date to error")
	}
}
