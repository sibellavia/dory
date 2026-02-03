package doryfile

import (
	"os"
	"sort"
)

func (df *DoryFile) sortedLiveEntries() []*Entry {
	ids := make([]string, 0, len(df.entries))
	for id := range df.entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	entries := make([]*Entry, 0, len(ids))
	for _, id := range ids {
		entry, err := df.Get(id)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	return entries
}

func isStateEmpty(state *State) bool {
	if state == nil {
		return true
	}
	return state.Goal == "" &&
		state.Progress == "" &&
		state.Blocker == "" &&
		len(state.Next) == 0 &&
		len(state.WorkingFiles) == 0 &&
		len(state.OpenQuestions) == 0 &&
		state.LastUpdated == ""
}

// Lessons returns all lessons from in-memory index.
func (df *DoryFile) Lessons() map[string]*MemoryEntry {
	result := make(map[string]*MemoryEntry)
	for id, entry := range df.entries {
		if entry.Type == "lesson" {
			result[id] = entry
		}
	}
	return result
}

// Decisions returns all decisions from in-memory index.
func (df *DoryFile) Decisions() map[string]*MemoryEntry {
	result := make(map[string]*MemoryEntry)
	for id, entry := range df.entries {
		if entry.Type == "decision" {
			result[id] = entry
		}
	}
	return result
}

// Patterns returns all patterns from in-memory index.
func (df *DoryFile) Patterns() map[string]*MemoryEntry {
	result := make(map[string]*MemoryEntry)
	for id, entry := range df.entries {
		if entry.Type == "pattern" {
			result[id] = entry
		}
	}
	return result
}

// Entries returns all entries from the in-memory index.
func (df *DoryFile) Entries() map[string]*MemoryEntry {
	result := make(map[string]*MemoryEntry)
	for id, entry := range df.entries {
		result[id] = entry
	}
	return result
}

// DumpKnowledge returns the raw knowledge file content.
func (df *DoryFile) DumpKnowledge() (string, error) {
	data, err := os.ReadFile(df.KnowledgePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DumpIndex returns the index file content.
func (df *DoryFile) DumpIndex() (string, error) {
	data, err := os.ReadFile(df.IndexPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
