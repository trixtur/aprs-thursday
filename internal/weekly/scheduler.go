package weekly

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/trixtur/aprs-thursday/internal/schedule"
	"github.com/trixtur/aprs-thursday/internal/send"
	"github.com/trixtur/aprs-thursday/internal/state"
)

type Scheduler struct {
	Location     *time.Location
	Hour         int
	Minute       int
	Regular      string
	RegularPath  string
	OverridesDir string
	Publisher    send.Publisher
	State        *state.Store
	Now          func() time.Time
	Sleep        func(context.Context, time.Duration) error
}

// Run waits for each configured Thursday occurrence and attempts one send.
// Now and Sleep are injectable so tests never wait on wall-clock time.
func (s Scheduler) Run(ctx context.Context) error {
	if s.Location == nil || s.Publisher == nil || s.State == nil {
		return fmt.Errorf("scheduler location, publisher, and state are required")
	}
	if s.Now == nil {
		s.Now = time.Now
	}
	if s.Sleep == nil {
		s.Sleep = sleep
	}
	for {
		now := s.Now()
		occurrence, err := schedule.NextOccurrence(now, time.Thursday, s.Hour, s.Minute, s.Location)
		if err != nil {
			return err
		}
		if err := s.Sleep(ctx, time.Until(occurrence)); err != nil {
			return err
		}
		regular := s.Regular
		if s.RegularPath != "" {
			data, err := os.ReadFile(s.RegularPath)
			if err != nil {
				return fmt.Errorf("read weekly message: %w", err)
			}
			regular = string(data)
		}
		if _, err := Attempt(ctx, s.Now(), occurrence, regular, s.OverridesDir, s.Publisher, s.State); err != nil && err != context.Canceled {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
}

func sleep(ctx context.Context, duration time.Duration) error {
	if duration < 0 {
		duration = 0
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
