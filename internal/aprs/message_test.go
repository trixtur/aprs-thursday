package aprs_test

import (
	"testing"
	"time"

	"github.com/trixtur/aprs-thursday/internal/aprs"
)

func TestParseDirectMessage(t *testing.T) {
	when := time.Date(2026, 12, 24, 10, 0, 0, 0, time.UTC)
	got, err := aprs.ParseMessage("W1ABC>APRS,TCPIP*:N0CALL   :Hello there", when)
	if err != nil {
		t.Fatal(err)
	}
	if got.From != "W1ABC" || got.To != "N0CALL" || got.Text != "Hello there" || got.IsHOTGReply() || !got.IsFor("n0call") {
		t.Fatalf("parsed message = %#v", got)
	}
	if got.ID == "" || !got.Received.Equal(when) {
		t.Fatalf("message identity/time not set: %#v", got)
	}
}

func TestParseHOTGReply(t *testing.T) {
	got, err := aprs.ParseMessage("W1ABC>ANSRVR,TCPIP*:N0CALL   :HOTG: greetings from the net", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsHOTGReply() || got.Text != "greetings from the net" {
		t.Fatalf("parsed group message = %#v", got)
	}
}

func TestParseRejectsNonMessagePayload(t *testing.T) {
	for _, line := range []string{"W1ABC>APRS:>status", "not a packet", "W1ABC>APRS::short"} {
		if _, err := aprs.ParseMessage(line, time.Now()); err == nil {
			t.Errorf("ParseMessage(%q) error = nil", line)
		}
	}
}
