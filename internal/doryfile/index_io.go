package doryfile

import (
	"fmt"
	"os"

	"github.com/sibellavia/dory/internal/fileio"
	"gopkg.in/yaml.v3"
)

// loadIndex reads the index file.
func (df *DoryFile) loadIndex() error {
	data, err := os.ReadFile(df.IndexPath)
	if err != nil {
		return fmt.Errorf("failed to read index: %w", err)
	}

	var index Index
	if err := yaml.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("failed to parse index: %w", err)
	}

	if index.Format == "" {
		index.Format = IndexFormat // backwards compat for old files
	}
	if index.Format != IndexFormat {
		return fmt.Errorf("unsupported format %q (expected %q)", index.Format, IndexFormat)
	}
	if index.State == nil {
		index.State = &State{}
	}

	df.Index = &index
	df.nextSeq = df.Index.AppliedSeq
	df.logOffset = df.Index.LogOffset

	return nil
}

// saveIndex writes the index file.
func (df *DoryFile) saveIndex() error {
	if df.Index == nil {
		return fmt.Errorf("index is nil")
	}

	df.Index.Format = IndexFormat
	df.Index.AppliedSeq = df.nextSeq
	if df.logOffset <= 0 && df.knowledge != nil {
		if stat, err := df.knowledge.Stat(); err == nil {
			df.logOffset = stat.Size()
		}
	}
	df.Index.LogOffset = df.logOffset

	head := make(map[string]*SnapshotHead, len(df.entries))
	for id, entry := range df.entries {
		head[id] = &SnapshotHead{
			Type:         entry.Type,
			Topic:        entry.Topic,
			Domain:       entry.Domain,
			Severity:     entry.Severity,
			Oneliner:     entry.Oneliner,
			Created:      entry.Created,
			Refs:         append([]string(nil), entry.Refs...),
			BodyOffset:   entry.Offset,
			BodyLen:      entry.BodyLen,
			LastEventSeq: df.nextSeq,
		}
	}
	df.Index.Heads = head

	data, err := yaml.Marshal(df.Index)
	if err != nil {
		return err
	}
	return fileio.WriteFileAtomic(df.IndexPath, data, 0644)
}
