package main

import (
	"brianmargolis/shades/client"
	"brianmargolis/shades/protocol"
	"bufio"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"go.uber.org/zap"
)

const MAX_CLIENTS = 100

type Server interface {
	Start(socketPath string) error
}

type server struct {
	currentTheme atomic.Pointer[string]
}

func NewServer() Server {
	return &server{}
}

func (s *server) Start(socketPath string) error {
	logger := zap.S()
	socket, err := net.Listen("unix", socketPath)
	if err != nil {
		logger.With("error", err).Error("Error listening on socket")
		panic(err)
	}
	planForDeath(socket)

	clients := []net.Conn{}
	clientMutex := sync.Mutex{}
	for {
		// wait for the next connection
		conn, err := socket.Accept()
		if err != nil {
			panic(err)
		}

		if len(clients) >= MAX_CLIENTS {
			panic("Too many clients!")
		}

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
	defer conn.Close()

	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			// cross me once, get the hell outta here forever
			logger.With("error", err).Error("error reading from client - disconnecting")
			unsubscribe(mutex, clients, conn)
			return
		}

		logger.With("message", msg).Debug("received message from client")
		parts := strings.Split(msg, ":")
		verb := parts[0]
		switch verb {
		case "subscribe":
			subscribe(mutex, clients, conn)
		case "unsubscribe":
			unsubscribe(mutex, clients, conn)
		case "propose":
			proposedTheme := parts[1]
			s.currentTheme.Store(&proposedTheme)
			broadcast(mutex, clients, protocol.Set(proposedTheme))
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

			whisper(mutex, conn, protocol.Set(*theme))
		}
	}
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

// broadcast sends a message to all the clients
func broadcast(mutex *sync.Mutex, clients *[]net.Conn, msg []byte) {
	logger := zap.S()

	mutex.Lock()
	defer mutex.Unlock()
	logger.With("nClients", len(*clients), "message", string(msg)).Debug("broadcasting message to clients")

	for _, c := range *clients {
		_, err := c.Write(msg)
		if err != nil {
			logger.With("error", err, "client", c.RemoteAddr()).Warn("error writing to client during broadcast")
		}
	}
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
