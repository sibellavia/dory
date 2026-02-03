package store

import (
	"fmt"
	"sort"
)

// Refs returns relationship information for an item.
func (s *Store) Refs(id string) (*RefInfo, error) {
	if err := s.openLatest(); err != nil {
		return nil, err
	}

	entries := s.df.Entries()

	entry, ok := entries[id]
	if !ok {
		return nil, fmt.Errorf("item %s not found", id)
	}

	referencedBy := make(map[string]bool)
	for otherID, otherEntry := range entries {
		if otherID == id {
			continue
		}
		for _, ref := range otherEntry.Refs {
			if ref == id {
				referencedBy[otherID] = true
				break
			}
		}
	}

	var refsTo []RefItem
	for _, refID := range entry.Refs {
		if refEntry, ok := entries[refID]; ok {
			refsTo = append(refsTo, RefItem{
				ID:       refID,
				Type:     refEntry.Type,
				Oneliner: refEntry.Oneliner,
			})
		} else {
			refsTo = append(refsTo, RefItem{
				ID:       refID,
				Type:     "unknown",
				Oneliner: "(not found)",
			})
		}
	}

	var refBy []RefItem
	var refByIDs []string
	for refID := range referencedBy {
		refByIDs = append(refByIDs, refID)
	}
	sort.Strings(refByIDs)

	for _, refID := range refByIDs {
		refEntry := entries[refID]
		refBy = append(refBy, RefItem{
			ID:       refID,
			Type:     refEntry.Type,
			Oneliner: refEntry.Oneliner,
		})
	}

	return &RefInfo{
		ID:           id,
		Type:         entry.Type,
		Oneliner:     entry.Oneliner,
		RefsTo:       refsTo,
		ReferencedBy: refBy,
	}, nil
}

// Expand returns an item and all items connected to it within depth hops.
func (s *Store) Expand(id string, depth int) (*ExpandResult, error) {
	if err := s.openLatest(); err != nil {
		return nil, err
	}

	if depth < 1 {
		depth = 1
	}

	entries := s.df.Entries()

	if _, ok := entries[id]; !ok {
		return nil, fmt.Errorf("item %s not found", id)
	}

	referencedBy := make(map[string][]string)
	for entryID, entry := range entries {
		for _, ref := range entry.Refs {
			referencedBy[ref] = append(referencedBy[ref], entryID)
		}
	}

	visited := make(map[string]bool)
	visited[id] = true
	queue := []struct {
		id    string
		depth int
	}{{id, 0}}

	var connectedIDs []string

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.depth >= depth {
			continue
		}

		entry := entries[current.id]

		for _, refID := range entry.Refs {
			if !visited[refID] {
				visited[refID] = true
				if _, exists := entries[refID]; exists {
					connectedIDs = append(connectedIDs, refID)
					queue = append(queue, struct {
						id    string
						depth int
					}{refID, current.depth + 1})
				}
			}
		}

		for _, refByID := range referencedBy[current.id] {
			if !visited[refByID] {
				visited[refByID] = true
				connectedIDs = append(connectedIDs, refByID)
				queue = append(queue, struct {
					id    string
					depth int
				}{refByID, current.depth + 1})
			}
		}
	}

	sort.Strings(connectedIDs)

	rootEntry, err := s.df.Get(id)
	if err != nil {
		return nil, err
	}

	result := &ExpandResult{
		Root: ExpandedItem{
			ID:       rootEntry.ID,
			Type:     rootEntry.Type,
			Oneliner: rootEntry.Oneliner,
			Topic:    rootEntry.Topic,
			Domain:   rootEntry.Domain,
			Refs:     rootEntry.Refs,
			Body:     rootEntry.Body,
		},
	}

	for _, connID := range connectedIDs {
		entry, err := s.df.Get(connID)
		if err != nil {
			continue
		}
		result.Connected = append(result.Connected, ExpandedItem{
			ID:       entry.ID,
			Type:     entry.Type,
			Oneliner: entry.Oneliner,
			Topic:    entry.Topic,
			Domain:   entry.Domain,
			Refs:     entry.Refs,
			Body:     entry.Body,
		})
	}

	return result, nil
}
