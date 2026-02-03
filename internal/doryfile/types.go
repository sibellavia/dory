package doryfile

import (
	"fmt"
	"os"
	"time"
)

const (
	MagicHeader = "DORYFILE:v1"

	EventDelim    = "---"
	KnowledgeFile = "knowledge.dory"
	IndexFile     = "index.yaml"
	IndexFormat   = "doryfile-v1"
)

const (
	opItemCreate = "item.create"
	opItemUpdate = "item.update"
	opItemDelete = "item.delete"
	opState      = "state.update"
	opCompact    = "compact"
)

// Entry represents a single knowledge item.
type Entry struct {
	ID       string    `yaml:"id"`
	Type     string    `yaml:"type"`
	Topic    string    `yaml:"topic,omitempty"`
	Domain   string    `yaml:"domain,omitempty"`
	Severity string    `yaml:"severity,omitempty"`
	Oneliner string    `yaml:"oneliner"`
	Created  time.Time `yaml:"created"`
	Refs     []string  `yaml:"refs,omitempty"`
	Body     string    `yaml:"body,omitempty"`
}

// State represents session state.
type State struct {
	Goal          string   `yaml:"goal,omitempty"`
	Progress      string   `yaml:"progress,omitempty"`
	Blocker       string   `yaml:"blocker,omitempty"`
	Next          []string `yaml:"next,omitempty"`
	WorkingFiles  []string `yaml:"working_files,omitempty"`
	OpenQuestions []string `yaml:"open_questions,omitempty"`
	LastUpdated   string   `yaml:"last_updated,omitempty"`
}

// SnapshotHead stores current-head metadata for snapshots.
type SnapshotHead struct {
	Type         string    `yaml:"type"`
	Topic        string    `yaml:"topic,omitempty"`
	Domain       string    `yaml:"domain,omitempty"`
	Severity     string    `yaml:"severity,omitempty"`
	Oneliner     string    `yaml:"oneliner"`
	Created      time.Time `yaml:"created"`
	Refs         []string  `yaml:"refs,omitempty"`
	BodyOffset   int64     `yaml:"body_offset"`
	BodyLen      int       `yaml:"body_len"`
	LastEventSeq uint64    `yaml:"last_event_seq"`
}

// Index holds project metadata and cached state.
type Index struct {
	Version     int                      `yaml:"version,omitempty"`
	Format      string                   `yaml:"format,omitempty"`
	Project     string                   `yaml:"project"`
	Description string                   `yaml:"description,omitempty"`
	State       *State                   `yaml:"state,omitempty"`
	Deleted     []string                 `yaml:"deleted,omitempty"`
	AppliedSeq  uint64                   `yaml:"applied_seq,omitempty"`
	LogOffset   int64                    `yaml:"log_offset,omitempty"`
	Heads       map[string]*SnapshotHead `yaml:"heads,omitempty"`
}

// MemoryEntry holds offset and metadata for fast lookup (in-memory only).
type MemoryEntry struct {
	Offset   int64
	BodyLen  int
	Type     string
	Topic    string
	Domain   string
	Severity string
	Oneliner string
	Created  time.Time
	Refs     []string
}

// DoryFile represents the dory storage.
type DoryFile struct {
	Dir           string
	KnowledgePath string
	IndexPath     string
	Index         *Index
	knowledge     *os.File

	nextSeq   uint64
	logOffset int64

	// In-memory index (computed on open).
	entries map[string]*MemoryEntry
}

type logEvent struct {
	Op string `yaml:"op"`

	Item  *Entry `yaml:"item,omitempty"`
	ID    string `yaml:"id,omitempty"`
	State *State `yaml:"state,omitempty"`
}

// CorruptionError indicates malformed knowledge log content.
type CorruptionError struct {
	Offset int64
	Reason string
}

func (e *CorruptionError) Error() string {
	return fmt.Sprintf("corrupt doryfile at offset %d: %s", e.Offset, e.Reason)
}

func corruptionError(offset int64, format string, args ...interface{}) error {
	return &CorruptionError{Offset: offset, Reason: fmt.Sprintf(format, args...)}
}
