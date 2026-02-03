package doryfile

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Create creates a new dory storage.
func Create(dir, project, description string) (*DoryFile, error) {
	knowledgePath := filepath.Join(dir, KnowledgeFile)
	indexPath := filepath.Join(dir, IndexFile)

	f, err := os.Create(knowledgePath)
	if err != nil {
		return nil, err
	}
	header := fmt.Sprintf("%s\n", MagicHeader)
	if _, err := f.WriteString(header); err != nil {
		f.Close()
		return nil, err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return nil, err
	}

	df := &DoryFile{
		Dir:           dir,
		KnowledgePath: knowledgePath,
		IndexPath:     indexPath,
		knowledge:     f,
		nextSeq:       0,
		logOffset:     int64(len(header)),
		Index: &Index{
			Version:     2,
			Format:      IndexFormat,
			Project:     project,
			Description: description,
			State:       &State{},
			AppliedSeq:  0,
			LogOffset:   int64(len(header)),
		},
		entries: make(map[string]*MemoryEntry),
	}

	if err := df.saveIndex(); err != nil {
		f.Close()
		return nil, err
	}

	return df, nil
}

// Open opens an existing dory storage.
func Open(dir string) (*DoryFile, error) {
	knowledgePath := filepath.Join(dir, KnowledgeFile)
	indexPath := filepath.Join(dir, IndexFile)

	f, err := os.OpenFile(knowledgePath, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	df := &DoryFile{
		Dir:           dir,
		KnowledgePath: knowledgePath,
		IndexPath:     indexPath,
		knowledge:     f,
		entries:       make(map[string]*MemoryEntry),
	}

	if err := df.loadIndex(); err != nil {
		f.Close()
		return nil, err
	}

	if err := df.scan(); err != nil {
		f.Close()
		return nil, err
	}

	return df, nil
}

// scan reads the knowledge file and builds the in-memory index.
func (df *DoryFile) scan() error {
	if _, err := df.knowledge.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek knowledge file: %w", err)
	}
	reader := bufio.NewReader(df.knowledge)

	header, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}
	if strings.TrimSpace(header) != MagicHeader {
		return fmt.Errorf("invalid dory file header: expected %s", MagicHeader)
	}

	return df.scanEvents(reader, int64(len(header)))
}

func (df *DoryFile) scanEvents(reader *bufio.Reader, startPos int64) error {
	// Try snapshot hydrate and replay only tail.
	if df.hydrateFromSnapshot(startPos) {
		if _, err := df.knowledge.Seek(df.logOffset, 0); err == nil {
			reader = bufio.NewReader(df.knowledge)
			if err := df.replayEvents(reader, df.logOffset, df.nextSeq+1); err == nil {
				return nil
			}
		}
	}

	// Fallback to full replay.
	df.entries = make(map[string]*MemoryEntry)
	df.nextSeq = 0
	df.logOffset = startPos
	df.Index.Deleted = nil
	df.Index.State = &State{}

	if _, err := df.knowledge.Seek(startPos, 0); err != nil {
		return fmt.Errorf("failed to seek replay start: %w", err)
	}
	reader = bufio.NewReader(df.knowledge)
	return df.replayEvents(reader, startPos, 1)
}

func (df *DoryFile) hydrateFromSnapshot(startPos int64) bool {
	if df.Index == nil || df.Index.Version != 2 || df.Index.Format != IndexFormat || len(df.Index.Heads) == 0 {
		return false
	}
	if df.Index.LogOffset < startPos {
		return false
	}

	df.entries = make(map[string]*MemoryEntry, len(df.Index.Heads))
	for id, head := range df.Index.Heads {
		df.entries[id] = &MemoryEntry{
			Offset:   head.BodyOffset,
			BodyLen:  head.BodyLen,
			Type:     head.Type,
			Topic:    head.Topic,
			Domain:   head.Domain,
			Severity: head.Severity,
			Oneliner: head.Oneliner,
			Created:  head.Created,
			Refs:     append([]string(nil), head.Refs...),
		}
	}
	if df.Index.State == nil {
		df.Index.State = &State{}
	}
	df.nextSeq = df.Index.AppliedSeq
	df.logOffset = df.Index.LogOffset
	return true
}

func (df *DoryFile) replayEvents(reader *bufio.Reader, startPos int64, minSeq uint64) error {
	pos := startPos
	seq := minSeq - 1

	for {
		lineStart := pos
		line, err := reader.ReadString('\n')
		if err == io.EOF && line == "" {
			break
		}
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed reading event delimiter: %w", err)
		}
		pos += int64(len(line))

		if strings.TrimSpace(line) == "" {
			if err == io.EOF {
				break
			}
			continue
		}
		if !isDelimiterLine(line) {
			return corruptionError(lineStart, "expected delimiter %q, got %q", EventDelim, strings.TrimRight(line, "\r\n"))
		}

		payloadOffset := pos
		payload, newPos, newReader, err := df.readEventPayload(reader, pos)
		if err != nil {
			return err
		}
		pos = newPos
		reader = newReader

		seq++
		if seq < minSeq {
			if err == io.EOF {
				break
			}
			continue
		}

		var ev logEvent
		if err := yaml.Unmarshal(payload, &ev); err != nil {
			return corruptionError(payloadOffset, "invalid event yaml: %v", err)
		}
		if err := df.applyEvent(seq, &ev, payloadOffset, len(payload)); err != nil {
			return corruptionError(payloadOffset, "%v", err)
		}

		if err == io.EOF {
			break
		}
	}

	df.logOffset = pos
	df.Index.AppliedSeq = df.nextSeq
	df.Index.LogOffset = df.logOffset
	return nil
}

func (df *DoryFile) readEventPayload(reader *bufio.Reader, startPos int64) ([]byte, int64, *bufio.Reader, error) {
	pos := startPos
	var content strings.Builder

	for {
		lineStart := pos
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			if len(line) > 0 {
				content.WriteString(line)
				pos += int64(len(line))
			}
			break
		}
		if err != nil {
			return nil, pos, reader, fmt.Errorf("failed reading event payload: %w", err)
		}
		if isDelimiterLine(line) {
			// Re-read this delimiter as the next record start.
			if _, err := df.knowledge.Seek(lineStart, 0); err != nil {
				return nil, pos, reader, fmt.Errorf("failed to seek delimiter position: %w", err)
			}
			payload := []byte(content.String())
			if len(strings.TrimSpace(string(payload))) == 0 {
				return nil, lineStart, reader, corruptionError(startPos, "empty event payload")
			}
			return payload, lineStart, bufio.NewReader(df.knowledge), nil
		}

		content.WriteString(line)
		pos += int64(len(line))
	}

	payload := []byte(content.String())
	if len(strings.TrimSpace(string(payload))) == 0 {
		return nil, pos, reader, corruptionError(startPos, "empty event payload")
	}
	return payload, pos, reader, nil
}

func isDelimiterLine(line string) bool {
	return strings.TrimRight(line, "\r\n") == EventDelim
}

// Close syncs and closes the knowledge file.
func (df *DoryFile) Close() error {
	if df.knowledge != nil {
		syncErr := df.knowledge.Sync() // Flush any buffered writes.
		closeErr := df.knowledge.Close()
		if syncErr != nil {
			return syncErr
		}
		return closeErr
	}
	return nil
}
