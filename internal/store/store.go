package store

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/sibellavia/dory/internal/doryfile"
	"github.com/sibellavia/dory/internal/models"
	"gopkg.in/yaml.v3"
)

const (
	// DoryDir is the default root directory for local project knowledge.
	DoryDir = ".dory"
)

// Store manages dory knowledge in a single-file format
type Store struct {
	Root string
	df   *doryfile.DoryFile
}

// New creates a new Store instance.
func New(root string) *Store {
	if root == "" {
		root = DoryDir
	}
	return &Store{
		Root: root,
	}
}

// Exists checks if the dory store exists
func (s *Store) Exists() bool {
	df, err := doryfile.Open(s.Root)
	if err != nil {
		return false
	}
	df.Close()
	return true
}

// Init initializes the .dory directory with single file format
func (s *Store) Init(project, description string) error {
	if s.Exists() {
		return fmt.Errorf("dory already initialized in %s", s.Root)
	}

	// Create directory
	if err := ensureDir(s.Root); err != nil {
		return fmt.Errorf("failed to create root directory: %w", err)
	}

	// Create the dory storage (content.dory + index.yaml)
	df, err := doryfile.Create(s.Root, project, description)
	if err != nil {
		return fmt.Errorf("failed to create dory storage: %w", err)
	}
	defer df.Close()

	return nil
}

// open opens the dory storage if not already open
func (s *Store) open() error {
	if s.df != nil {
		return nil
	}
	df, err := doryfile.Open(s.Root)
	if err != nil {
		return err
	}
	s.df = df
	return nil
}

// Close closes the dory file
func (s *Store) Close() error {
	if s.df != nil {
		err := s.df.Close()
		s.df = nil
		return err
	}
	return nil
}

// Learn adds a new lesson
func (s *Store) Learn(oneliner, topic string, severity models.Severity, body string, refs []string) (string, error) {
	if err := s.open(); err != nil {
		return "", err
	}

	id := s.df.NextID("L")
	created := time.Now()

	fullBody := body
	if fullBody == "" {
		fullBody = fmt.Sprintf("# %s\n\n## Details\n\n(Add details here)\n", oneliner)
	}

	entry := &doryfile.Entry{
		ID:       id,
		Type:     "lesson",
		Topic:    topic,
		Severity: string(severity),
		Oneliner: oneliner,
		Created:  created,
		Refs:     refs,
		Body:     fullBody,
	}

	if err := s.df.Append(entry); err != nil {
		return "", fmt.Errorf("failed to append lesson: %w", err)
	}

	return id, nil
}

// Decide adds a new decision
func (s *Store) Decide(oneliner, topic, rationale, body string, refs []string) (string, error) {
	if err := s.open(); err != nil {
		return "", err
	}

	id := s.df.NextID("D")
	created := time.Now()

	fullBody := body
	if fullBody == "" {
		fullBody = fmt.Sprintf("# %s\n\n## Context\n\n(Add context here)\n\n## Decision\n\n%s\n\n## Rationale\n\n%s\n", oneliner, oneliner, rationale)
	}

	entry := &doryfile.Entry{
		ID:       id,
		Type:     "decision",
		Topic:    topic,
		Oneliner: oneliner,
		Created:  created,
		Refs:     refs,
		Body:     fullBody,
	}

	if err := s.df.Append(entry); err != nil {
		return "", fmt.Errorf("failed to append decision: %w", err)
	}

	return id, nil
}

// Pattern adds a new pattern
func (s *Store) Pattern(oneliner, domain, body string, refs []string) (string, error) {
	if err := s.open(); err != nil {
		return "", err
	}

	id := s.df.NextID("P")
	created := time.Now()

	fullBody := body
	if fullBody == "" {
		fullBody = fmt.Sprintf("# %s\n\n## Pattern\n\n%s\n\n## Implementation\n\n(Add implementation details here)\n", oneliner, oneliner)
	}

	entry := &doryfile.Entry{
		ID:       id,
		Type:     "pattern",
		Domain:   domain,
		Oneliner: oneliner,
		Created:  created,
		Refs:     refs,
		Body:     fullBody,
	}

	if err := s.df.Append(entry); err != nil {
		return "", fmt.Errorf("failed to append pattern: %w", err)
	}

	return id, nil
}

// CreateCustom adds a new custom type entry.
func (s *Store) CreateCustom(itemType, oneliner, topic, body string, refs []string) (string, error) {
	if err := s.open(); err != nil {
		return "", err
	}

	id := s.df.NextID("K")
	created := time.Now()

	fullBody := body
	if fullBody == "" {
		fullBody = fmt.Sprintf("# %s\n\n## Context\n\n(Add context here)\n", oneliner)
	}

	entry := &doryfile.Entry{
		ID:       id,
		Type:     itemType,
		Topic:    topic,
		Oneliner: oneliner,
		Created:  created,
		Refs:     refs,
		Body:     fullBody,
	}

	if err := s.df.Append(entry); err != nil {
		return "", fmt.Errorf("failed to append %s: %w", itemType, err)
	}

	return id, nil
}

// Show returns the full content for a specific item
func (s *Store) Show(id string) (string, error) {
	if err := s.open(); err != nil {
		return "", err
	}

	entry, err := s.df.Get(id)
	if err != nil {
		return "", err
	}

	// Format as YAML frontmatter + body for human-readable output.
	frontmatter := map[string]interface{}{
		"id":       entry.ID,
		"type":     entry.Type,
		"oneliner": entry.Oneliner,
		"created":  entry.Created.Format(time.RFC3339),
	}
	if entry.Topic != "" {
		frontmatter["topic"] = entry.Topic
	}
	if entry.Domain != "" {
		frontmatter["domain"] = entry.Domain
	}
	if entry.Severity != "" {
		frontmatter["severity"] = entry.Severity
	}
	if len(entry.Refs) > 0 {
		frontmatter["refs"] = entry.Refs
	}

	yamlData, err := yaml.Marshal(frontmatter)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlData)
	buf.WriteString("---\n\n")
	buf.WriteString(entry.Body)

	return buf.String(), nil
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

// List returns items matching the filters
func (s *Store) List(topic, itemType string, severity models.Severity, since, until time.Time) ([]ListItem, error) {
	if err := s.open(); err != nil {
		return nil, err
	}

	var items []ListItem

	for id, entry := range s.df.Entries() {
		if itemType != "" && entry.Type != itemType {
			continue
		}
		if topic != "" && entry.Topic != topic && entry.Domain != topic {
			continue
		}
		if severity != "" && models.Severity(entry.Severity) != severity {
			continue
		}
		if !since.IsZero() && entry.Created.Before(since) {
			continue
		}
		if !until.IsZero() && entry.Created.After(until) {
			continue
		}

		item := ListItem{
			ID:       id,
			Type:     entry.Type,
			Oneliner: entry.Oneliner,
			Created:  entry.Created.Format("2006-01-02"),
		}
		if entry.Topic != "" {
			item.Topic = entry.Topic
		}
		if entry.Domain != "" {
			item.Domain = entry.Domain
		}
		if entry.Severity != "" {
			item.Severity = models.Severity(entry.Severity)
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return items, nil
}

// Remove deletes an item by ID
func (s *Store) Remove(id string) error {
	if err := s.open(); err != nil {
		return err
	}
	return s.df.Delete(id)
}

// UpdateStatus updates the session state
func (s *Store) UpdateStatus(goal, progress, blocker string, next, workingFiles, openQuestions []string) error {
	if err := s.open(); err != nil {
		return err
	}

	state := s.df.Index.State
	if state == nil {
		state = &doryfile.State{}
	}

	if goal != "" {
		state.Goal = goal
	}
	if progress != "" {
		state.Progress = progress
	}
	if blocker != "" {
		state.Blocker = blocker
	}
	if len(next) > 0 {
		state.Next = next
	}
	if len(workingFiles) > 0 {
		state.WorkingFiles = workingFiles
	}
	if len(openQuestions) > 0 {
		state.OpenQuestions = openQuestions
	}
	state.LastUpdated = time.Now().Format(time.RFC3339)

	return s.df.UpdateState(state)
}

// Recall returns all knowledge for a topic with summaries
func (s *Store) Recall(topic string) (string, error) {
	if err := s.open(); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("# Knowledge for topic: %s\n\n", topic))

	grouped := make(map[string][]string)
	entries := s.df.Entries()
	for id, entry := range entries {
		if entry.Topic == topic || entry.Domain == topic {
			grouped[entry.Type] = append(grouped[entry.Type], id)
		}
	}

	var types []string
	for itemType := range grouped {
		types = append(types, itemType)
	}
	sort.Strings(types)

	for _, itemType := range types {
		ids := grouped[itemType]
		sort.Strings(ids)
		buf.WriteString(fmt.Sprintf("## %s\n\n", itemType))
		for _, id := range ids {
			entry := entries[id]
			if entry.Severity != "" {
				buf.WriteString(fmt.Sprintf("### %s [%s]\n", id, entry.Severity))
			} else {
				buf.WriteString(fmt.Sprintf("### %s\n", id))
			}
			buf.WriteString(fmt.Sprintf("%s\n\n", entry.Oneliner))
		}
	}

	if len(types) == 0 {
		buf.WriteString("No knowledge found for this topic.\n")
	}

	return buf.String(), nil
}

// TopicInfo represents a topic with its item count.
type TopicInfo struct {
	Name  string `json:"name" yaml:"name"`
	Count int    `json:"count" yaml:"count"`
}

// Topics returns all topics with their item counts
func (s *Store) Topics() ([]TopicInfo, error) {
	if err := s.open(); err != nil {
		return nil, err
	}

	counts := make(map[string]int)

	for _, entry := range s.df.Entries() {
		if entry.Topic != "" {
			counts[entry.Topic]++
		} else if entry.Domain != "" {
			counts[entry.Domain]++
		}
	}

	var topics []TopicInfo
	for name, count := range counts {
		topics = append(topics, TopicInfo{Name: name, Count: count})
	}

	sort.Slice(topics, func(i, j int) bool {
		return topics[i].Name < topics[j].Name
	})

	return topics, nil
}

// Compact removes deleted entries and rebuilds the file
func (s *Store) Compact() error {
	if err := s.open(); err != nil {
		return err
	}
	return s.df.Compact()
}

// DumpContent returns the raw content file (for debugging/viewing)
func (s *Store) DumpContent() (string, error) {
	if err := s.open(); err != nil {
		return "", err
	}
	return s.df.DumpKnowledge()
}

// DumpIndex returns the index file content
func (s *Store) DumpIndex() (string, error) {
	if err := s.open(); err != nil {
		return "", err
	}
	return s.df.DumpIndex()
}

// helper to ensure directory exists
func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
