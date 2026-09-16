package observe_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/trixtur/aprs-thursday/internal/observe"
)

func TestLoggerWritesStructuredEvent(t *testing.T) {
	var output bytes.Buffer
	if err := observe.New(&output).Event("message.accepted", map[string]any{"sender": "W1ABC"}); err != nil {
		t.Fatal(err)
	}
	var event observe.Event
	if err := json.Unmarshal(output.Bytes(), &event); err != nil {
		t.Fatal(err)
	}
	if event.Kind != "message.accepted" || event.Fields["sender"] != "W1ABC" || event.Time.IsZero() {
		t.Fatalf("event = %#v", event)
	}
}
