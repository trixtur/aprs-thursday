package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/trixtur/aprs-thursday/internal/aprs"
	"github.com/trixtur/aprs-thursday/internal/card"
	"github.com/trixtur/aprs-thursday/internal/inbox"
	"github.com/trixtur/aprs-thursday/internal/outbox"
	"github.com/trixtur/aprs-thursday/internal/receive"
	"github.com/trixtur/aprs-thursday/internal/service"
)

type reconnectSource struct {
	err    error
	closed bool
}

func (s *reconnectSource) Receive() (string, error) { return "", s.err }
func (s *reconnectSource) Close() error             { s.closed = true; return nil }

type reconnectDelivery struct{}

func (reconnectDelivery) Deliver(string, aprs.Message) error { return nil }

func TestReceiveWithReconnectRetriesFactoryAndStopsWhenCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	attempts := 0
	first := &reconnectSource{err: errors.New("connection dropped")}
	factory := func() (service.LineSource, error) {
		attempts++
		if attempts == 1 {
			return first, nil
		}
		cancel()
		return nil, errors.New("stopping test")
	}
	inboxStore, _ := inbox.Open(t.TempDir() + "/inbox.json")
	outboxStore, _ := outbox.Open(t.TempDir() + "/outbox.json")
	pipeline := &receive.Pipeline{Operator: "N0CALL", Inbox: inboxStore, Outbox: outboxStore, Cards: card.Config{OperatorCallsign: "N0CALL", Greeting: "Greetings", OutputDir: t.TempDir()}, Delivery: reconnectDelivery{}}
	err := service.ReceiveWithReconnect(ctx, factory, "N0CALL", pipeline, nil, time.Millisecond)
	if err != context.Canceled || attempts != 2 || !first.closed {
		t.Fatalf("result=%v attempts=%d closed=%v", err, attempts, first.closed)
	}
}
