// Package send coordinates one weekly publication.
package send

import (
	"fmt"
	"time"

	"github.com/trixtur/aprs-thursday/internal/state"
)

// Publisher is the boundary implemented by the APRS-IS client.
type Publisher interface {
	Publish(body string) error
}

// Once claims occurrence and publishes body. A claimed occurrence is skipped,
// including after a process restart, to avoid duplicate APRS packets.
func Once(store *state.Store, publisher Publisher, occurrence time.Time, body string) (bool, error) {
	if store == nil {
		return false, fmt.Errorf("send state is required")
	}
	if publisher == nil {
		return false, fmt.Errorf("publisher is required")
	}
	claimed, err := store.Claim(occurrence)
	if err != nil {
		return false, err
	}
	if !claimed {
		return false, nil
	}
	if err := publisher.Publish(body); err != nil {
		return true, fmt.Errorf("publish weekly message: %w", err)
	}
	return true, nil
}
