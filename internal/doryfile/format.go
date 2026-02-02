package doryfile

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	MagicHeader   = "DORY:v1"
	ItemDelim     = "==="
	KnowledgeFile = "knowledge.dory"
	IndexFile     = "index.yaml"
)

// Entry represents a single knowledge item
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

// State represents session state
type State struct {
	Goal          string   `yaml:"goal,omitempty"`
	Progress      string   `yaml:"progress,omitempty"`
	Blocker       string   `yaml:"blocker,omitempty"`
	Next          []string `yaml:"next,omitempty"`
	WorkingFiles  []string `yaml:"working_files,omitempty"`
	OpenQuestions []string `yaml:"open_questions,omitempty"`
	LastUpdated   string   `yaml:"last_updated,omitempty"`
}

// Index holds project metadata and state (no offsets!)
type Index struct {
	Project string   `yaml:"project"`
	State   *State   `yaml:"state,omitempty"`
	Deleted []string `yaml:"deleted,omitempty"` // IDs of deleted entries (until compact)
}

// MemoryEntry holds offset and metadata for fast lookup (in-memory only)
type MemoryEntry struct {
	Offset   int64
	Type     string
	Topic    string
	Domain   string
	Severity string
	Oneliner string
	Created  time.Time
	Refs     []string
}

// DoryFile represents the dory storage
type DoryFile struct {
	Dir           string
	KnowledgePath string
	IndexPath     string
	Index         *Index
	knowledge     *os.File

	// In-memory index (computed on open)
	entries map[string]*MemoryEntry
}

// Create creates a new dory storage
func Create(dir, project string) (*DoryFile, error) {
	knowledgePath := filepath.Join(dir, KnowledgeFile)
	indexPath := filepath.Join(dir, IndexFile)

	// Create knowledge file with magic header
	f, err := os.Create(knowledgePath)
	if err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(f, "%s\n", MagicHeader); err != nil {
		f.Close()
		return nil, err
	}

	df := &DoryFile{
		Dir:           dir,
		KnowledgePath: knowledgePath,
		IndexPath:     indexPath,
		knowledge:     f,
		Index: &Index{
			Project: project,
			State:   &State{},
		},
		entries: make(map[string]*MemoryEntry),
	}

	// Write initial index (state only)
	if err := df.saveIndex(); err != nil {
		f.Close()
		return nil, err
	}

	return df, nil
}

// Open opens an existing dory storage
func Open(dir string) (*DoryFile, error) {
	knowledgePath := filepath.Join(dir, KnowledgeFile)
	indexPath := filepath.Join(dir, IndexFile)

	// Open knowledge file
	f, err := os.OpenFile(knowledgePath, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	df := &DoryFile{
		Dir:           dir,
		KnowledgePath: knowledgePath,
		IndexPath:     indexPath,
		knowledge:     f,
		entries:       make(map[string]*MemoryEntry),
	}

	// Load index (state only)
	if err := df.loadIndex(); err != nil {
		f.Close()
		return nil, err
	}

	// Scan knowledge file, build in-memory index
	if err := df.scan(); err != nil {
		f.Close()
		return nil, err
	}

	return df, nil
}

// scan reads the knowledge file and builds the in-memory index
func (df *DoryFile) scan() error {
	df.knowledge.Seek(0, 0)
	reader := bufio.NewReader(df.knowledge)

	// Verify magic header
	header, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}
	if strings.TrimSpace(header) != MagicHeader {
		return fmt.Errorf("invalid dory file: expected %s", MagicHeader)
	}

	// Track current position (after header)
	pos := int64(len(header))

	for {
		// Read delimiter line
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if strings.TrimSpace(line) != ItemDelim {
			pos += int64(len(line))
			continue
		}

		// Record offset where entry starts (at the delimiter)
		entryOffset := pos
		pos += int64(len(line))

		// Read entry content until next delimiter or EOF
		var content strings.Builder
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				content.WriteString(line)
				pos += int64(len(line))
				break
			}
			if err != nil {
				return err
			}
			if strings.TrimSpace(line) == ItemDelim {
				// Seek back to read the delimiter again in the outer loop
				seekPos := pos // Don't increment pos, just seek to current position
				df.knowledge.Seek(seekPos, 0)
				reader = bufio.NewReader(df.knowledge)
				break
			}
			content.WriteString(line)
			pos += int64(len(line))
		}

		// Parse entry
		var entry Entry
		if err := yaml.Unmarshal([]byte(content.String()), &entry); err != nil {
			continue // Skip malformed entries
		}

		// Skip deleted entries
		isDeleted := false
		for _, delID := range df.Index.Deleted {
			if delID == entry.ID {
				isDeleted = true
				break
			}
		}
		if isDeleted {
			continue
		}

		// Add to in-memory index
		df.entries[entry.ID] = &MemoryEntry{
			Offset:   entryOffset,
			Type:     entry.Type,
			Topic:    entry.Topic,
			Domain:   entry.Domain,
			Severity: entry.Severity,
			Oneliner: entry.Oneliner,
			Created:  entry.Created,
			Refs:     entry.Refs,
		}
	}

	return nil
}

// loadIndex reads the index file (state only)
func (df *DoryFile) loadIndex() error {
	data, err := os.ReadFile(df.IndexPath)
	if err != nil {
		return fmt.Errorf("failed to read index: %w", err)
	}

	var index Index
	if err := yaml.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("failed to parse index: %w", err)
	}

	df.Index = &index
	if df.Index.State == nil {
		df.Index.State = &State{}
	}

	return nil
}

// saveIndex writes the index file (state only)
func (df *DoryFile) saveIndex() error {
	data, err := yaml.Marshal(df.Index)
	if err != nil {
		return err
	}
	return os.WriteFile(df.IndexPath, data, 0644)
}

// Append adds a new entry (append-only)
func (df *DoryFile) Append(entry *Entry) error {
	// Get current file size (where entry will start)
	stat, err := df.knowledge.Stat()
	if err != nil {
		return err
	}
	entryOffset := stat.Size()

	// Write entry
	var buf strings.Builder
	buf.WriteString(ItemDelim + "\n")

	entryData, err := yaml.Marshal(entry)
	if err != nil {
		return err
	}
	buf.Write(entryData)
	buf.WriteString("\n")

	if _, err := df.knowledge.WriteString(buf.String()); err != nil {
		return err
	}

	// Note: We don't Sync() on every append for performance.
	// Data is flushed on Close() or by the OS.

	// Add to in-memory index
	df.entries[entry.ID] = &MemoryEntry{
		Offset:   entryOffset,
		Type:     entry.Type,
		Topic:    entry.Topic,
		Domain:   entry.Domain,
		Severity: entry.Severity,
		Oneliner: entry.Oneliner,
		Created:  entry.Created,
		Refs:     entry.Refs,
	}

	return nil
}

// Get retrieves an entry by ID
func (df *DoryFile) Get(id string) (*Entry, error) {
	mem, ok := df.entries[id]
	if !ok {
		return nil, fmt.Errorf("item %s not found", id)
	}

	// Seek to offset
	if _, err := df.knowledge.Seek(mem.Offset, 0); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(df.knowledge)

	// Skip delimiter
	if _, err := reader.ReadString('\n'); err != nil {
		return nil, err
	}

	// Read until next delimiter or EOF
	var content strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			content.WriteString(line)
			break
		}
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(line) == ItemDelim {
			break
		}
		content.WriteString(line)
	}

	var entry Entry
	if err := yaml.Unmarshal([]byte(content.String()), &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

// Delete removes an entry (persisted in index, content stays until compact)
func (df *DoryFile) Delete(id string) error {
	if _, ok := df.entries[id]; !ok {
		return fmt.Errorf("item %s not found", id)
	}
	delete(df.entries, id)

	// Persist deletion in index
	df.Index.Deleted = append(df.Index.Deleted, id)
	return df.saveIndex()
}

// UpdateState updates the session state
func (df *DoryFile) UpdateState(state *State) error {
	df.Index.State = state
	return df.saveIndex()
}

// Compact rewrites the knowledge file, removing deleted entries
func (df *DoryFile) Compact() error {
	// Collect all current entries
	var entries []*Entry
	for id := range df.entries {
		entry, err := df.Get(id)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	df.knowledge.Close()

	// Create new knowledge file
	tmpPath := df.KnowledgePath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	// Write header
	if _, err := fmt.Fprintf(tmpFile, "%s\n", MagicHeader); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}

	// Clear in-memory index
	df.entries = make(map[string]*MemoryEntry)

	// Write all entries
	for _, entry := range entries {
		stat, _ := tmpFile.Stat()
		offset := stat.Size()

		var buf strings.Builder
		buf.WriteString(ItemDelim + "\n")
		entryData, _ := yaml.Marshal(entry)
		buf.Write(entryData)
		buf.WriteString("\n")

		if _, err := tmpFile.WriteString(buf.String()); err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
			return err
		}

		df.entries[entry.ID] = &MemoryEntry{
			Offset:   offset,
			Type:     entry.Type,
			Topic:    entry.Topic,
			Domain:   entry.Domain,
			Severity: entry.Severity,
			Oneliner: entry.Oneliner,
			Created:  entry.Created,
			Refs:     entry.Refs,
		}
	}

	tmpFile.Close()

	// Replace old file
	if err := os.Rename(tmpPath, df.KnowledgePath); err != nil {
		return err
	}

	// Reopen
	df.knowledge, err = os.OpenFile(df.KnowledgePath, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// Clear deleted list since entries are now physically removed
	df.Index.Deleted = nil
	return df.saveIndex()
}

// Close syncs and closes the knowledge file
func (df *DoryFile) Close() error {
	if df.knowledge != nil {
		df.knowledge.Sync() // Flush any buffered writes
		return df.knowledge.Close()
	}
	return nil
}

// NextID generates the next ID for a type
func (df *DoryFile) NextID(prefix string) string {
	var max int
	for id := range df.entries {
		if strings.HasPrefix(id, prefix) {
			numStr := strings.TrimPrefix(id, prefix)
			if num, err := strconv.Atoi(numStr); err == nil && num > max {
				max = num
			}
		}
	}
	return fmt.Sprintf("%s%03d", prefix, max+1)
}

// Lessons returns all lessons from in-memory index
func (df *DoryFile) Lessons() map[string]*MemoryEntry {
	result := make(map[string]*MemoryEntry)
	for id, entry := range df.entries {
		if entry.Type == "lesson" {
			result[id] = entry
		}
	}
	return result
}

// Decisions returns all decisions from in-memory index
func (df *DoryFile) Decisions() map[string]*MemoryEntry {
	result := make(map[string]*MemoryEntry)
	for id, entry := range df.entries {
		if entry.Type == "decision" {
			result[id] = entry
		}
	}
	return result
}

// Patterns returns all patterns from in-memory index
func (df *DoryFile) Patterns() map[string]*MemoryEntry {
	result := make(map[string]*MemoryEntry)
	for id, entry := range df.entries {
		if entry.Type == "pattern" {
			result[id] = entry
		}
	}
	return result
}

// DumpKnowledge returns the raw knowledge file content
func (df *DoryFile) DumpKnowledge() (string, error) {
	data, err := os.ReadFile(df.KnowledgePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DumpIndex returns the index file content
func (df *DoryFile) DumpIndex() (string, error) {
	data, err := os.ReadFile(df.IndexPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
