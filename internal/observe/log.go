// Package observe writes safe, structured events for the service and sandbox.
package observe

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

type Event struct {
	Time   time.Time      `json:"time"`
	Kind   string         `json:"kind"`
	Fields map[string]any `json:"fields,omitempty"`
}

type Logger struct {
	mu sync.Mutex
	w  io.Writer
}

func New(w io.Writer) *Logger { return &Logger{w: w} }

func (l *Logger) Event(kind string, fields map[string]any) error {
	if l == nil || l.w == nil {
		return fmt.Errorf("event logger writer is required")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return json.NewEncoder(l.w).Encode(Event{Time: time.Now().UTC(), Kind: kind, Fields: fields})
}
