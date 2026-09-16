package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/trixtur/aprs-thursday/internal/aprs"
	"github.com/trixtur/aprs-thursday/internal/aprsis"
	"github.com/trixtur/aprs-thursday/internal/card"
	"github.com/trixtur/aprs-thursday/internal/config"
	"github.com/trixtur/aprs-thursday/internal/inbox"
	"github.com/trixtur/aprs-thursday/internal/observe"
	"github.com/trixtur/aprs-thursday/internal/outbox"
	"github.com/trixtur/aprs-thursday/internal/receive"
	"github.com/trixtur/aprs-thursday/internal/service"
	"github.com/trixtur/aprs-thursday/internal/state"
	"github.com/trixtur/aprs-thursday/internal/weekly"
)

type archiveDelivery struct{}

func (archiveDelivery) Deliver(string, aprs.Message) error { return nil }

func main() {
	cfg, err := config.FromEnv(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	if err := cfg.EnsureDirectories(); err != nil {
		log.Fatal(err)
	}
	inboxStore, err := inbox.Open(cfg.InboxPath)
	if err != nil {
		log.Fatal(err)
	}
	outboxStore, err := outbox.Open(cfg.OutboxPath)
	if err != nil {
		log.Fatal(err)
	}
	weeklyState, err := state.Open(filepath.Join(filepath.Dir(cfg.OutboxPath), "weekly-send.json"))
	if err != nil {
		log.Fatal(err)
	}
	logger := observe.New(os.Stderr)
	pipeline := &receive.Pipeline{Operator: cfg.OperatorCallsign, Inbox: inboxStore, Outbox: outboxStore, Cards: card.Config{OperatorCallsign: cfg.OperatorCallsign, Greeting: cfg.CardGreeting, Location: cfg.CardLocation, OutputDir: cfg.CardOutputDir}, Delivery: archiveDelivery{}}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	shared := aprsis.NewShared(cfg.APRSCallsign)
	factory := func() (service.LineSource, error) {
		client, err := aprsis.Dial(cfg.APRSISServer, cfg.APRSCallsign, cfg.APRSPasscode)
		if err != nil {
			return nil, err
		}
		shared.Set(client)
		return client, nil
	}
	hour, minute, location, err := cfg.Schedule()
	if err != nil {
		log.Fatal(err)
	}
	scheduler := weekly.Scheduler{Location: location, Hour: hour, Minute: minute, RegularPath: cfg.WeeklyMessage, OverridesDir: cfg.MessageOverrides, Publisher: shared, State: weeklyState}
	errs := make(chan error, 2)
	go func() {
		errs <- service.ReceiveWithReconnect(ctx, factory, cfg.OperatorCallsign, pipeline, logger, 5*time.Second)
	}()
	go func() { errs <- scheduler.Run(ctx) }()
	if err := <-errs; err != nil && err != context.Canceled {
		log.Fatal(err)
	}
	stop()
}
