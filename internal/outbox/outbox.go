// Package outbox tracks QSL cards awaiting operator delivery.
package outbox

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/trixtur/aprs-thursday/internal/aprs"
)

type Status string

const (
	Pending Status = "pending"
	Sent    Status = "sent"
	Failed  Status = "failed"
)

type Entry struct {
	Message   aprs.Message `json:"message"`
	CardPath  string       `json:"card_path"`
	Status    Status       `json:"status"`
	LastError string       `json:"last_error,omitempty"`
	Updated   time.Time    `json:"updated"`
}

type Store struct {
	path    string
	mu      sync.Mutex
	entries map[string]Entry
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, entries: make(map[string]Entry)}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read outbox: %w", err)
	}
	if err := json.Unmarshal(data, &s.entries); err != nil {
		return nil, fmt.Errorf("decode outbox: %w", err)
	}
	return s, nil
}

func (s *Store) Add(message aprs.Message, cardPath string) (Entry, bool, error) {
	if s == nil || s.path == "" {
		return Entry{}, false, fmt.Errorf("outbox path is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry, ok := s.entries[message.ID]; ok {
		return entry, false, nil
	}
	entry := Entry{Message: message, CardPath: cardPath, Status: Pending, Updated: time.Now().UTC()}
	s.entries[message.ID] = entry
	if err := s.writeLocked(); err != nil {
		delete(s.entries, message.ID)
		return Entry{}, false, err
	}
	return entry, true, nil
}

func (s *Store) Get(id string) (Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[id]
	return entry, ok
}

func (s *Store) MarkSent(id string) error { return s.mark(id, Sent, "") }

func (s *Store) MarkFailed(id, reason string) error { return s.mark(id, Failed, reason) }

func (s *Store) Pending() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Entry, 0)
	for _, entry := range s.entries {
		if entry.Status == Pending || entry.Status == Failed {
			result = append(result, entry)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Updated.Before(result[j].Updated) })
	return result
}

func (s *Store) mark(id string, status Status, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[id]
	if !ok {
		return fmt.Errorf("outbox entry %q not found", id)
	}
	entry.Status, entry.LastError, entry.Updated = status, reason, time.Now().UTC()
	s.entries[id] = entry
	return s.writeLocked()
}

func (s *Store) writeLocked() error {
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("encode outbox: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o750); err != nil {
		return fmt.Errorf("create outbox directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".outbox-*")
	if err != nil {
		return fmt.Errorf("stage outbox: %w", err)
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err := tmp.Chmod(0o640); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, s.path)
}
