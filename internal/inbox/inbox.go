// Package inbox persists received APRS messages and suppresses duplicates.
package inbox

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/trixtur/aprs-thursday/internal/aprs"
)

type Store struct {
	path     string
	mu       sync.Mutex
	messages map[string]Record
}

type Record struct {
	Message   aprs.Message `json:"message"`
	Delivered bool         `json:"delivered"`
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, messages: make(map[string]Record)}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read inbox: %w", err)
	}
	if err := json.Unmarshal(data, &s.messages); err != nil {
		return nil, fmt.Errorf("decode inbox: %w", err)
	}
	return s, nil
}

// Add saves a message and reports whether it was new. The file is replaced
// atomically so a restart cannot leave a partially written inbox.
func (s *Store) Add(message aprs.Message) (bool, error) {
	if s == nil || s.path == "" {
		return false, fmt.Errorf("inbox path is required")
	}
	if message.ID == "" {
		return false, fmt.Errorf("message ID is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.messages[message.ID]; exists {
		return false, nil
	}
	s.messages[message.ID] = Record{Message: message}
	if err := s.writeLocked(); err != nil {
		delete(s.messages, message.ID)
		return false, err
	}
	return true, nil
}

// Get returns the persisted record for an ID.
func (s *Store) Get(id string) (Record, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.messages[id]
	return record, ok
}

// MarkDelivered records successful card delivery.
func (s *Store) MarkDelivered(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.messages[id]
	if !ok {
		return fmt.Errorf("message %q is not in inbox", id)
	}
	if record.Delivered {
		return nil
	}
	record.Delivered = true
	s.messages[id] = record
	if err := s.writeLocked(); err != nil {
		record.Delivered = false
		s.messages[id] = record
		return err
	}
	return nil
}

func (s *Store) writeLocked() error {
	data, err := json.MarshalIndent(s.messages, "", "  ")
	if err != nil {
		return fmt.Errorf("encode inbox: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o750); err != nil {
		return fmt.Errorf("create inbox directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".inbox-*")
	if err != nil {
		return fmt.Errorf("stage inbox: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o640); err != nil {
		tmp.Close()
		return fmt.Errorf("protect inbox: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write inbox: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close inbox: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("install inbox: %w", err)
	}
	return nil
}
