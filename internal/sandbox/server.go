// Package sandbox provides a temporary, local APRS-IS-like line server.
package sandbox

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type Server struct {
	listener net.Listener
	mu       sync.Mutex
	conn     net.Conn
	captured chan string
	closed   chan struct{}
}

func Start() (*Server, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("start APRS sandbox: %w", err)
	}
	s := &Server{listener: listener, captured: make(chan string, 32), closed: make(chan struct{})}
	go s.accept()
	return s, nil
}

func (s *Server) Addr() string { return s.listener.Addr().String() }

func (s *Server) Captured() <-chan string { return s.captured }

func (s *Server) accept() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			close(s.closed)
			close(s.captured)
			return
		}
		s.mu.Lock()
		s.conn = conn
		s.mu.Unlock()
		go s.read(conn)
	}
}

func (s *Server) read(conn net.Conn) {
	defer conn.Close()
	for scanner := bufio.NewScanner(conn); scanner.Scan(); {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		select {
		case s.captured <- line:
		case <-s.closed:
			return
		}
	}
}

// Inject sends a packet to the connected test client.
func (s *Server) Inject(packet string) error {
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()
	if conn == nil {
		return fmt.Errorf("sandbox has no connected client")
	}
	_, err := fmt.Fprintf(conn, "%s\r\n", packet)
	return err
}

func (s *Server) Close() error {
	err := s.listener.Close()
	s.mu.Lock()
	if s.conn != nil {
		s.conn.Close()
	}
	s.mu.Unlock()
	return err
}
