package aprsis_test

import (
	"net"
	"testing"

	"github.com/trixtur/aprs-thursday/internal/aprsis"
)

func TestSharedPublishesOnCurrentClient(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()
	shared := aprsis.NewShared("N0CALL")
	shared.Set(aprsis.New(clientConn, "N0CALL"))
	read := make(chan string, 1)
	go func() { buf := make([]byte, 128); n, _ := serverConn.Read(buf); read <- string(buf[:n]) }()
	if err := shared.Publish("CQ HOTG test"); err != nil {
		t.Fatal(err)
	}
	if got := <-read; got != "N0CALL>APRS,TCPIP*::ANSRVR   :CQ HOTG test\r\n" {
		t.Fatalf("packet = %q", got)
	}
}
