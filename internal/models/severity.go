package models

// Severity levels for lessons.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityNormal   Severity = "normal"
	SeverityLow      Severity = "low"
)
