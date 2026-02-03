package store

import (
	"time"

	"github.com/sibellavia/dory/internal/doryfile"
	"github.com/sibellavia/dory/internal/models"
)

const (
	// DoryDir is the default root directory for local project knowledge.
	DoryDir = ".dory"

	writeLockFile     = "write.lock"
	writeLockTimeout  = 10 * time.Second
	writeLockRetry    = 25 * time.Millisecond
	writeLockStaleAge = 30 * time.Minute
)

// Store manages dory knowledge in a single-file format.
type Store struct {
	Root string
	df   *doryfile.DoryFile
}

// ListItem represents an item in list output.
type ListItem struct {
	ID       string          `json:"id" yaml:"id"`
	Type     string          `json:"type" yaml:"type"`
	Oneliner string          `json:"oneliner" yaml:"oneliner"`
	Topic    string          `json:"topic,omitempty" yaml:"topic,omitempty"`
	Domain   string          `json:"domain,omitempty" yaml:"domain,omitempty"`
	Severity models.Severity `json:"severity,omitempty" yaml:"severity,omitempty"`
	Created  string          `json:"created" yaml:"created"`
}

// TopicInfo represents a topic with its item count.
type TopicInfo struct {
	Name  string `json:"name" yaml:"name"`
	Count int    `json:"count" yaml:"count"`
}

// RefInfo represents relationship information for an item.
type RefInfo struct {
	ID           string    `json:"id" yaml:"id"`
	Type         string    `json:"type" yaml:"type"`
	Oneliner     string    `json:"oneliner" yaml:"oneliner"`
	RefsTo       []RefItem `json:"refs_to,omitempty" yaml:"refs_to,omitempty"`
	ReferencedBy []RefItem `json:"referenced_by,omitempty" yaml:"referenced_by,omitempty"`
}

// RefItem represents a referenced item with its metadata.
type RefItem struct {
	ID       string `json:"id" yaml:"id"`
	Type     string `json:"type" yaml:"type"`
	Oneliner string `json:"oneliner" yaml:"oneliner"`
}

// ExpandedItem represents an item with its full content in expand output.
type ExpandedItem struct {
	ID       string   `json:"id" yaml:"id"`
	Type     string   `json:"type" yaml:"type"`
	Oneliner string   `json:"oneliner" yaml:"oneliner"`
	Topic    string   `json:"topic,omitempty" yaml:"topic,omitempty"`
	Domain   string   `json:"domain,omitempty" yaml:"domain,omitempty"`
	Refs     []string `json:"refs,omitempty" yaml:"refs,omitempty"`
	Body     string   `json:"body" yaml:"body"`
}

// ExpandResult contains the root item and all connected items.
type ExpandResult struct {
	Root      ExpandedItem   `json:"root" yaml:"root"`
	Connected []ExpandedItem `json:"connected,omitempty" yaml:"connected,omitempty"`
}

// ContextResult contains smart context for agent session start.
type ContextResult struct {
	Project  string        `json:"project" yaml:"project"`
	State    *ContextState `json:"state,omitempty" yaml:"state,omitempty"`
	Critical []ListItem    `json:"critical,omitempty" yaml:"critical,omitempty"`
	Recent   []ListItem    `json:"recent,omitempty" yaml:"recent,omitempty"`
	Topic    []ListItem    `json:"topic,omitempty" yaml:"topic,omitempty"`
}

// ContextState is session state for context output.
type ContextState struct {
	Goal        string   `json:"goal,omitempty" yaml:"goal,omitempty"`
	Progress    string   `json:"progress,omitempty" yaml:"progress,omitempty"`
	Blocker     string   `json:"blocker,omitempty" yaml:"blocker,omitempty"`
	Next        []string `json:"next,omitempty" yaml:"next,omitempty"`
	LastUpdated string   `json:"last_updated,omitempty" yaml:"last_updated,omitempty"`
}
