package weekly_test

import (
	"context"
	"testing"
	"time"

	"github.com/trixtur/aprs-thursday/internal/send"
	"github.com/trixtur/aprs-thursday/internal/state"
	"github.com/trixtur/aprs-thursday/internal/weekly"
)

type schedulerPublisher struct {
	bodies []string
	cancel context.CancelFunc
}

func (p *schedulerPublisher) Publish(body string) error {
	p.bodies = append(p.bodies, body)
	p.cancel()
	return nil
}

func TestSchedulerRunsAtNextThursdayWithoutWallClockWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	when := time.Date(2026, 12, 23, 12, 0, 0, 0, time.UTC)
	p := &schedulerPublisher{cancel: cancel}
	s, err := state.Open(t.TempDir() + "/state.json")
	if err != nil {
		t.Fatal(err)
	}
	clock := func() time.Time { return when }
	scheduler := weekly.Scheduler{Location: time.UTC, Hour: 9, Minute: 0, Regular: "CQ HOTG scheduled", OverridesDir: t.TempDir(), Publisher: p, State: s, Now: clock, Sleep: func(context.Context, time.Duration) error {
		when = time.Date(2026, 12, 24, 9, 0, 0, 0, time.UTC)
		return nil
	}}
	err = scheduler.Run(ctx)
	if err != context.Canceled {
		t.Fatalf("Run() = %v", err)
	}
	if len(p.bodies) != 1 || p.bodies[0] != "CQ HOTG scheduled" {
		t.Fatalf("bodies = %#v", p.bodies)
	}
}

var _ send.Publisher = (*schedulerPublisher)(nil)
