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
	"github.com/trixtur/aprs-thursday/internal/schedule"
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
	client, err := aprsis.Dial(cfg.APRSISServer, cfg.APRSCallsign, cfg.APRSPasscode)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	publisher := aprsis.NewShared(cfg.APRSCallsign)
	publisher.Set(client)
	if err := publisher.Publish(body); err != nil {
		log.Fatal(err)
	}
	fmt.Println("sent one APRS Thursday message")
}
