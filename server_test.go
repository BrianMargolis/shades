package main

import (
	"bufio"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// startTestServer serves on a fresh socket and returns a function that dials
// it. The socket lives under os.MkdirTemp rather than t.TempDir because macOS
// caps unix socket paths at 104 bytes, and t.TempDir paths embed the test name.
func startTestServer(t *testing.T, s *server) func() net.Conn {
	t.Helper()

	// no config, so the server sends set lines without a palette ahead of them
	t.Setenv("SHADES_CONFIG", filepath.Join(t.TempDir(), "missing.yaml"))

	directory, err := os.MkdirTemp("", "shades")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(directory) })

	socketPath := filepath.Join(directory, "s.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })

	go s.serve(listener)

	return func() net.Conn {
		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.Close() })
		return conn
	}
}

func send(t *testing.T, conn net.Conn, message string) {
	t.Helper()
	if _, err := conn.Write([]byte(message)); err != nil {
		t.Fatal(err)
	}
}

func expectLine(t *testing.T, reader *bufio.Reader, conn net.Conn, want string) {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	got, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("waiting for %q: %v", want, err)
	}
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestServerSurvivesMalformedMessages(t *testing.T) {
	dial := startTestServer(t, &server{maxConnections: 10})

	subscriber := dial()
	subscriberReader := bufio.NewReader(subscriber)
	// messages on one connection are handled in order, so seeing this set
	// back means the subscribe has landed
	send(t, subscriber, "subscribe:test\npropose:everforest;dark-medium\n")
	expectLine(t, subscriberReader, subscriber, "set:everforest;dark-medium\n")

	proposer := dial()
	send(t, proposer, "propose\n\npropose:everforest;light-medium\n")
	expectLine(t, subscriberReader, subscriber, "set:everforest;light-medium\n")
}

func TestServerRefusesConnectionsOverTheLimit(t *testing.T) {
	dial := startTestServer(t, &server{maxConnections: 1})

	first := dial()
	firstReader := bufio.NewReader(first)
	send(t, first, "subscribe:test\npropose:everforest;dark-medium\n")
	expectLine(t, firstReader, first, "set:everforest;dark-medium\n")

	second := dial()
	second.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := bufio.NewReader(second).ReadString('\n'); err != io.EOF {
		t.Fatalf("expected the connection over the limit to be closed, got %v", err)
	}

	send(t, first, "propose:everforest;light-medium\n")
	expectLine(t, firstReader, first, "set:everforest;light-medium\n")
}
