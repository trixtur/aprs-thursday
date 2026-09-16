// Package state stores scheduled sends that have already been claimed.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Store records an occurrence before it is sent. Claim returns true only the
// first time an occurrence is seen, including after a process restart.
type Store struct {
	path   string
	mu     sync.Mutex
	claims map[string]time.Time
}

// Open loads a store, creating an empty one when path does not exist.
func Open(path string) (*Store, error) {
	s := &Store{path: path, claims: make(map[string]time.Time)}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read send state: %w", err)
	}
	if err := json.Unmarshal(data, &s.claims); err != nil {
		return nil, fmt.Errorf("decode send state: %w", err)
	}
	return s, nil
}

// Claim records occurrence and reports whether it was newly claimed.
func (s *Store) Claim(occurrence time.Time) (bool, error) {
	if s == nil || s.path == "" {
		return false, fmt.Errorf("send state path is required")
	}
	key := occurrence.Format(time.RFC3339)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.claims[key]; exists {
		return false, nil
	}
	s.claims[key] = occurrence
	data, err := json.MarshalIndent(s.claims, "", "  ")
	if err != nil {
		return false, fmt.Errorf("encode send state: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o750); err != nil {
		return false, fmt.Errorf("create send state directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".send-state-*")
	if err != nil {
		return false, fmt.Errorf("stage send state: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o640); err != nil {
		tmp.Close()
		return false, fmt.Errorf("protect send state: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return false, fmt.Errorf("write send state: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return false, fmt.Errorf("close send state: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return false, fmt.Errorf("install send state: %w", err)
	}
	return true, nil
}
