package doryfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sibellavia/dory/internal/fileio"
	"gopkg.in/yaml.v3"
)

// Append adds a new entry (append-only).
func (df *DoryFile) Append(entry *Entry) error {
	op := opItemCreate
	if _, exists := df.entries[entry.ID]; exists {
		op = opItemUpdate
	}
	ev := &logEvent{Op: op, Item: entry}

	payloadOffset, payloadLen, seq, err := df.appendEvent(ev)
	if err != nil {
		return err
	}
	if err := df.applyEvent(seq, ev, payloadOffset, payloadLen); err != nil {
		return err
	}
	return df.saveIndex()
}

func (df *DoryFile) appendEvent(ev *logEvent) (int64, int, uint64, error) {
	if df.knowledge == nil {
		return 0, 0, 0, fmt.Errorf("knowledge file is not open")
	}
	stat, err := df.knowledge.Stat()
	if err != nil {
		return 0, 0, 0, err
	}
	start := stat.Size()

	payload, err := marshalEvent(ev)
	if err != nil {
		return 0, 0, 0, err
	}

	seq := df.nextSeq + 1
	if _, err := df.knowledge.WriteString(EventDelim + "\n"); err != nil {
		return 0, 0, 0, err
	}
	if _, err := df.knowledge.Write(payload); err != nil {
		return 0, 0, 0, err
	}
	if err := df.knowledge.Sync(); err != nil {
		return 0, 0, 0, err
	}

	payloadOffset := start + int64(len(EventDelim)+1)
	df.nextSeq = seq
	df.logOffset = payloadOffset + int64(len(payload))
	return payloadOffset, len(payload), seq, nil
}

func marshalEvent(ev *logEvent) ([]byte, error) {
	payload, err := yaml.Marshal(ev)
	if err != nil {
		return nil, err
	}
	if len(payload) == 0 || payload[len(payload)-1] != '\n' {
		payload = append(payload, '\n')
	}
	return payload, nil
}

// Get retrieves an entry by ID.
func (df *DoryFile) Get(id string) (*Entry, error) {
	mem, ok := df.entries[id]
	if !ok {
		return nil, fmt.Errorf("item %s not found", id)
	}
	if mem.BodyLen <= 0 {
		return nil, fmt.Errorf("item %s has invalid payload length", id)
	}

	if _, err := df.knowledge.Seek(mem.Offset, 0); err != nil {
		return nil, err
	}
	payload := make([]byte, mem.BodyLen)
	if _, err := io.ReadFull(df.knowledge, payload); err != nil {
		return nil, err
	}

	var ev logEvent
	if err := yaml.Unmarshal(payload, &ev); err != nil {
		return nil, err
	}
	if ev.Item == nil {
		return nil, fmt.Errorf("item %s payload missing item body", id)
	}
	return ev.Item, nil
}

// Delete removes an entry.
func (df *DoryFile) Delete(id string) error {
	if _, ok := df.entries[id]; !ok {
		return fmt.Errorf("item %s not found", id)
	}

	ev := &logEvent{Op: opItemDelete, ID: id}
	payloadOffset, payloadLen, seq, err := df.appendEvent(ev)
	if err != nil {
		return err
	}
	if err := df.applyEvent(seq, ev, payloadOffset, payloadLen); err != nil {
		return err
	}
	return df.saveIndex()
}

func (df *DoryFile) removeDeletedID(id string) bool {
	for i, deletedID := range df.Index.Deleted {
		if deletedID == id {
			df.Index.Deleted = append(df.Index.Deleted[:i], df.Index.Deleted[i+1:]...)
			return true
		}
	}
	return false
}

// UpdateState updates the session state.
func (df *DoryFile) UpdateState(state *State) error {
	ev := &logEvent{Op: opState, State: cloneState(state)}
	payloadOffset, payloadLen, seq, err := df.appendEvent(ev)
	if err != nil {
		return err
	}
	if err := df.applyEvent(seq, ev, payloadOffset, payloadLen); err != nil {
		return err
	}
	return df.saveIndex()
}

// Compact rewrites the knowledge file, removing deleted entries.
func (df *DoryFile) Compact() error {
	entries := df.sortedLiveEntries()
	state := cloneState(df.Index.State)

	if err := df.knowledge.Close(); err != nil {
		return err
	}

	tmpPath := df.KnowledgePath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	header := fmt.Sprintf("%s\n", MagicHeader)
	if _, err := tmpFile.WriteString(header); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}

	df.entries = make(map[string]*MemoryEntry)
	seq := uint64(0)
	currentOffset := int64(len(header))

	writeEvent := func(ev *logEvent) (int64, int, uint64, error) {
		payload, err := marshalEvent(ev)
		if err != nil {
			return 0, 0, 0, err
		}
		seq++

		start := currentOffset
		if _, err := tmpFile.WriteString(EventDelim + "\n"); err != nil {
			return 0, 0, 0, err
		}
		if _, err := tmpFile.Write(payload); err != nil {
			return 0, 0, 0, err
		}

		payloadOffset := start + int64(len(EventDelim)+1)
		currentOffset = payloadOffset + int64(len(payload))
		return payloadOffset, len(payload), seq, nil
	}

	if state != nil && !isStateEmpty(state) {
		ev := &logEvent{Op: opState, State: state}
		if _, _, _, err := writeEvent(ev); err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
			return err
		}
	}

	for _, entry := range entries {
		ev := &logEvent{Op: opItemCreate, Item: entry}
		offset, payloadLen, eventSeq, err := writeEvent(ev)
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
			return err
		}
		df.entries[entry.ID] = memoryEntryFromEntry(entry, offset, payloadLen)
		df.nextSeq = eventSeq
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, df.KnowledgePath); err != nil {
		return err
	}
	if err := fileio.SyncDir(filepath.Dir(df.KnowledgePath)); err != nil {
		return err
	}

	df.knowledge, err = os.OpenFile(df.KnowledgePath, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	df.logOffset = currentOffset
	df.Index.Deleted = nil
	return df.saveIndex()
}
