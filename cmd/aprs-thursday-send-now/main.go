// Command aprs-thursday-send-now sends one configured weekly message on demand.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/trixtur/aprs-thursday/internal/aprsis"
	"github.com/trixtur/aprs-thursday/internal/config"
	"github.com/trixtur/aprs-thursday/internal/message"
	"github.com/trixtur/aprs-thursday/internal/rf"
	"github.com/trixtur/aprs-thursday/internal/schedule"
	"github.com/trixtur/aprs-thursday/internal/send"
)

func main() {
	confirmed := flag.Bool("confirm", false, "confirm one immediate APRS transmission")
	flag.Parse()
	if !*confirmed {
		fmt.Fprintln(os.Stderr, "refusing to send; rerun with --confirm")
		os.Exit(2)
	}
	cfg, err := config.FromEnv(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	regular, err := os.ReadFile(cfg.WeeklyMessage)
	if err != nil {
		log.Fatalf("read weekly message: %v", err)
	}
	hour, minute, location, err := cfg.Schedule()
	if err != nil {
		log.Fatal(err)
	}
	now := time.Now()
	occurrence, err := schedule.NextOccurrence(now, time.Thursday, hour, minute, location)
	if err != nil {
		log.Fatal(err)
	}
	if now.In(location).Weekday() == time.Thursday {
		occurrence = now.In(location)
	}
	body, err := message.Resolve(string(regular), cfg.MessageOverrides, occurrence)
	if err != nil {
		log.Fatalf("resolve weekly message: %v", err)
	}
	var publisher send.Publisher
	if cfg.TXMode == "rf" {
		publisher, err = rf.New(cfg.APRSCallsign, cfg.RFPath, cfg.RFDevice, cfg.RFDelay)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		client, err := aprsis.Dial(cfg.APRSISServer, cfg.APRSCallsign, cfg.APRSPasscode)
		if err != nil {
			log.Fatal(err)
		}
		defer client.Close()
		shared := aprsis.NewShared(cfg.APRSCallsign)
		shared.Set(client)
		publisher = shared
	}
	if err := publisher.Publish(body); err != nil {
		log.Fatal(err)
	}
	fmt.Println("sent one APRS Thursday message")
}
