package state_test

import (
	"testing"
	"time"

	"github.com/trixtur/aprs-thursday/internal/state"
)

func TestClaimIsIdempotentAcrossRestart(t *testing.T) {
	path := t.TempDir() + "/state.json"
	occurrence := time.Date(2026, time.January, 1, 9, 0, 0, 0, time.UTC)

	first, err := state.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := first.Claim(occurrence)
	if err != nil || !claimed {
		t.Fatalf("first Claim() = %v, %v; want true, nil", claimed, err)
	}
	claimed, err = first.Claim(occurrence)
	if err != nil || claimed {
		t.Fatalf("second Claim() = %v, %v; want false, nil", claimed, err)
	}

	second, err := state.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	claimed, err = second.Claim(occurrence)
	if err != nil || claimed {
		t.Fatalf("restart Claim() = %v, %v; want false, nil", claimed, err)
	}
}
