package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/trixtur/aprs-thursday/internal/outbox"
)

func main() {
	defaultPath := filepath.Join("/var/lib", "aprs-thursday", "outbox.json")
	path := flag.String("outbox", defaultPath, "path to the APRS Thursday outbox JSON file")
	flag.Parse()

	store, err := outbox.Open(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	entries := store.Pending()
	if len(entries) == 0 {
		fmt.Println("No pending or failed QSL cards.")
		return
	}
	fmt.Printf("%d QSL card(s) need attention:\n\n", len(entries))
	for _, entry := range entries {
		fmt.Printf("From:     %s\n", entry.Message.From)
		fmt.Printf("Received: %s\n", entry.Message.Received.UTC().Format("2006-01-02 15:04:05 UTC"))
		fmt.Printf("Status:   %s\n", entry.Status)
		fmt.Printf("Card:     %s\n", entry.CardPath)
		if entry.LastError != "" {
			fmt.Printf("Error:    %s\n", entry.LastError)
		}
		fmt.Println()
	}
}
