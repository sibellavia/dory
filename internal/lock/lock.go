package lock

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

var ErrTimeout = errors.New("timed out waiting for write lock")

type Options struct {
	Timeout       time.Duration
	RetryInterval time.Duration
	StaleAfter    time.Duration
}

type FileLock struct {
	path  string
	token string
}

type metadata struct {
	PID      int    `yaml:"pid"`
	Host     string `yaml:"host"`
	Token    string `yaml:"token"`
	Acquired string `yaml:"acquired"`
}

func Acquire(path string, opts Options) (*FileLock, error) {
	if opts.Timeout <= 0 {
		opts.Timeout = 5 * time.Second
	}
	if opts.RetryInterval <= 0 {
		opts.RetryInterval = 50 * time.Millisecond
	}
	if opts.StaleAfter <= 0 {
		opts.StaleAfter = 10 * time.Minute
	}

	token, err := newToken()
	if err != nil {
		return nil, err
	}

	host, _ := os.Hostname()
	meta := metadata{
		PID:      os.Getpid(),
		Host:     host,
		Token:    token,
		Acquired: time.Now().UTC().Format(time.RFC3339Nano),
	}

	deadline := time.Now().Add(opts.Timeout)
	for {
		created, err := tryCreate(path, meta)
		if err == nil && created {
			return &FileLock{path: path, token: token}, nil
		}
		if err != nil && !errors.Is(err, fs.ErrExist) {
			return nil, fmt.Errorf("failed to acquire lock: %w", err)
		}

		stale, owner := isStale(path, opts.StaleAfter)
		if stale {
			_ = os.Remove(path)
			continue
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("%w at %s (owner pid=%d host=%s acquired=%s)",
				ErrTimeout, path, owner.PID, owner.Host, owner.Acquired)
		}
		time.Sleep(opts.RetryInterval)
	}
}

func (l *FileLock) Release() error {
	if l == nil || l.path == "" {
		return nil
	}

	data, err := os.ReadFile(l.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed reading lock file: %w", err)
	}

	var meta metadata
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("failed parsing lock file: %w", err)
	}
	if meta.Token != l.token {
		return fmt.Errorf("refusing to release lock not owned by this process")
	}

	if err := os.Remove(l.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed removing lock file: %w", err)
	}
	return nil
}

func tryCreate(path string, meta metadata) (bool, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return false, err
	}

	data, err := yaml.Marshal(meta)
	if err != nil {
		f.Close()
		_ = os.Remove(path)
		return false, err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		_ = os.Remove(path)
		return false, err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		_ = os.Remove(path)
		return false, err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return false, err
	}
	return true, nil
}

func isStale(path string, staleAfter time.Duration) (bool, metadata) {
	var owner metadata
	info, err := os.Stat(path)
	if err != nil {
		return false, owner
	}

	data, err := os.ReadFile(path)
	if err == nil {
		_ = yaml.Unmarshal(data, &owner)
	}

	if owner.PID > 0 && !isProcessAlive(owner.PID) {
		return true, owner
	}

	// If metadata is unreadable/missing and file is old, treat as stale.
	if owner.PID == 0 && time.Since(info.ModTime()) > staleAfter {
		return true, owner
	}

	return false, owner
}

func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = p.Signal(syscall.Signal(0))
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrProcessDone) {
		return false
	}

	// EPERM means the process exists but we lack permission.
	var errno syscall.Errno
	if errors.As(err, &errno) && errno == syscall.EPERM {
		return true
	}

	// Windows reports unsupported signal; avoid false stale detection.
	if runtime.GOOS == "windows" && strings.Contains(strings.ToLower(err.Error()), "not supported") {
		return true
	}

	return false
}

func newToken() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("failed generating lock token: %w", err)
	}
	return hex.EncodeToString(buf[:]), nil
}
