package store

import (
	"fmt"
	"time"

	"github.com/sibellavia/dory/internal/doryfile"
	"github.com/sibellavia/dory/internal/idgen"
	"github.com/sibellavia/dory/internal/models"
)

// Learn adds a new lesson.
func (s *Store) Learn(oneliner, topic string, severity models.Severity, body string, refs []string) (string, error) {
	var id string
	err := s.withWriteLock(func() error {
		if err := s.open(); err != nil {
			return err
		}

		var err error
		id, err = idgen.NewItemID("lesson")
		if err != nil {
			return err
		}

		fullBody := body
		if fullBody == "" {
			fullBody = fmt.Sprintf("# %s\n\n## Details\n\n(Add details here)\n", oneliner)
		}

		entry := newEntry(id, "lesson", oneliner, topic, "", string(severity), fullBody, refs)
		if err := s.df.Append(entry); err != nil {
			return fmt.Errorf("failed to append lesson: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

// Decide adds a new decision.
func (s *Store) Decide(oneliner, topic, rationale, body string, refs []string) (string, error) {
	var id string
	err := s.withWriteLock(func() error {
		if err := s.open(); err != nil {
			return err
		}

		var err error
		id, err = idgen.NewItemID("decision")
		if err != nil {
			return err
		}

		fullBody := body
		if fullBody == "" {
			fullBody = fmt.Sprintf("# %s\n\n## Context\n\n(Add context here)\n\n## Decision\n\n%s\n\n## Rationale\n\n%s\n", oneliner, oneliner, rationale)
		}

		entry := newEntry(id, "decision", oneliner, topic, "", "", fullBody, refs)
		if err := s.df.Append(entry); err != nil {
			return fmt.Errorf("failed to append decision: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

// Pattern adds a new pattern.
func (s *Store) Pattern(oneliner, domain, body string, refs []string) (string, error) {
	var id string
	err := s.withWriteLock(func() error {
		if err := s.open(); err != nil {
			return err
		}

		var err error
		id, err = idgen.NewItemID("pattern")
		if err != nil {
			return err
		}

		fullBody := body
		if fullBody == "" {
			fullBody = fmt.Sprintf("# %s\n\n## Pattern\n\n%s\n\n## Implementation\n\n(Add implementation details here)\n", oneliner, oneliner)
		}

		entry := newEntry(id, "pattern", oneliner, "", domain, "", fullBody, refs)
		if err := s.df.Append(entry); err != nil {
			return fmt.Errorf("failed to append pattern: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

// CreateCustom adds a new custom type entry.
func (s *Store) CreateCustom(itemType, oneliner, topic, body string, refs []string) (string, error) {
	var id string
	err := s.withWriteLock(func() error {
		if err := s.open(); err != nil {
			return err
		}

		var err error
		id, err = idgen.NewItemID(itemType)
		if err != nil {
			return err
		}

		fullBody := body
		if fullBody == "" {
			fullBody = fmt.Sprintf("# %s\n\n## Context\n\n(Add context here)\n", oneliner)
		}

		entry := newEntry(id, itemType, oneliner, topic, "", "", fullBody, refs)
		if err := s.df.Append(entry); err != nil {
			return fmt.Errorf("failed to append %s: %w", itemType, err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

// UpdateEntry appends a new version of an existing entry.
func (s *Store) UpdateEntry(entry *doryfile.Entry) error {
	return s.withWriteLock(func() error {
		if err := s.open(); err != nil {
			return err
		}
		return s.df.Append(entry)
	})
}

// Remove deletes an item by ID.
func (s *Store) Remove(id string) error {
	return s.withWriteLock(func() error {
		if err := s.open(); err != nil {
			return err
		}
		return s.df.Delete(id)
	})
}

// UpdateStatus updates the session state.
func (s *Store) UpdateStatus(goal, progress, blocker string, next, workingFiles, openQuestions []string) error {
	return s.withWriteLock(func() error {
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
		state.LastUpdated = time.Now().UTC().Format(time.RFC3339)

		return s.df.UpdateState(state)
	})
}

// Compact removes deleted entries and rebuilds the file.
func (s *Store) Compact() error {
	return s.withWriteLock(func() error {
		if err := s.open(); err != nil {
			return err
		}
		return s.df.Compact()
	})
}

func newEntry(id, itemType, oneliner, topic, domain, severity, body string, refs []string) *doryfile.Entry {
	return &doryfile.Entry{
		ID:       id,
		Type:     itemType,
		Topic:    topic,
		Domain:   domain,
		Severity: severity,
		Oneliner: oneliner,
		Created:  time.Now().UTC(),
		Refs:     refs,
		Body:     body,
	}
}
