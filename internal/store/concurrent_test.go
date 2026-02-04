package store

import (
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sibellavia/dory/internal/models"
)

// TestConcurrentStoreWrites tests multi-agent writes through the Store layer (with locking).
func TestConcurrentStoreWrites(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	s := New(root)
	if err := s.Init("concurrent-test", ""); err != nil {
		t.Fatal(err)
	}
	s.Close()

	const numWriters = 10
	const writesPerWriter = 20

	var wg sync.WaitGroup
	var successCount atomic.Int64
	var errorCount atomic.Int64
	errors := make(chan error, numWriters*writesPerWriter)

	start := time.Now()

	for w := 0; w < numWriters; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < writesPerWriter; i++ {
				// Each write uses a fresh store instance (simulating separate agent sessions)
				store := New(root)

				id, err := store.Learn(
					fmt.Sprintf("Lesson from worker %d iteration %d", workerID, i),
					"stress",
					models.SeverityNormal,
					fmt.Sprintf("Body content for worker %d iteration %d", workerID, i),
					nil,
				)

				if err != nil {
					errors <- fmt.Errorf("worker %d write %d: %w", workerID, i, err)
					errorCount.Add(1)
					store.Close()
					continue
				}

				if err := store.Close(); err != nil {
					errors <- fmt.Errorf("worker %d close %d: %w", workerID, i, err)
					errorCount.Add(1)
					continue
				}

				_ = id
				successCount.Add(1)
			}
		}(w)
	}

	wg.Wait()
	close(errors)
	elapsed := time.Since(start)

	// Collect errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	t.Logf("Concurrent store writes: %d success, %d errors in %v", successCount.Load(), errorCount.Load(), elapsed)
	t.Logf("Throughput: %.1f writes/sec", float64(successCount.Load())/elapsed.Seconds())

	if len(errs) > 0 {
		for _, err := range errs[:min(5, len(errs))] {
			t.Logf("  Error: %v", err)
		}
		if len(errs) > 5 {
			t.Logf("  ... and %d more errors", len(errs)-5)
		}
	}

	// Verify all successful writes are readable
	finalStore := New(root)
	items, err := finalStore.List("", "lesson", "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("Final list failed: %v", err)
	}
	finalStore.Close()

	if int64(len(items)) != successCount.Load() {
		t.Errorf("Entry count mismatch: got %d, expected %d", len(items), successCount.Load())
	} else {
		t.Logf("✓ Verified %d entries readable after concurrent writes", len(items))
	}

	// Check for duplicate IDs (would indicate race condition)
	ids := make(map[string]bool)
	for _, item := range items {
		if ids[item.ID] {
			t.Errorf("Duplicate ID found: %s", item.ID)
		}
		ids[item.ID] = true
	}
}

// TestConcurrentReadsWhileWriting tests that reads see consistent data during writes.
func TestConcurrentReadsWhileWriting(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".dory")
	s := New(root)
	if err := s.Init("read-write-test", ""); err != nil {
		t.Fatal(err)
	}
	// Seed with initial data
	for i := 0; i < 50; i++ {
		s.Learn(fmt.Sprintf("Seed lesson %d", i), "seed", models.SeverityNormal, "", nil)
	}
	s.Close()

	const numReaders = 5
	const numWriters = 3
	const duration = 2 * time.Second

	var wg sync.WaitGroup
	var readSuccess, writeSuccess atomic.Int64
	var readErrors, writeErrors atomic.Int64

	ctx := make(chan struct{})

	// Writers
	for w := 0; w < numWriters; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			i := 0
			for {
				select {
				case <-ctx:
					return
				default:
				}

				store := New(root)
				_, err := store.Learn(
					fmt.Sprintf("Writer %d lesson %d", id, i),
					"concurrent",
					models.SeverityNormal,
					"",
					nil,
				)
				store.Close()

				if err != nil {
					writeErrors.Add(1)
				} else {
					writeSuccess.Add(1)
				}
				i++
				time.Sleep(10 * time.Millisecond)
			}
		}(w)
	}

	// Readers
	for r := 0; r < numReaders; r++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-ctx:
					return
				default:
				}

				store := New(root)
				items, err := store.List("", "", "", time.Time{}, time.Time{})
				store.Close()

				if err != nil {
					readErrors.Add(1)
				} else if len(items) >= 50 { // At least seed data
					readSuccess.Add(1)
				}
				time.Sleep(5 * time.Millisecond)
			}
		}(r)
	}

	time.Sleep(duration)
	close(ctx)
	wg.Wait()

	t.Logf("Mixed workload over %v:", duration)
	t.Logf("  Writes: %d success, %d errors", writeSuccess.Load(), writeErrors.Load())
	t.Logf("  Reads: %d success, %d errors", readSuccess.Load(), readErrors.Load())

	if writeErrors.Load() > 0 {
		t.Errorf("Write errors occurred: %d", writeErrors.Load())
	}
	if readErrors.Load() > 0 {
		t.Errorf("Read errors occurred: %d", readErrors.Load())
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
