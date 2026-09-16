package outbox_test

import (
	"testing"
	"time"

	"github.com/trixtur/aprs-thursday/internal/aprs"
	"github.com/trixtur/aprs-thursday/internal/outbox"
)

func TestOutboxTracksPendingFailureAndSentAcrossRestart(t *testing.T) {
	path := t.TempDir() + "/outbox.json"
	m := aprs.Message{ID: "packet-1", From: "W1ABC", Text: "hello", Received: time.Now()}
	first, err := outbox.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	entry, added, err := first.Add(m, "/cards/packet-1.svg")
	if err != nil || !added || entry.Status != outbox.Pending {
		t.Fatalf("Add() = %#v, %v, %v", entry, added, err)
	}
	if err := first.MarkFailed(m.ID, "mailbox unavailable"); err != nil {
		t.Fatal(err)
	}
	second, err := outbox.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	entry, ok := second.Get(m.ID)
	if !ok || entry.Status != outbox.Failed || entry.LastError != "mailbox unavailable" {
		t.Fatalf("reloaded entry = %#v", entry)
	}
	if err := second.MarkSent(m.ID); err != nil {
		t.Fatal(err)
	}
	entry, ok = second.Get(m.ID)
	if !ok || entry.Status != outbox.Sent {
		t.Fatalf("sent entry = %#v", entry)
	}
}
