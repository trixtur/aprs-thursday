package message_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/trixtur/aprs-thursday/internal/message"
)

func TestResolveUsesSpecialMessageForDesignatedPeriod(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2026, time.December, 24, 9, 0, 0, 0, time.UTC)
	writeMessage(t, filepath.Join(dir, "2026-12-24.txt"), "CQ HOTG Holiday greeting")

	got, err := message.Resolve("CQ HOTG Weekly greeting", dir, date)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if want := "CQ HOTG Holiday greeting"; got != want {
		t.Fatalf("Resolve() = %q, want %q", got, want)
	}

	got, err = message.Resolve("CQ HOTG Weekly greeting", dir, date.AddDate(0, 0, 7))
	if err != nil {
		t.Fatalf("Resolve() following week error = %v", err)
	}
	if want := "CQ HOTG Weekly greeting"; got != want {
		t.Fatalf("Resolve() following week = %q, want %q", got, want)
	}
}

func TestResolveUsesSpecialMessageThroughoutDesignatedWeek(t *testing.T) {
	dir := t.TempDir()
	thursday := time.Date(2026, time.December, 24, 9, 0, 0, 0, time.UTC)
	writeMessage(t, filepath.Join(dir, "2026-12-24.txt"), "CQ HOTG Holiday greeting")

	for _, day := range []int{0, 1, 6} {
		at := thursday.AddDate(0, 0, day)
		got, err := message.Resolve("CQ HOTG Weekly greeting", dir, at)
		if err != nil {
			t.Fatalf("Resolve() day %d error = %v", day, err)
		}
		if want := "CQ HOTG Holiday greeting"; got != want {
			t.Fatalf("Resolve() day %d = %q, want %q", day, got, want)
		}
	}
}

func TestResolveUsesRegularMessageWhenOverrideIsAbsent(t *testing.T) {
	got, err := message.Resolve("CQ HOTG Weekly greeting", t.TempDir(), time.Date(2026, time.January, 1, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if want := "CQ HOTG Weekly greeting"; got != want {
		t.Fatalf("Resolve() = %q, want %q", got, want)
	}
}

func TestResolveTrimsRegularMessageFileLineEnding(t *testing.T) {
	got, err := message.Resolve("CQ HOTG Weekly greeting\n", t.TempDir(), time.Date(2026, time.January, 1, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if want := "CQ HOTG Weekly greeting"; got != want {
		t.Fatalf("Resolve() = %q, want %q", got, want)
	}
}

func TestResolveRejectsInvalidOrOversizedMessages(t *testing.T) {
	tests := []struct {
		name    string
		regular string
	}{
		{name: "missing group prefix", regular: "Hello APRS Thursday"},
		{name: "too many bytes", regular: "CQ HOTG " + string(make([]byte, 60))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := message.Resolve(tt.regular, t.TempDir(), time.Now()); err == nil {
				t.Fatal("Resolve() error = nil, want validation error")
			}
		})
	}
}

func TestResolveRejectsInvalidSpecialOverride(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2026, time.January, 1, 9, 0, 0, 0, time.UTC)
	writeMessage(t, filepath.Join(dir, "2026-01-01.txt"), "not a group message")

	if _, err := message.Resolve("CQ HOTG Weekly greeting", dir, date); err == nil {
		t.Fatal("Resolve() error = nil, want validation error")
	}
}

func TestResolveDoesNotConsumeOverrideFile(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2026, time.January, 1, 9, 0, 0, 0, time.UTC)
	path := filepath.Join(dir, "2026-01-01.txt")
	writeMessage(t, path, "CQ HOTG Holiday greeting")

	if _, err := message.Resolve("CQ HOTG Weekly greeting", dir, date); err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("override should remain available for idempotent retries: %v", err)
	}
}

func writeMessage(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content+"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}
