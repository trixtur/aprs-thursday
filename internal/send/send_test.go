package send_test

import (
	"errors"
	"testing"
	"time"

	"github.com/trixtur/aprs-thursday/internal/send"
	"github.com/trixtur/aprs-thursday/internal/state"
)

type fakePublisher struct {
	bodies []string
	err    error
}

func (f *fakePublisher) Publish(body string) error {
	f.bodies = append(f.bodies, body)
	return f.err
}

func TestOncePublishesOnlyTheFirstTime(t *testing.T) {
	store, err := state.Open(t.TempDir() + "/state.json")
	if err != nil {
		t.Fatal(err)
	}
	publisher := &fakePublisher{}
	when := time.Date(2026, time.January, 1, 9, 0, 0, 0, time.UTC)

	for i := 0; i < 2; i++ {
		claimed, err := send.Once(store, publisher, when, "CQ HOTG test")
		if err != nil {
			t.Fatal(err)
		}
		if (i == 0) != claimed {
			t.Fatalf("call %d claimed = %v", i, claimed)
		}
	}
	if len(publisher.bodies) != 1 || publisher.bodies[0] != "CQ HOTG test" {
		t.Fatalf("published bodies = %#v", publisher.bodies)
	}
}

func TestOnceClaimsBeforeAnUncertainPublishFailure(t *testing.T) {
	store, err := state.Open(t.TempDir() + "/state.json")
	if err != nil {
		t.Fatal(err)
	}
	publisher := &fakePublisher{err: errors.New("connection lost")}
	when := time.Date(2026, time.January, 1, 9, 0, 0, 0, time.UTC)
	claimed, err := send.Once(store, publisher, when, "CQ HOTG test")
	if !claimed || err == nil {
		t.Fatalf("Once() = %v, %v; want claimed true and error", claimed, err)
	}

	publisher.err = nil
	claimed, err = send.Once(store, publisher, when, "CQ HOTG test")
	if err != nil || claimed {
		t.Fatalf("retry Once() = %v, %v; want false, nil", claimed, err)
	}
}
