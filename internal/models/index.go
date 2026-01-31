package models

import "time"

// IndexEntry represents a single entry in the index
type IndexEntry struct {
	Oneliner string    `yaml:"oneliner"`
	Topic    string    `yaml:"topic,omitempty"`
	Domain   string    `yaml:"domain,omitempty"` // For patterns
	Severity Severity  `yaml:"severity,omitempty"`
	Created  time.Time `yaml:"created"`
}

// Index represents the index.yaml file structure
type Index struct {
	Version     int                   `yaml:"version"`
	Project     string                `yaml:"project"`
	Description string                `yaml:"description,omitempty"`
	Lessons     map[string]IndexEntry `yaml:"lessons,omitempty"`
	Decisions   map[string]IndexEntry `yaml:"decisions,omitempty"`
	Patterns    map[string]IndexEntry `yaml:"patterns,omitempty"`
	Topics      []string              `yaml:"topics,omitempty"`
}

// NewIndex creates a new empty index
func NewIndex(project string) *Index {
	return &Index{
		Version:   1,
		Project:   project,
		Lessons:   make(map[string]IndexEntry),
		Decisions: make(map[string]IndexEntry),
		Patterns:  make(map[string]IndexEntry),
		Topics:    []string{},
	}
}

// AddLesson adds a lesson entry to the index
func (idx *Index) AddLesson(id, oneliner, topic string, severity Severity, created time.Time) {
	idx.Lessons[id] = IndexEntry{
		Oneliner: oneliner,
		Topic:    topic,
		Severity: severity,
		Created:  created,
	}
	idx.addTopic(topic)
}

// AddDecision adds a decision entry to the index
func (idx *Index) AddDecision(id, oneliner, topic string, created time.Time) {
	idx.Decisions[id] = IndexEntry{
		Oneliner: oneliner,
		Topic:    topic,
		Created:  created,
	}
	idx.addTopic(topic)
}

// AddPattern adds a pattern entry to the index
func (idx *Index) AddPattern(id, oneliner, domain string, created time.Time) {
	idx.Patterns[id] = IndexEntry{
		Oneliner: oneliner,
		Domain:   domain,
		Created:  created,
	}
	idx.addTopic(domain)
}

// RemoveItem removes an item from the index by ID
func (idx *Index) RemoveItem(id string) bool {
	if _, ok := idx.Lessons[id]; ok {
		delete(idx.Lessons, id)
		idx.rebuildTopics()
		return true
	}
	if _, ok := idx.Decisions[id]; ok {
		delete(idx.Decisions, id)
		idx.rebuildTopics()
		return true
	}
	if _, ok := idx.Patterns[id]; ok {
		delete(idx.Patterns, id)
		idx.rebuildTopics()
		return true
	}
	return false
}

// addTopic adds a topic if it doesn't exist
func (idx *Index) addTopic(topic string) {
	if topic == "" {
		return
	}
	for _, t := range idx.Topics {
		if t == topic {
			return
		}
	}
	idx.Topics = append(idx.Topics, topic)
}

// rebuildTopics rebuilds the topics list from all entries
func (idx *Index) rebuildTopics() {
	topicSet := make(map[string]bool)
	for _, entry := range idx.Lessons {
		if entry.Topic != "" {
			topicSet[entry.Topic] = true
		}
	}
	for _, entry := range idx.Decisions {
		if entry.Topic != "" {
			topicSet[entry.Topic] = true
		}
	}
	for _, entry := range idx.Patterns {
		if entry.Domain != "" {
			topicSet[entry.Domain] = true
		}
	}
	idx.Topics = make([]string, 0, len(topicSet))
	for topic := range topicSet {
		idx.Topics = append(idx.Topics, topic)
	}
}

// GetTopicCounts returns a map of topic -> count of items
func (idx *Index) GetTopicCounts() map[string]int {
	counts := make(map[string]int)
	for _, entry := range idx.Lessons {
		if entry.Topic != "" {
			counts[entry.Topic]++
		}
	}
	for _, entry := range idx.Decisions {
		if entry.Topic != "" {
			counts[entry.Topic]++
		}
	}
	for _, entry := range idx.Patterns {
		if entry.Domain != "" {
			counts[entry.Domain]++
		}
	}
	return counts
}
