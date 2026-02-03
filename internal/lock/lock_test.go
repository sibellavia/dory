package lock

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAcquireRelease(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "write.lock")

	first, err := Acquire(lockPath, Options{
		Timeout:       500 * time.Millisecond,
		RetryInterval: 10 * time.Millisecond,
		StaleAfter:    time.Minute,
	})
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	defer first.Release()

	_, err = Acquire(lockPath, Options{
		Timeout:       100 * time.Millisecond,
		RetryInterval: 10 * time.Millisecond,
		StaleAfter:    time.Minute,
	})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got %v", err)
	}

	if err := first.Release(); err != nil {
		t.Fatalf("release failed: %v", err)
	}

	second, err := Acquire(lockPath, Options{
		Timeout:       200 * time.Millisecond,
		RetryInterval: 10 * time.Millisecond,
		StaleAfter:    time.Minute,
	})
	if err != nil {
		t.Fatalf("second acquire failed: %v", err)
	}
	if err := second.Release(); err != nil {
		t.Fatalf("second release failed: %v", err)
	}
}

func TestAcquireReclaimsStaleLock(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "write.lock")
	// Dead PID and very old timestamp => stale.
	if err := os.WriteFile(lockPath, []byte("pid: 999999\nhost: test\ntoken: stale\nacquired: 2020-01-01T00:00:00Z\n"), 0644); err != nil {
		t.Fatalf("write stale lock: %v", err)
	}
	old := time.Now().Add(-24 * time.Hour)
	if err := os.Chtimes(lockPath, old, old); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	l, err := Acquire(lockPath, Options{
		Timeout:       500 * time.Millisecond,
		RetryInterval: 10 * time.Millisecond,
		StaleAfter:    time.Minute,
	})
	if err != nil {
		t.Fatalf("acquire after stale failed: %v", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("release failed: %v", err)
	}
}
