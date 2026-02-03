package store

import (
	"sort"
	"time"
)

// Context returns smart context for agent session start.
func (s *Store) Context(topic string, recentDays int, full bool) (*ContextResult, error) {
	if err := s.openLatest(); err != nil {
		return nil, err
	}

	result := &ContextResult{Project: s.df.Index.Project}

	if s.df.Index.State != nil {
		st := s.df.Index.State
		result.State = &ContextState{
			Goal:        st.Goal,
			Progress:    st.Progress,
			Blocker:     st.Blocker,
			Next:        st.Next,
			LastUpdated: st.LastUpdated,
		}
	}

	recentCutoff := time.Now().AddDate(0, 0, -recentDays)
	entries := s.df.Entries()

	critical := make([]ListItem, 0)
	recent := make([]ListItem, 0)
	topicItems := make([]ListItem, 0)

	for id, entry := range entries {
		item := toListItem(id, entry)

		if entry.Type == "lesson" && (entry.Severity == "critical" || entry.Severity == "high") {
			critical = append(critical, item)
		}

		if entry.Created.After(recentCutoff) {
			recent = append(recent, item)
		}

		if topic != "" && (entry.Topic == topic || entry.Domain == topic) {
			topicItems = append(topicItems, item)
		}

		if full && !entry.Created.After(recentCutoff) {
			recent = append(recent, item)
		}
	}

	sortItems := func(items []ListItem) {
		sort.Slice(items, func(i, j int) bool {
			return items[i].ID < items[j].ID
		})
	}

	sortItems(critical)
	sortItems(recent)
	sortItems(topicItems)

	criticalIDs := make(map[string]bool)
	for _, item := range critical {
		criticalIDs[item.ID] = true
	}
	dedupedRecent := make([]ListItem, 0)
	for _, item := range recent {
		if !criticalIDs[item.ID] {
			dedupedRecent = append(dedupedRecent, item)
		}
	}

	result.Critical = critical
	result.Recent = dedupedRecent
	if topic != "" {
		result.Topic = topicItems
	}

	return result, nil
}
