package commands

import (
	"fmt"
	"regexp"
	"time"

	"github.com/sibellavia/dory/internal/models"
)

var customTypePattern = regexp.MustCompile(`^[a-z][a-z0-9._-]*$`)

func isValidSeverity(severity models.Severity) bool {
	switch severity {
	case models.SeverityCritical, models.SeverityHigh, models.SeverityNormal, models.SeverityLow:
		return true
	default:
		return false
	}
}

func validateSeverityFlag(severity models.Severity) error {
	if severity == "" {
		return nil
	}
	if isValidSeverity(severity) {
		return nil
	}
	return fmt.Errorf("invalid severity %q (use critical, high, normal, or low)", severity)
}

func validateItemType(itemType string) error {
	if itemType == "" {
		return nil
	}
	if !customTypePattern.MatchString(itemType) {
		return fmt.Errorf("invalid type %q (use lowercase names like lesson, decision, pattern, custom_type)", itemType)
	}
	return nil
}

func isCoreItemType(itemType string) bool {
	switch itemType {
	case "lesson", "decision", "pattern":
		return true
	default:
		return false
	}
}

func parseDateFlag(value, flagName string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	t, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s date %q (expected YYYY-MM-DD)", flagName, value)
	}
	return t, nil
}
