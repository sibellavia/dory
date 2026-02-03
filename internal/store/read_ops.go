package store

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/sibellavia/dory/internal/doryfile"
	"github.com/sibellavia/dory/internal/models"
	"gopkg.in/yaml.v3"
)

// Show returns the full content for a specific item.
func (s *Store) Show(id string) (string, error) {
	entry, err := s.GetEntry(id)
	if err != nil {
		return "", err
	}

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

// GetEntry returns a raw entry by ID.
func (s *Store) GetEntry(id string) (*doryfile.Entry, error) {
	if err := s.openLatest(); err != nil {
		return nil, err
	}
	return s.df.Get(id)
}

// List returns items matching the filters.
func (s *Store) List(topic, itemType string, severity models.Severity, since, until time.Time) ([]ListItem, error) {
	if err := s.openLatest(); err != nil {
		return nil, err
	}

	items := make([]ListItem, 0)
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

		items = append(items, toListItem(id, entry))
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return items, nil
}

// Recall returns all knowledge for a topic with summaries.
func (s *Store) Recall(topic string) (string, error) {
	if err := s.openLatest(); err != nil {
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

// Topics returns all topics with their item counts.
func (s *Store) Topics() ([]TopicInfo, error) {
	if err := s.openLatest(); err != nil {
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

	topics := make([]TopicInfo, 0)
	for name, count := range counts {
		topics = append(topics, TopicInfo{Name: name, Count: count})
	}

	sort.Slice(topics, func(i, j int) bool {
		return topics[i].Name < topics[j].Name
	})

	return topics, nil
}

// DumpContent returns the raw content file (for debugging/viewing).
func (s *Store) DumpContent() (string, error) {
	if err := s.openLatest(); err != nil {
		return "", err
	}
	return s.df.DumpKnowledge()
}

// DumpIndex returns the index file content.
func (s *Store) DumpIndex() (string, error) {
	if err := s.openLatest(); err != nil {
		return "", err
	}
	return s.df.DumpIndex()
}

func toListItem(id string, entry *doryfile.MemoryEntry) ListItem {
	item := ListItem{
		ID:        id,
		Type:      entry.Type,
		Oneliner:  entry.Oneliner,
		Created:   entry.Created.Format("2006-01-02"),
		CreatedAt: entry.Created.UTC().Format(time.RFC3339Nano),
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
	return item
}
