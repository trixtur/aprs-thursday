package sandbox_test

import (
	"bufio"
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"github.com/trixtur/aprs-thursday/internal/aprs"
	"github.com/trixtur/aprs-thursday/internal/card"
	"github.com/trixtur/aprs-thursday/internal/inbox"
	"github.com/trixtur/aprs-thursday/internal/observe"
	"github.com/trixtur/aprs-thursday/internal/outbox"
	"github.com/trixtur/aprs-thursday/internal/receive"
	"github.com/trixtur/aprs-thursday/internal/sandbox"
	"github.com/trixtur/aprs-thursday/internal/state"
	"github.com/trixtur/aprs-thursday/internal/weekly"
)

type delivery struct{ paths []string }

func (d *delivery) Deliver(path string, _ aprs.Message) error {
	d.paths = append(d.paths, path)
	return nil
}

type sandboxPublisher struct{ conn net.Conn }

func (p sandboxPublisher) Publish(body string) error {
	_, err := p.conn.Write([]byte("N0CALL>APRS,TCPIP*::ANSRVR   :" + body + "\r\n"))
	return err
}

func TestSandboxDynamicWeeklySend(t *testing.T) {
	server, err := sandbox.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	when := time.Date(2026, time.December, 24, 9, 0, 0, 0, time.UTC)
	stateStore, err := state.Open(t.TempDir() + "/state.json")
	if err != nil {
		t.Fatal(err)
	}
	attempted, err := weekly.Attempt(context.Background(), when, when, "CQ HOTG sandbox weekly", t.TempDir(), sandboxPublisher{conn}, stateStore)
	if err != nil || !attempted {
		t.Fatalf("Attempt() = %v, %v", attempted, err)
	}
	select {
	case packet := <-server.Captured():
		if packet != "N0CALL>APRS,TCPIP*::ANSRVR   :CQ HOTG sandbox weekly" {
			t.Fatalf("packet = %q", packet)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for weekly packet")
	}
}

func TestSandboxEndToEndDirectMessageFlow(t *testing.T) {
	server, err := sandbox.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	var events bytes.Buffer
	logger := observe.New(&events)
	weekly := "N0CALL>ANSRVR:CQ HOTG test"
	if _, err := conn.Write([]byte(weekly + "\r\n")); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-server.Captured():
		if got != weekly {
			t.Fatalf("captured weekly packet = %q", got)
		}
		if err := logger.Event("sandbox.outbound", map[string]any{"packet": got}); err != nil {
			t.Fatal(err)
		}
		t.Logf("captured outbound: %s", got)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for outbound packet")
	}

	delivery := &delivery{}
	pipeline := receive.Pipeline{Operator: "N0CALL", Inbox: mustInbox(t), Outbox: mustOutbox(t), Cards: card.Config{OperatorCallsign: "N0CALL", Greeting: "Greetings", Location: "Test location", OutputDir: t.TempDir()}, Delivery: delivery}
	if err := server.Inject("W1ABC>APRS,TCPIP*:N0CALL   :Hello from the sandbox"); err != nil {
		t.Fatal(err)
	}
	if err := server.Inject("W1ABC>ANSRVR,TCPIP*:N0CALL   :HOTG:group reply"); err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(conn)
	for i := 0; i < 2; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		message, err := aprs.ParseMessage(line, time.Now().UTC())
		if err != nil {
			t.Fatal(err)
		}
		created, err := pipeline.Handle(message)
		if err != nil {
			t.Fatal(err)
		}
		if message.IsHOTGReply() && created {
			t.Fatal("HOTG reply created a card")
		}
		if !message.IsHOTGReply() && !created {
			t.Fatal("direct message did not create a card")
		}
	}
	if len(delivery.paths) != 1 {
		t.Fatalf("delivery count = %d, want 1", len(delivery.paths))
	}
	t.Logf("generated card: %s", delivery.paths[0])
	t.Logf("event log: %s", events.String())
	if !bytes.Contains(events.Bytes(), []byte(`"kind":"sandbox.outbound"`)) {
		t.Fatalf("event log = %s", events.String())
	}
}

func mustInbox(t *testing.T) *inbox.Store {
	t.Helper()
	store, err := inbox.Open(t.TempDir() + "/inbox.json")
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func mustOutbox(t *testing.T) *outbox.Store {
	t.Helper()
	store, err := outbox.Open(t.TempDir() + "/outbox.json")
	if err != nil {
		t.Fatal(err)
	}
	return store
}
