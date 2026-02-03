package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sibellavia/dory/internal/doryfile"
	"github.com/sibellavia/dory/internal/lock"
)

// New creates a new Store instance.
func New(root string) *Store {
	if root == "" {
		root = DoryDir
	}
	return &Store{Root: root}
}

// Exists checks if the dory store exists.
func (s *Store) Exists() bool {
	df, err := doryfile.Open(s.Root)
	if err != nil {
		return false
	}
	df.Close()
	return true
}

// Init initializes the .dory directory with single file format.
func (s *Store) Init(project, description string) error {
	if s.Exists() {
		return fmt.Errorf("dory already initialized in %s", s.Root)
	}

	if err := ensureDir(s.Root); err != nil {
		return fmt.Errorf("failed to create root directory: %w", err)
	}

	df, err := doryfile.Create(s.Root, project, description)
	if err != nil {
		return fmt.Errorf("failed to create dory storage: %w", err)
	}
	defer df.Close()

	return nil
}

// open opens the dory storage if not already open.
func (s *Store) open() error {
	if s.df != nil {
		return nil
	}
	df, err := doryfile.Open(s.Root)
	if err != nil {
		return err
	}
	s.df = df
	return nil
}

// openLatest refreshes the open handle so reads see latest multi-agent writes.
func (s *Store) openLatest() error {
	if s.df != nil {
		if err := s.df.Close(); err != nil {
			return err
		}
		s.df = nil
	}
	return s.open()
}

// Close closes the dory file.
func (s *Store) Close() error {
	if s.df != nil {
		err := s.df.Close()
		s.df = nil
		return err
	}
	return nil
}

func (s *Store) withWriteLock(fn func() error) error {
	lockPath := filepath.Join(s.Root, writeLockFile)
	l, err := lock.Acquire(lockPath, lock.Options{
		Timeout:       writeLockTimeout,
		RetryInterval: writeLockRetry,
		StaleAfter:    writeLockStaleAge,
	})
	if err != nil {
		return err
	}
	defer l.Release()

	// Always reopen within the lock so writes are based on the latest log/index state.
	if s.df != nil {
		if err := s.df.Close(); err != nil {
			return err
		}
		s.df = nil
	}

	return fn()
}

// helper to ensure directory exists.
func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
