package sandbox_test

import (
	"bufio"
	"net"
	"testing"
	"time"

	"github.com/trixtur/aprs-thursday/internal/sandbox"
)

func TestSandboxCapturesOutboundAndInjectsInbound(t *testing.T) {
	server, err := sandbox.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.Write([]byte("N0CALL>ANSRVR:CQ HOTG test\r\n")); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-server.Captured():
		if got != "N0CALL>ANSRVR:CQ HOTG test" {
			t.Fatalf("captured = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for captured packet")
	}
	if err := server.Inject("W1ABC>APRS*:N0CALL   :hello"); err != nil {
		t.Fatal(err)
	}
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if line != "W1ABC>APRS*:N0CALL   :hello\r\n" {
		t.Fatalf("injected = %q", line)
	}
}
