package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/brianmargolis/shades/client"

	"github.com/brianmargolis/shades/protocol"

	"go.uber.org/zap"
)

const maxConnections = 100

type Server interface {
	Start(socketPath string) error
}

type server struct {
	currentTheme   atomic.Pointer[string]
	connections    atomic.Int32
	maxConnections int32
}

func NewServer() Server {
	return &server{maxConnections: maxConnections}
}

func (s *server) Start(socketPath string) error {
	logger := zap.S()

	// Remove any stale socket left by a crashed daemon; net.Listen("unix", ...)
	// fails if the path already exists even when nothing is listening.
	os.Remove(socketPath)

	if err := os.MkdirAll(filepath.Dir(socketPath), 0o700); err != nil {
		return fmt.Errorf("create socket dir: %w", err)
	}

	socket, err := net.Listen("unix", socketPath)
	if err != nil {
		logger.With("error", err).Error("Error listening on socket")
		panic(err)
	}
	defer os.Remove(socketPath)
	planForDeath(socket)

	return s.serve(socket)
}

// serve accepts connections until the listener is closed. Nothing a single
// client does should end this loop, since every other client's theming goes
// down with it.
func (s *server) serve(socket net.Listener) error {
	logger := zap.S()

	clients := []net.Conn{}
	clientMutex := sync.Mutex{}
	for {
		conn, err := socket.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			logger.Warnw("failed to accept connection", "error", err)
			continue
		}

		if s.connections.Load() >= s.maxConnections {
			logger.Warnw("too many connections, refusing a new one", "limit", s.maxConnections)
			conn.Close()
			continue
		}
		s.connections.Add(1)

		go s.talkToClient(conn, &clients, &clientMutex)
	}
}

func (s *server) talkToClient(
	conn net.Conn,
	clients *[]net.Conn,
	mutex *sync.Mutex,
) {
	logger := zap.S()

	logger.Debug("talkToClient: ", conn.RemoteAddr())
	defer func() {
		conn.Close()
		s.connections.Add(-1)
	}()

	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// short-lived CLI clients (toggle, set) intentionally disconnect after
				// proposing a theme — this is not an error
				logger.Debug("client disconnected")
			} else {
				logger.With("error", err).Warn("error reading from client - disconnecting")
			}
			unsubscribe(mutex, clients, conn)
			return
		}

		logger.With("message", msg).Debug("received message from client")
		verb, noun, err := protocol.Parse(msg)
		if err != nil {
			logger.Warnw("malformed message from client, skipping", "message", msg, "error", err)
			continue
		}

		switch verb {
		case "subscribe":
			subscribe(mutex, clients, conn)
		case "unsubscribe":
			unsubscribe(mutex, clients, conn)
		case "propose":
			// the client's own trailing newline has to come off here: protocol.Set
			// supplies the delimiter, so keeping this one would end every broadcast
			// with a blank line.
			proposedTheme := strings.TrimSpace(noun)
			s.currentTheme.Store(&proposedTheme)
			broadcast(mutex, clients, themeMessage(proposedTheme))
		case "get":
			theme := s.currentTheme.Load()
			if theme == nil {
				currentlyDark, err := isCurrentlyDark()
				if err != nil {
					continue
				}

				defaultLightTheme, defaultDarkTheme, err := getDefaults()
				if err != nil {
					continue
				}
				theme = &defaultLightTheme
				if currentlyDark {
					theme = &defaultDarkTheme
				}
			}

			whisper(mutex, conn, themeMessage(*theme))
		}
	}
}

// themeMessage renders everything a client needs to re-theme. The palette
// comes first so the colors are in hand by the time set names the theme, and
// both lines go out in one write so a concurrent propose cannot interleave
// somebody else's palette between this pair.
func themeMessage(themeAndVariant string) []byte {
	set := protocol.Set(themeAndVariant)

	palette, err := paletteMessage(themeAndVariant)
	if err != nil {
		zap.S().Warnw("could not resolve palette, sending set alone", "theme", themeAndVariant, "error", err)
		return set
	}

	return append(palette, set...)
}

// paletteMessage resolves a theme id into its colors. The daemon already reads
// shades.yaml, so resolving here saves every client from having to.
func paletteMessage(themeAndVariant string) ([]byte, error) {
	config, err := client.GetConfig()
	if err != nil {
		return nil, err
	}

	variant, err := config.Themes.GetVariant(themeAndVariant)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(variant.Colors)
	if err != nil {
		return nil, err
	}

	return protocol.Palette(string(payload)), nil
}

func subscribe(mutex *sync.Mutex, clients *[]net.Conn, conn net.Conn) {
	mutex.Lock()
	*clients = append(*clients, conn)
	mutex.Unlock()
}

func unsubscribe(mutex *sync.Mutex, clients *[]net.Conn, conn net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()
	for i, c := range *clients {
		if c == conn {
			*clients = append((*clients)[:i], (*clients)[i+1:]...)
			return
		}
	}
}

// broadcast sends a message to all subscribed clients, pruning any whose
// connections have gone dead since they last communicated.
func broadcast(mutex *sync.Mutex, clients *[]net.Conn, msg []byte) {
	logger := zap.S()

	mutex.Lock()
	defer mutex.Unlock()
	logger.Debugw("broadcasting message to clients", "nClients", len(*clients))

	active := make([]net.Conn, 0, len(*clients))
	for _, c := range *clients {
		_, err := c.Write(msg)
		if err != nil {
			logger.Warnw("client write failed during broadcast, removing", "error", err)
			c.Close()
		} else {
			active = append(active, c)
		}
	}
	*clients = active
}

// whisper sends a message to just one client
func whisper(mutex *sync.Mutex, conn net.Conn, msg []byte) {
	mutex.Lock()
	defer mutex.Unlock()
	_, err := conn.Write(msg)
	if err != nil {
		zap.S().With("error", err, "client", conn.RemoteAddr()).Warn("error writing to individual client (whispering")
	}
}

func isCurrentlyDark() (bool, error) {
	script := `tell application "System Events" to tell appearance preferences to get dark mode`

	output, err := client.RunApplescript(script)
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(output)) == "true", nil
}

func planForDeath(socket net.Listener) {
	// listen for SIGINT and SIGTERM signals, because we are a well behaved daemon.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		zap.L().Info("Received shutdown signal, cleaning up and exiting...")
		socket.Close()
		os.Exit(1)
	}()
}

func getDefaults() (string, string, error) {
	config, err := client.GetConfig()
	if err != nil {
		return "", "", err
	}
	return config.DefaultLightTheme, config.DefaultDarkTheme, nil
}
