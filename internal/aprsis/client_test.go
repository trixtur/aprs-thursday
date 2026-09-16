package aprsis_test

import (
	"net"
	"testing"

	"github.com/trixtur/aprs-thursday/internal/aprsis"
)

func TestLoginSendAndReceive(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()
	client := aprsis.New(clientConn, "N0CALL-1")
	serverDone := make(chan error, 1)
	go func() {
		buf := make([]byte, 128)
		n, err := serverConn.Read(buf)
		if err != nil {
			serverDone <- err
			return
		}
		want := "user N0CALL-1 pass 12345 vers aprs-thursday 0\r\n"
		if string(buf[:n]) != want {
			serverDone <- errMismatch(string(buf[:n]), want)
			return
		}
		if _, err := serverConn.Write([]byte("# logresp N0CALL-1 verified, server T2TEST\r\n")); err != nil {
			serverDone <- err
			return
		}
		n, err = serverConn.Read(buf)
		if err != nil {
			serverDone <- err
			return
		}
		want = "N0CALL>APRS,TCPIP*::ANSRVR   :CQ HOTG test\r\n"
		if string(buf[:n]) != want {
			serverDone <- errMismatch(string(buf[:n]), want)
			return
		}
		_, err = serverConn.Write([]byte("W1ABC>APRS*:N0CALL   :hello\r\n"))
		serverDone <- err
	}()
	if err := client.Login("12345"); err != nil {
		t.Fatal(err)
	}
	packet, err := aprsis.MessagePacket("N0CALL", "ANSRVR", "CQ HOTG test")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Send(packet); err != nil {
		t.Fatal(err)
	}
	line, err := client.Receive()
	if err != nil {
		t.Fatal(err)
	}
	if line != "W1ABC>APRS*:N0CALL   :hello" {
		t.Fatalf("received %q", line)
	}
	if err := <-serverDone; err != nil {
		t.Fatal(err)
	}
}

func TestMessagePacketRejectsInvalidInput(t *testing.T) {
	for _, test := range [][3]string{{"", "ANSRVR", "body"}, {"N0CALL", "", "body"}, {"N0CALL", "ANSRVR", "line\nbreak"}} {
		if _, err := aprsis.MessagePacket(test[0], test[1], test[2]); err == nil {
			t.Errorf("MessagePacket(%q, %q, %q) error = nil", test[0], test[1], test[2])
		}
	}
}

type mismatch string

func (m mismatch) Error() string         { return string(m) }
func errMismatch(got, want string) error { return mismatch("got " + got + "; want " + want) }
