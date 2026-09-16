package inbox_test

import (
	"testing"
	"time"

	"github.com/trixtur/aprs-thursday/internal/aprs"
	"github.com/trixtur/aprs-thursday/internal/inbox"
)

func TestAddIsIdempotentAcrossRestart(t *testing.T) {
	path := t.TempDir() + "/inbox.json"
	m := aprs.Message{ID: "packet-1", From: "W1ABC", Text: "hello", Received: time.Now()}
	first, err := inbox.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	added, err := first.Add(m)
	if err != nil || !added {
		t.Fatalf("first Add() = %v, %v", added, err)
	}
	added, err = first.Add(m)
	if err != nil || added {
		t.Fatalf("second Add() = %v, %v", added, err)
	}
	second, err := inbox.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	added, err = second.Add(m)
	if err != nil || added {
		t.Fatalf("restart Add() = %v, %v", added, err)
	}
}
