package doryfile

import "fmt"

func (df *DoryFile) applyEvent(seq uint64, ev *logEvent, payloadOffset int64, payloadLen int) error {
	switch ev.Op {
	case opItemCreate, opItemUpdate:
		if ev.Item == nil || ev.Item.ID == "" {
			return fmt.Errorf("invalid %s event: missing item", ev.Op)
		}
		df.entries[ev.Item.ID] = memoryEntryFromEntry(ev.Item, payloadOffset, payloadLen)
		df.removeDeletedID(ev.Item.ID)
	case opItemDelete:
		if ev.ID == "" {
			return fmt.Errorf("invalid %s event: missing id", ev.Op)
		}
		delete(df.entries, ev.ID)
		if !containsString(df.Index.Deleted, ev.ID) {
			df.Index.Deleted = append(df.Index.Deleted, ev.ID)
		}
	case opState:
		if ev.State != nil {
			df.Index.State = cloneState(ev.State)
		}
	case opCompact:
		// Metadata marker only.
	default:
		return fmt.Errorf("unknown event op %q", ev.Op)
	}

	if seq > df.nextSeq {
		df.nextSeq = seq
	}
	return nil
}

func memoryEntryFromEntry(entry *Entry, payloadOffset int64, payloadLen int) *MemoryEntry {
	return &MemoryEntry{
		Offset:   payloadOffset,
		BodyLen:  payloadLen,
		Type:     entry.Type,
		Topic:    entry.Topic,
		Domain:   entry.Domain,
		Severity: entry.Severity,
		Oneliner: entry.Oneliner,
		Created:  entry.Created,
		Refs:     append([]string(nil), entry.Refs...),
	}
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func cloneState(state *State) *State {
	if state == nil {
		return nil
	}
	copied := *state
	copied.Next = append([]string(nil), state.Next...)
	copied.WorkingFiles = append([]string(nil), state.WorkingFiles...)
	copied.OpenQuestions = append([]string(nil), state.OpenQuestions...)
	return &copied
}
