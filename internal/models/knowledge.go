package models

import "time"

// Severity levels for lessons
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityNormal   Severity = "normal"
	SeverityLow      Severity = "low"
)

// KnowledgeType represents the type of knowledge item
type KnowledgeType string

const (
	TypeLesson   KnowledgeType = "lesson"
	TypeDecision KnowledgeType = "decision"
	TypePattern  KnowledgeType = "pattern"
)

// Lesson represents a learned lesson (something discovered, often the hard way)
type Lesson struct {
	ID       string    `yaml:"id"`
	Type     string    `yaml:"type"`
	Topic    string    `yaml:"topic"`
	Severity Severity  `yaml:"severity"`
	Created  time.Time `yaml:"created"`
	Tags     []string  `yaml:"tags,omitempty"`
	Summary  string    `yaml:"summary,omitempty"`
}

// Decision represents an architectural/technical choice
type Decision struct {
	ID      string    `yaml:"id"`
	Type    string    `yaml:"type"`
	Topic   string    `yaml:"topic"`
	Created time.Time `yaml:"created"`
	Status  string    `yaml:"status,omitempty"` // active, superseded, etc.
	Tags    []string  `yaml:"tags,omitempty"`
	Summary string    `yaml:"summary,omitempty"`
}

// Pattern represents an established convention
type Pattern struct {
	ID      string    `yaml:"id"`
	Type    string    `yaml:"type"`
	Domain  string    `yaml:"domain"`
	Created time.Time `yaml:"created"`
	Tags    []string  `yaml:"tags,omitempty"`
	Summary string    `yaml:"summary,omitempty"`
}

// KnowledgeItem is a generic interface for all knowledge types
type KnowledgeItem interface {
	GetID() string
	GetType() KnowledgeType
	GetTopic() string
	GetCreated() time.Time
}

func (l *Lesson) GetID() string           { return l.ID }
func (l *Lesson) GetType() KnowledgeType  { return TypeLesson }
func (l *Lesson) GetTopic() string        { return l.Topic }
func (l *Lesson) GetCreated() time.Time   { return l.Created }

func (d *Decision) GetID() string           { return d.ID }
func (d *Decision) GetType() KnowledgeType  { return TypeDecision }
func (d *Decision) GetTopic() string        { return d.Topic }
func (d *Decision) GetCreated() time.Time   { return d.Created }

func (p *Pattern) GetID() string           { return p.ID }
func (p *Pattern) GetType() KnowledgeType  { return TypePattern }
func (p *Pattern) GetTopic() string        { return p.Domain }
func (p *Pattern) GetCreated() time.Time   { return p.Created }

// EngFile represents a parsed .eng file with frontmatter and body
type EngFile struct {
	Frontmatter map[string]interface{} `yaml:",inline"`
	Body        string                 `yaml:"-"`
}
