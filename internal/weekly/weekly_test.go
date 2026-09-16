package weekly_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/trixtur/aprs-thursday/internal/send"
	"github.com/trixtur/aprs-thursday/internal/state"
	"github.com/trixtur/aprs-thursday/internal/weekly"
)

type publisher struct{ bodies []string }

func (p *publisher) Publish(body string) error { p.bodies = append(p.bodies, body); return nil }

func TestAttemptSendsRegularMessageOnceOnThursday(t *testing.T) {
	now := time.Date(2026, 12, 24, 9, 0, 0, 0, time.UTC)
	p := &publisher{}
	s, _ := state.Open(t.TempDir() + "/state.json")
	for i := 0; i < 2; i++ {
		attempted, err := weekly.Attempt(context.Background(), now, now, "CQ HOTG weekly", t.TempDir(), p, s)
		if err != nil {
			t.Fatal(err)
		}
		if (i == 0) != attempted {
			t.Fatalf("attempt %d = %v", i, attempted)
		}
	}
	if len(p.bodies) != 1 || p.bodies[0] != "CQ HOTG weekly" {
		t.Fatalf("bodies = %#v", p.bodies)
	}
}

func TestAttemptUsesSpecialMessageForTheOccurrence(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "2026-12-24.txt"), []byte("CQ HOTG holiday\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 12, 24, 9, 0, 0, 0, time.UTC)
	p := &publisher{}
	s, _ := state.Open(t.TempDir() + "/state.json")
	if _, err := weekly.Attempt(context.Background(), now, now, "CQ HOTG weekly", dir, p, s); err != nil {
		t.Fatal(err)
	}
	if len(p.bodies) != 1 || p.bodies[0] != "CQ HOTG holiday" {
		t.Fatalf("bodies = %#v", p.bodies)
	}
}

func TestAttemptDoesNotSendOutsideThursdayNetWindow(t *testing.T) {
	friday := time.Date(2026, 12, 25, 9, 0, 0, 0, time.UTC)
	p := &publisher{}
	s, _ := state.Open(t.TempDir() + "/state.json")
	attempted, err := weekly.Attempt(context.Background(), friday, friday, "CQ HOTG weekly", t.TempDir(), p, s)
	if err != nil {
		t.Fatal(err)
	}
	if attempted || len(p.bodies) != 0 {
		t.Fatalf("attempted=%v bodies=%v", attempted, p.bodies)
	}
}

var _ send.Publisher = (*publisher)(nil)
