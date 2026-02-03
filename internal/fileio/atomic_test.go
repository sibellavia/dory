package fileio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomicCreatesAndOverwrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.yaml")

	if err := WriteFileAtomic(path, []byte("one"), 0644); err != nil {
		t.Fatalf("first WriteFileAtomic failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after first write: %v", err)
	}
	if string(got) != "one" {
		t.Fatalf("unexpected first content: %q", string(got))
	}

	if err := WriteFileAtomic(path, []byte("two"), 0644); err != nil {
		t.Fatalf("second WriteFileAtomic failed: %v", err)
	}

	got, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after second write: %v", err)
	}
	if string(got) != "two" {
		t.Fatalf("unexpected second content: %q", string(got))
	}
}
