// Package aprsis contains the APRS-IS wire client. It has no policy about
// which packets to send or accept; those decisions remain in higher layers.
package aprsis

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

type Client struct {
	conn net.Conn
	read *bufio.Reader
	call string
}

func Dial(address, callsign, passcode string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("connect to APRS-IS: %w", err)
	}
	c := New(conn, callsign)
	if err := c.Login(passcode); err != nil {
		conn.Close()
		return nil, err
	}
	return c, nil
}

func New(conn net.Conn, callsign string) *Client {
	return &Client{conn: conn, read: bufio.NewReader(conn), call: callsign}
}

func (c *Client) Login(passcode string) error {
	if c == nil || c.conn == nil {
		return fmt.Errorf("APRS-IS connection is required")
	}
	if strings.TrimSpace(c.call) == "" || strings.TrimSpace(passcode) == "" {
		return fmt.Errorf("APRS-IS callsign and passcode are required")
	}
	if _, err := fmt.Fprintf(c.conn, "user %s pass %s vers aprs-thursday 0\r\n", c.call, passcode); err != nil {
		return fmt.Errorf("send APRS-IS login: %w", err)
	}
	line, err := c.read.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read APRS-IS login response: %w", err)
	}
	if strings.Contains(strings.ToLower(line), "unverified") || strings.Contains(strings.ToLower(line), "reject") {
		return fmt.Errorf("APRS-IS login rejected: %s", strings.TrimSpace(line))
	}
	return nil
}

func (c *Client) Send(line string) error {
	if c == nil || c.conn == nil {
		return fmt.Errorf("APRS-IS connection is required")
	}
	if strings.ContainsAny(line, "\r\n") || strings.TrimSpace(line) == "" {
		return fmt.Errorf("APRS packet must be one non-empty line")
	}
	_, err := fmt.Fprintf(c.conn, "%s\r\n", line)
	return err
}

func (c *Client) Receive() (string, error) {
	if c == nil || c.read == nil {
		return "", fmt.Errorf("APRS-IS connection is required")
	}
	line, err := c.read.ReadString('\n')
	return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r"), err
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func MessagePacket(from, to, body string) (string, error) {
	from = strings.TrimSpace(strings.ToUpper(from))
	to = strings.TrimSpace(strings.ToUpper(to))
	if from == "" || len(to) > 9 || to == "" {
		return "", fmt.Errorf("source and recipient callsigns are required")
	}
	if strings.ContainsAny(body, "\r\n") || body == "" {
		return "", fmt.Errorf("message body must be one non-empty line")
	}
	return fmt.Sprintf("%s>APRS,TCPIP*::%-9s:%s", from, to, body), nil
}

var _ io.Closer = (*Client)(nil)
