package store

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sibellavia/dory/internal/fileio"
	"github.com/sibellavia/dory/internal/models"
)

const (
	DoryDir      = ".dory"
	IndexFile    = "index.yaml"
	KnowledgeDir = "knowledge"
)

// Store manages the dory knowledge store
type Store struct {
	Root         string
	IndexPath    string
	KnowledgeDir string
}

// New creates a new Store instance
func New(root string) *Store {
	if root == "" {
		root = DoryDir
	}
	return &Store{
		Root:         root,
		IndexPath:    filepath.Join(root, IndexFile),
		KnowledgeDir: filepath.Join(root, KnowledgeDir),
	}
}

// Exists checks if the dory store exists
func (s *Store) Exists() bool {
	return fileio.FileExists(s.IndexPath)
}

// Init initializes the .dory directory structure
func (s *Store) Init(project, description string) error {
	if s.Exists() {
		return fmt.Errorf("dory already initialized in %s", s.Root)
	}

	// Create directories
	if err := fileio.EnsureDir(s.Root); err != nil {
		return fmt.Errorf("failed to create root directory: %w", err)
	}
	if err := fileio.EnsureDir(s.KnowledgeDir); err != nil {
		return fmt.Errorf("failed to create knowledge directory: %w", err)
	}

	// Create index.yaml with embedded state
	index := models.NewIndex(project)
	index.Description = description
	index.State = models.NewState()
	if err := fileio.WriteYAML(s.IndexPath, index); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// LoadIndex loads the index from disk
func (s *Store) LoadIndex() (*models.Index, error) {
	var index models.Index
	if err := fileio.ReadYAML(s.IndexPath, &index); err != nil {
		return nil, err
	}
	// Initialize maps if nil
	if index.Lessons == nil {
		index.Lessons = make(map[string]models.IndexEntry)
	}
	if index.Decisions == nil {
		index.Decisions = make(map[string]models.IndexEntry)
	}
	if index.Patterns == nil {
		index.Patterns = make(map[string]models.IndexEntry)
	}
	if index.Edges == nil {
		index.Edges = make(map[string][]string)
	}
	return &index, nil
}

// SaveIndex saves the index to disk
func (s *Store) SaveIndex(index *models.Index) error {
	return fileio.WriteYAML(s.IndexPath, index)
}

// Learn adds a new lesson
func (s *Store) Learn(oneliner, topic string, severity models.Severity, summary, body string, refs []string) (string, error) {
	id, err := fileio.NextID(s.KnowledgeDir, "L")
	if err != nil {
		return "", fmt.Errorf("failed to generate ID: %w", err)
	}

	created := time.Now()

	// Create the .eng file
	frontmatter := map[string]interface{}{
		"id":       id,
		"type":     "lesson",
		"oneliner": oneliner,
		"topic":    topic,
		"severity": string(severity),
		"created":  created.Format(time.RFC3339),
	}
	if summary != "" {
		frontmatter["summary"] = summary
	}
	if len(refs) > 0 {
		frontmatter["refs"] = refs
	}

	engBody := body
	if engBody == "" {
		engBody = fmt.Sprintf("# %s\n\n## Details\n\n(Add details here)\n", oneliner)
	}

	engPath := filepath.Join(s.KnowledgeDir, id+".eng")
	if err := fileio.WriteEngFile(engPath, frontmatter, engBody); err != nil {
		return "", fmt.Errorf("failed to write lesson file: %w", err)
	}

	// Update index
	index, err := s.LoadIndex()
	if err != nil {
		return "", fmt.Errorf("failed to load index: %w", err)
	}
	index.AddLesson(id, oneliner, topic, severity, created)
	for _, ref := range refs {
		index.AddEdge(id, ref)
	}
	if err := s.SaveIndex(index); err != nil {
		return "", fmt.Errorf("failed to save index: %w", err)
	}

	return id, nil
}

// Decide adds a new decision
func (s *Store) Decide(oneliner, topic, rationale, summary, body string, refs []string) (string, error) {
	id, err := fileio.NextID(s.KnowledgeDir, "D")
	if err != nil {
		return "", fmt.Errorf("failed to generate ID: %w", err)
	}

	created := time.Now()

	// Create the .eng file
	frontmatter := map[string]interface{}{
		"id":       id,
		"type":     "decision",
		"oneliner": oneliner,
		"topic":    topic,
		"created":  created.Format(time.RFC3339),
	}
	if summary != "" {
		frontmatter["summary"] = summary
	}
	if len(refs) > 0 {
		frontmatter["refs"] = refs
	}

	engBody := body
	if engBody == "" {
		engBody = fmt.Sprintf("# %s\n\n## Context\n\n(Add context here)\n\n## Decision\n\n%s\n\n## Rationale\n\n%s\n", oneliner, oneliner, rationale)
	}

	engPath := filepath.Join(s.KnowledgeDir, id+".eng")
	if err := fileio.WriteEngFile(engPath, frontmatter, engBody); err != nil {
		return "", fmt.Errorf("failed to write decision file: %w", err)
	}

	// Update index
	index, err := s.LoadIndex()
	if err != nil {
		return "", fmt.Errorf("failed to load index: %w", err)
	}
	index.AddDecision(id, oneliner, topic, created)
	for _, ref := range refs {
		index.AddEdge(id, ref)
	}
	if err := s.SaveIndex(index); err != nil {
		return "", fmt.Errorf("failed to save index: %w", err)
	}

	return id, nil
}

// Pattern adds a new pattern
func (s *Store) Pattern(oneliner, domain, summary, body string, refs []string) (string, error) {
	id, err := fileio.NextID(s.KnowledgeDir, "P")
	if err != nil {
		return "", fmt.Errorf("failed to generate ID: %w", err)
	}

	created := time.Now()

	// Create the .eng file
	frontmatter := map[string]interface{}{
		"id":       id,
		"type":     "pattern",
		"oneliner": oneliner,
		"domain":   domain,
		"created":  created.Format(time.RFC3339),
	}
	if summary != "" {
		frontmatter["summary"] = summary
	}
	if len(refs) > 0 {
		frontmatter["refs"] = refs
	}

	engBody := body
	if engBody == "" {
		engBody = fmt.Sprintf("# %s\n\n## Pattern\n\n%s\n\n## Implementation\n\n(Add implementation details here)\n", oneliner, oneliner)
	}

	engPath := filepath.Join(s.KnowledgeDir, id+".eng")
	if err := fileio.WriteEngFile(engPath, frontmatter, engBody); err != nil {
		return "", fmt.Errorf("failed to write pattern file: %w", err)
	}

	// Update index
	index, err := s.LoadIndex()
	if err != nil {
		return "", fmt.Errorf("failed to load index: %w", err)
	}
	index.AddPattern(id, oneliner, domain, created)
	for _, ref := range refs {
		index.AddEdge(id, ref)
	}
	if err := s.SaveIndex(index); err != nil {
		return "", fmt.Errorf("failed to save index: %w", err)
	}

	return id, nil
}

// UpdateStatus updates the session state
func (s *Store) UpdateStatus(goal, progress, blocker string, next, workingFiles, openQuestions []string) error {
	index, err := s.LoadIndex()
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	if index.State == nil {
		index.State = models.NewState()
	}

	index.State.Update(goal, progress, blocker, next)
	if len(workingFiles) > 0 {
		index.State.WorkingFiles = workingFiles
	}
	if len(openQuestions) > 0 {
		index.State.OpenQuestions = openQuestions
	}

	if err := s.SaveIndex(index); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	return nil
}

// Recall returns all knowledge for a topic with summaries
func (s *Store) Recall(topic string) (string, error) {
	index, err := s.LoadIndex()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("# Knowledge for topic: %s\n\n", topic))

	// Lessons
	var lessonIDs []string
	for id, entry := range index.Lessons {
		if entry.Topic == topic {
			lessonIDs = append(lessonIDs, id)
		}
	}
	sort.Strings(lessonIDs)
	if len(lessonIDs) > 0 {
		buf.WriteString("## Lessons\n\n")
		for _, id := range lessonIDs {
			entry := index.Lessons[id]
			buf.WriteString(fmt.Sprintf("### %s [%s]\n", id, entry.Severity))
			buf.WriteString(fmt.Sprintf("%s\n\n", entry.Oneliner))

			// Load summary from file if available
			engFile, err := fileio.ParseEngFile(filepath.Join(s.KnowledgeDir, id+".eng"))
			if err == nil {
				if summary, ok := engFile.Frontmatter["summary"].(string); ok && summary != "" {
					buf.WriteString(fmt.Sprintf("%s\n\n", strings.TrimSpace(summary)))
				}
			}
		}
	}

	// Decisions
	var decisionIDs []string
	for id, entry := range index.Decisions {
		if entry.Topic == topic {
			decisionIDs = append(decisionIDs, id)
		}
	}
	sort.Strings(decisionIDs)
	if len(decisionIDs) > 0 {
		buf.WriteString("## Decisions\n\n")
		for _, id := range decisionIDs {
			entry := index.Decisions[id]
			buf.WriteString(fmt.Sprintf("### %s\n", id))
			buf.WriteString(fmt.Sprintf("%s\n\n", entry.Oneliner))

			engFile, err := fileio.ParseEngFile(filepath.Join(s.KnowledgeDir, id+".eng"))
			if err == nil {
				if summary, ok := engFile.Frontmatter["summary"].(string); ok && summary != "" {
					buf.WriteString(fmt.Sprintf("%s\n\n", strings.TrimSpace(summary)))
				}
			}
		}
	}

	// Patterns
	var patternIDs []string
	for id, entry := range index.Patterns {
		if entry.Domain == topic {
			patternIDs = append(patternIDs, id)
		}
	}
	sort.Strings(patternIDs)
	if len(patternIDs) > 0 {
		buf.WriteString("## Patterns\n\n")
		for _, id := range patternIDs {
			entry := index.Patterns[id]
			buf.WriteString(fmt.Sprintf("### %s\n", id))
			buf.WriteString(fmt.Sprintf("%s\n\n", entry.Oneliner))

			engFile, err := fileio.ParseEngFile(filepath.Join(s.KnowledgeDir, id+".eng"))
			if err == nil {
				if summary, ok := engFile.Frontmatter["summary"].(string); ok && summary != "" {
					buf.WriteString(fmt.Sprintf("%s\n\n", strings.TrimSpace(summary)))
				}
			}
		}
	}

	if len(lessonIDs) == 0 && len(decisionIDs) == 0 && len(patternIDs) == 0 {
		buf.WriteString("No knowledge found for this topic.\n")
	}

	return buf.String(), nil
}

// Show returns the full content for a specific item
func (s *Store) Show(id string) (string, error) {
	engPath := filepath.Join(s.KnowledgeDir, id+".eng")
	if !fileio.FileExists(engPath) {
		return "", fmt.Errorf("item %s not found", id)
	}
	return fileio.ReadFileContent(engPath)
}

// ListItem represents an item in the list output
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
	index, err := s.LoadIndex()
	if err != nil {
		return nil, err
	}

	var items []ListItem

	// Filter lessons
	if itemType == "" || itemType == "lesson" {
		for id, entry := range index.Lessons {
			if topic != "" && entry.Topic != topic {
				continue
			}
			if severity != "" && entry.Severity != severity {
				continue
			}
			if !since.IsZero() && entry.Created.Before(since) {
				continue
			}
			if !until.IsZero() && entry.Created.After(until) {
				continue
			}
			items = append(items, ListItem{
				ID:       id,
				Type:     "lesson",
				Oneliner: entry.Oneliner,
				Topic:    entry.Topic,
				Severity: entry.Severity,
				Created:  entry.Created.Format("2006-01-02"),
			})
		}
	}

	// Filter decisions
	if itemType == "" || itemType == "decision" {
		for id, entry := range index.Decisions {
			if topic != "" && entry.Topic != topic {
				continue
			}
			if severity != "" {
				continue // decisions don't have severity
			}
			if !since.IsZero() && entry.Created.Before(since) {
				continue
			}
			if !until.IsZero() && entry.Created.After(until) {
				continue
			}
			items = append(items, ListItem{
				ID:       id,
				Type:     "decision",
				Oneliner: entry.Oneliner,
				Topic:    entry.Topic,
				Created:  entry.Created.Format("2006-01-02"),
			})
		}
	}

	// Filter patterns
	if itemType == "" || itemType == "pattern" {
		for id, entry := range index.Patterns {
			if topic != "" && entry.Domain != topic {
				continue
			}
			if severity != "" {
				continue // patterns don't have severity
			}
			if !since.IsZero() && entry.Created.Before(since) {
				continue
			}
			if !until.IsZero() && entry.Created.After(until) {
				continue
			}
			items = append(items, ListItem{
				ID:       id,
				Type:     "pattern",
				Oneliner: entry.Oneliner,
				Domain:   entry.Domain,
				Created:  entry.Created.Format("2006-01-02"),
			})
		}
	}

	// Sort by ID
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return items, nil
}

// TopicInfo represents a topic with its count
type TopicInfo struct {
	Name  string `json:"name" yaml:"name"`
	Count int    `json:"count" yaml:"count"`
}

// Topics returns all topics with their item counts
func (s *Store) Topics() ([]TopicInfo, error) {
	index, err := s.LoadIndex()
	if err != nil {
		return nil, err
	}

	counts := index.GetTopicCounts()
	var topics []TopicInfo
	for name, count := range counts {
		topics = append(topics, TopicInfo{Name: name, Count: count})
	}

	// Sort by name
	sort.Slice(topics, func(i, j int) bool {
		return topics[i].Name < topics[j].Name
	})

	return topics, nil
}

// Remove deletes an item by ID
func (s *Store) Remove(id string) error {
	engPath := filepath.Join(s.KnowledgeDir, id+".eng")
	if !fileio.FileExists(engPath) {
		return fmt.Errorf("item %s not found", id)
	}

	// Remove from index
	index, err := s.LoadIndex()
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}
	if !index.RemoveItem(id) {
		return fmt.Errorf("item %s not found in index", id)
	}
	if err := s.SaveIndex(index); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	// Remove file
	if err := os.Remove(engPath); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	return nil
}

// Rebuild rebuilds the index from .eng files
func (s *Store) Rebuild() error {
	files, err := fileio.ListEngFiles(s.KnowledgeDir)
	if err != nil {
		return fmt.Errorf("failed to list knowledge files: %w", err)
	}

	// Load existing index to preserve project info and state
	index, err := s.LoadIndex()
	if err != nil {
		index = models.NewIndex("unknown")
	}

	// Preserve state
	state := index.State

	// Clear existing entries
	index.Lessons = make(map[string]models.IndexEntry)
	index.Decisions = make(map[string]models.IndexEntry)
	index.Patterns = make(map[string]models.IndexEntry)
	index.Topics = []string{}
	index.Edges = make(map[string][]string)
	index.State = state

	for _, file := range files {
		engFile, err := fileio.ParseEngFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", file, err)
			continue
		}

		id, _ := engFile.Frontmatter["id"].(string)
		itemType, _ := engFile.Frontmatter["type"].(string)
		oneliner := extractOneliner(engFile)
		created := parseCreated(engFile.Frontmatter["created"])

		switch itemType {
		case "lesson":
			topic, _ := engFile.Frontmatter["topic"].(string)
			severityStr, _ := engFile.Frontmatter["severity"].(string)
			severity := models.Severity(severityStr)
			index.AddLesson(id, oneliner, topic, severity, created)
		case "decision":
			topic, _ := engFile.Frontmatter["topic"].(string)
			index.AddDecision(id, oneliner, topic, created)
		case "pattern":
			domain, _ := engFile.Frontmatter["domain"].(string)
			index.AddPattern(id, oneliner, domain, created)
		}

		// Extract refs and add edges
		if refsRaw, ok := engFile.Frontmatter["refs"]; ok {
			if refsList, ok := refsRaw.([]interface{}); ok {
				for _, ref := range refsList {
					if refID, ok := ref.(string); ok {
						index.AddEdge(id, refID)
					}
				}
			}
		}
	}

	if err := s.SaveIndex(index); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	return nil
}

// GetFilePath returns the path to an item's .eng file
func (s *Store) GetFilePath(id string) string {
	return filepath.Join(s.KnowledgeDir, id+".eng")
}

// extractOneliner tries to extract a oneliner from the eng file
func extractOneliner(eng *fileio.EngFile) string {
	// First try to get from frontmatter oneliner field
	if oneliner, ok := eng.Frontmatter["oneliner"].(string); ok && oneliner != "" {
		return oneliner
	}

	// Try to get from summary
	if summary, ok := eng.Frontmatter["summary"].(string); ok && summary != "" {
		lines := strings.Split(strings.TrimSpace(summary), "\n")
		if len(lines) > 0 {
			return lines[0]
		}
	}

	// Try to extract from body - first heading
	lines := strings.Split(eng.Body, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}

	return "(no description)"
}

// parseCreated parses a created date from frontmatter
func parseCreated(v interface{}) time.Time {
	switch val := v.(type) {
	case string:
		// Try RFC3339 first (new format with time)
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			return t
		}
		// Fall back to date-only format (legacy)
		if t, err := time.Parse("2006-01-02", val); err == nil {
			return t
		}
	case time.Time:
		return val
	}
	return time.Now()
}
