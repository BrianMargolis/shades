package main

import (
	"bufio"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/brianmargolis/shades/client"
)

// startTestServer serves on a fresh socket and returns a function that dials
// it. The socket lives under os.MkdirTemp rather than t.TempDir because macOS
// caps unix socket paths at 104 bytes, and t.TempDir paths embed the test name.
func startTestServer(t *testing.T, s *server) func() net.Conn {
	t.Helper()

	directory, err := os.MkdirTemp("", "shades")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(directory) })
	useTestConfig(t, directory)

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

// useTestConfig points the config and state at directory, with a config that
// has one theme in two variants.
func useTestConfig(t *testing.T, directory string) {
	t.Helper()

	config := "themes:\n  everforest:\n    variants:\n      dark-medium:\n        light: false\n      light-medium:\n        light: true\n"
	configPath := filepath.Join(directory, "shades.yaml")
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHADES_CONFIG", configPath)
	t.Setenv("SHADES_STATE", filepath.Join(directory, "state.yaml"))
}

func send(t *testing.T, conn net.Conn, message string) {
	t.Helper()
	if _, err := conn.Write([]byte(message)); err != nil {
		t.Fatal(err)
	}
}

// expectLine waits for the next line that isn't a palette, since every set is
// preceded by one.
func expectLine(t *testing.T, reader *bufio.Reader, conn net.Conn, want string) {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for {
		got, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("waiting for %q: %v", want, err)
		}
		if strings.HasPrefix(got, "palette:") {
			continue
		}
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
		return
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

func TestServerRefusesThemesNotInTheConfig(t *testing.T) {
	dial := startTestServer(t, &server{maxConnections: 10})

	subscriber := dial()
	subscriberReader := bufio.NewReader(subscriber)
	send(t, subscriber, "subscribe:test\npropose:everforest;dark-medium\n")
	expectLine(t, subscriberReader, subscriber, "set:everforest;dark-medium\n")

	// the typo is dropped, so the next line the subscriber sees is the valid
	// proposal after it
	proposer := dial()
	send(t, proposer, "propose:everforst;dark-medium\npropose:everforest;light-medium\n")
	expectLine(t, subscriberReader, subscriber, "set:everforest;light-medium\n")

	getter := dial()
	getterReader := bufio.NewReader(getter)
	send(t, getter, "get:\n")
	expectLine(t, getterReader, getter, "set:everforest;light-medium\n")
}

func TestServerRestoresTheThemeAfterARestart(t *testing.T) {
	dial := startTestServer(t, &server{maxConnections: 10})

	conn := dial()
	reader := bufio.NewReader(conn)
	// the theme is saved before it's broadcast, so seeing the set means it's
	// on disk
	send(t, conn, "subscribe:test\npropose:everforest;light-medium\n")
	expectLine(t, reader, conn, "set:everforest;light-medium\n")

	restarted := &server{maxConnections: 10}
	restarted.restoreCurrentTheme()
	theme := restarted.currentTheme.Load()
	if theme == nil || *theme != "everforest;light-medium" {
		t.Fatalf("restored theme = %v, want everforest;light-medium", theme)
	}
}

func TestServerSkipsASavedThemeNoLongerInTheConfig(t *testing.T) {
	useTestConfig(t, t.TempDir())
	if err := client.SaveCurrentTheme("gruvbox;dark"); err != nil {
		t.Fatal(err)
	}

	restarted := &server{maxConnections: 10}
	restarted.restoreCurrentTheme()
	if theme := restarted.currentTheme.Load(); theme != nil {
		t.Fatalf("restored theme = %q, want none", *theme)
	}
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
