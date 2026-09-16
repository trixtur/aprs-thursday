// Package weekly coordinates one scheduled APRS Thursday publication.
package weekly

import (
	"context"
	"fmt"
	"time"

	"github.com/trixtur/aprs-thursday/internal/message"
	"github.com/trixtur/aprs-thursday/internal/send"
	"github.com/trixtur/aprs-thursday/internal/state"
)

// Attempt performs one scheduled send. The occurrence identifies the local
// scheduled Thursday; now is used for the UTC net-open check.
func Attempt(ctx context.Context, now, occurrence time.Time, regular, overridesDir string, publisher send.Publisher, sent *state.Store) (bool, error) {
	if ctx == nil {
		return false, fmt.Errorf("context is required")
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if occurrence.Weekday() != time.Thursday || now.UTC().Weekday() != time.Thursday {
		return false, nil
	}
	body, err := message.Resolve(regular, overridesDir, occurrence)
	if err != nil {
		return false, err
	}
	return send.Once(sent, publisher, occurrence, body)
}
