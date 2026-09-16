package receive_test

import (
	"testing"
	"time"

	"github.com/trixtur/aprs-thursday/internal/aprs"
	"github.com/trixtur/aprs-thursday/internal/receive"
)

func TestDirectForOperatorAcceptsDirectMessage(t *testing.T) {
	m, err := aprs.ParseMessage("W1ABC>APRS*:N0CALL   :Hello", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !receive.DirectForOperator(m, "N0CALL") {
		t.Fatal("direct message was rejected")
	}
}

func TestDirectForOperatorRejectsHOTGReply(t *testing.T) {
	m, err := aprs.ParseMessage("W1ABC>ANSRVR,TCPIP*:N0CALL   :HOTG:Hello", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if receive.DirectForOperator(m, "N0CALL") {
		t.Fatal("HOTG reply was accepted as a direct message")
	}
}

func TestDirectForOperatorRejectsMessageForAnotherStation(t *testing.T) {
	m, err := aprs.ParseMessage("W1ABC>APRS*:OTHER    :Hello", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if receive.DirectForOperator(m, "N0CALL") {
		t.Fatal("message for another station was accepted")
	}
}
