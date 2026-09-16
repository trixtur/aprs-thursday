package receive_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/trixtur/aprs-thursday/internal/aprs"
	"github.com/trixtur/aprs-thursday/internal/card"
	"github.com/trixtur/aprs-thursday/internal/inbox"
	"github.com/trixtur/aprs-thursday/internal/outbox"
	"github.com/trixtur/aprs-thursday/internal/receive"
)

type fakeDelivery struct {
	paths    []string
	messages []aprs.Message
	err      error
}

func (f *fakeDelivery) Deliver(path string, message aprs.Message) error {
	f.paths = append(f.paths, path)
	f.messages = append(f.messages, message)
	return f.err
}

func TestPipelineRetriesDeliveryWithoutCreatingAnotherInboxRecord(t *testing.T) {
	m, err := aprs.ParseMessage("W1ABC>APRS,TCPIP*:N0CALL   :Hello again", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	delivery := &fakeDelivery{err: errors.New("temporary delivery failure")}
	pipeline := receive.Pipeline{Operator: "N0CALL", Inbox: mustInbox(t), Outbox: mustOutbox(t), Cards: card.Config{OperatorCallsign: "N0CALL", Greeting: "Greetings", OutputDir: t.TempDir()}, Delivery: delivery}
	if created, err := pipeline.Handle(m); created || err == nil {
		t.Fatalf("first Handle() = %v, %v; want failed delivery", created, err)
	}
	delivery.err = nil
	if created, err := pipeline.Handle(m); !created || err != nil {
		t.Fatalf("retry Handle() = %v, %v; want delivery success", created, err)
	}
	if len(delivery.paths) != 2 {
		t.Fatalf("delivery attempts = %d, want 2", len(delivery.paths))
	}
}

func TestPipelineDeliversOneCardForOneDirectMessage(t *testing.T) {
	when := time.Date(2026, 12, 24, 10, 0, 0, 0, time.UTC)
	m, err := aprs.ParseMessage("W1ABC>APRS,TCPIP*:N0CALL   :Hello there", when)
	if err != nil {
		t.Fatal(err)
	}
	delivery := &fakeDelivery{}
	pipeline := receive.Pipeline{Operator: "N0CALL", Inbox: mustInbox(t), Outbox: mustOutbox(t), Cards: card.Config{OperatorCallsign: "N0CALL", Greeting: "Greetings", OutputDir: t.TempDir()}, Delivery: delivery}
	for i := 0; i < 2; i++ {
		created, err := pipeline.Handle(m)
		if err != nil {
			t.Fatal(err)
		}
		if (i == 0) != created {
			t.Fatalf("call %d created = %v", i, created)
		}
	}
	if len(delivery.paths) != 1 || len(delivery.messages) != 1 {
		t.Fatalf("deliveries = %#v", delivery.paths)
	}
	if _, err := os.Stat(delivery.paths[0]); err != nil {
		t.Fatalf("card was not written: %v", err)
	}
}

func TestPipelineIgnoresHOTGReply(t *testing.T) {
	m, err := aprs.ParseMessage("W1ABC>ANSRVR,TCPIP*:N0CALL   :HOTG:Hello there", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	delivery := &fakeDelivery{}
	pipeline := receive.Pipeline{Operator: "N0CALL", Inbox: mustInbox(t), Outbox: mustOutbox(t), Cards: card.Config{OperatorCallsign: "N0CALL", Greeting: "Greetings", OutputDir: t.TempDir()}, Delivery: delivery}
	created, err := pipeline.Handle(m)
	if err != nil || created {
		t.Fatalf("Handle() = %v, %v; want false, nil", created, err)
	}
	if len(delivery.paths) != 0 {
		t.Fatalf("group reply was delivered: %#v", delivery.paths)
	}
}

func mustInbox(t *testing.T) *inbox.Store {
	t.Helper()
	s, err := inbox.Open(t.TempDir() + "/inbox.json")
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func mustOutbox(t *testing.T) *outbox.Store {
	t.Helper()
	s, err := outbox.Open(t.TempDir() + "/outbox.json")
	if err != nil {
		t.Fatal(err)
	}
	return s
}
