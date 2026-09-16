package card_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/trixtur/aprs-thursday/internal/aprs"
	"github.com/trixtur/aprs-thursday/internal/card"
)

func TestRenderWritesEscapedQSLCard(t *testing.T) {
	message := aprs.Message{ID: "abc123", From: "W1ABC", Text: `<hello & APRS>`, Received: time.Date(2026, 12, 24, 10, 0, 0, 0, time.UTC)}
	path, err := card.Render(card.Config{OperatorCallsign: "N0CALL", Greeting: "Greetings from somewhere", Location: "Test location", PhotoURL: "photo.jpg", OutputDir: t.TempDir()}, message)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	contents := string(data)
	for _, want := range []string{"APRS MESSAGE RECEIPT", "W1ABC", "N0CALL", "MODE: APRS messaging", "BAND/FREQ: N/A", "&lt;hello &amp; APRS&gt;", "Greetings from somewhere", "photo.jpg"} {
		if !strings.Contains(contents, want) {
			t.Errorf("card does not contain %q: %s", want, contents)
		}
	}
}
