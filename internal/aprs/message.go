// Package aprs contains the small APRS message parser used by the receiver.
package aprs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type Message struct {
	ID        string
	From      string
	To        string
	Text      string
	Received  time.Time
	Group     string
	RawPacket string
}

// ParseMessage parses a TNC2 APRS-IS line containing an APRS message payload.
func ParseMessage(line string, received time.Time) (Message, error) {
	line = strings.TrimSpace(line)
	withoutComment := strings.SplitN(line, " {", 2)[0]
	headerEnd := strings.IndexByte(withoutComment, ':')
	if headerEnd < 0 {
		return Message{}, fmt.Errorf("packet has no payload")
	}
	header, payload := withoutComment[:headerEnd], withoutComment[headerEnd:]
	sourceEnd := strings.IndexByte(header, '>')
	if sourceEnd <= 0 {
		return Message{}, fmt.Errorf("packet has no source and destination")
	}
	from := strings.ToUpper(strings.TrimSpace(header[:sourceEnd]))
	if from == "" {
		return Message{}, fmt.Errorf("packet source is empty")
	}
	destination := strings.ToUpper(strings.SplitN(header[sourceEnd+1:], ",", 2)[0])
	if len(payload) < 11 || payload[0] != ':' || payload[10] != ':' {
		return Message{}, fmt.Errorf("payload is not an APRS message")
	}
	to := strings.TrimSpace(strings.ToUpper(payload[1:10]))
	text := strings.TrimSpace(payload[11:])
	if to == "" || text == "" {
		return Message{}, fmt.Errorf("APRS message recipient or text is empty")
	}
	group := ""
	if destination == "ANSRVR" {
		upper := strings.ToUpper(text)
		if strings.HasPrefix(upper, "HOTG:") {
			group = "HOTG"
			text = strings.TrimSpace(text[len("HOTG:"):])
		}
	}
	hash := sha256.Sum256([]byte(line))
	return Message{ID: hex.EncodeToString(hash[:]), From: from, To: to, Text: text, Received: received, Group: group, RawPacket: line}, nil
}

func (m Message) IsFor(operator string) bool {
	return strings.EqualFold(strings.TrimSpace(m.To), strings.TrimSpace(operator))
}

func (m Message) IsHOTGReply() bool { return strings.EqualFold(m.Group, "HOTG") }
