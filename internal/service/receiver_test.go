package service_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/trixtur/aprs-thursday/internal/aprs"
	"github.com/trixtur/aprs-thursday/internal/card"
	"github.com/trixtur/aprs-thursday/internal/inbox"
	"github.com/trixtur/aprs-thursday/internal/observe"
	"github.com/trixtur/aprs-thursday/internal/outbox"
	"github.com/trixtur/aprs-thursday/internal/receive"
	"github.com/trixtur/aprs-thursday/internal/service"
)

type source struct {
	lines  []string
	closed bool
}

func (s *source) Receive() (string, error) {
	if len(s.lines) == 0 {
		return "", io.EOF
	}
	line := s.lines[0]
	s.lines = s.lines[1:]
	return line, nil
}
func (s *source) Close() error { s.closed = true; return nil }

type delivery struct{ count int }

func (d *delivery) Deliver(_ string, _ aprs.Message) error { d.count++; return nil }

func TestReceiveOnlyProcessesDirectMessagesAndLogsIgnoredGroup(t *testing.T) {
	stream := &source{lines: []string{
		"W1ABC>APRS,TCPIP*:N0CALL   :hello",
		"W1ABC>ANSRVR,TCPIP*:N0CALL   :HOTG:group reply",
	}}
	d := &delivery{}
	inboxStore, _ := inbox.Open(t.TempDir() + "/inbox.json")
	outboxStore, _ := outbox.Open(t.TempDir() + "/outbox.json")
	pipeline := &receive.Pipeline{Operator: "N0CALL", Inbox: inboxStore, Outbox: outboxStore, Cards: card.Config{OperatorCallsign: "N0CALL", Greeting: "Greetings", OutputDir: t.TempDir()}, Delivery: d}
	var events bytes.Buffer
	err := service.ReceiveOnly(context.Background(), stream, "N0CALL", pipeline, observe.New(&events))
	if err != io.EOF {
		t.Fatalf("ReceiveOnly() = %v, want EOF", err)
	}
	if d.count != 1 || !stream.closed {
		t.Fatalf("delivery count=%d closed=%v", d.count, stream.closed)
	}
	if !strings.Contains(events.String(), "aprs.message_ignored") {
		t.Fatalf("events = %s", events.String())
	}
}

func TestReceiveOnlyStopsOnCancellation(t *testing.T) {
	stream := &source{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	inboxStore, _ := inbox.Open(t.TempDir() + "/inbox.json")
	outboxStore, _ := outbox.Open(t.TempDir() + "/outbox.json")
	pipeline := &receive.Pipeline{Operator: "N0CALL", Inbox: inboxStore, Outbox: outboxStore, Cards: card.Config{OperatorCallsign: "N0CALL", Greeting: "Greetings", OutputDir: t.TempDir()}, Delivery: &delivery{}}
	err := service.ReceiveOnly(ctx, stream, "N0CALL", pipeline, nil)
	if err != context.Canceled || !stream.closed {
		t.Fatalf("ReceiveOnly() = %v, closed=%v", err, stream.closed)
	}
}
