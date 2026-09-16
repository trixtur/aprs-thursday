package aprsis

import (
	"fmt"
	"sync"
)

// Shared serializes access to the single APRS-IS identity used by receive and
// scheduled-send paths. A reconnect replaces the active client.
type Shared struct {
	mu     sync.Mutex
	client *Client
	call   string
}

func NewShared(callsign string) *Shared { return &Shared{call: callsign} }

func (s *Shared) Set(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client = client
}

func (s *Shared) Publish(body string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return fmt.Errorf("APRS-IS connection is not ready")
	}
	packet, err := MessagePacket(s.call, "ANSRVR", body)
	if err != nil {
		return err
	}
	return s.client.Send(packet)
}
