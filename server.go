package main

import (
	"brianmargolis/shades/protocol"
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

const MAX_CLIENTS = 100

type Server interface {
	Start(socketPath string) error
}

type server struct {
	defaultDarkTheme  string
	defaultLightTheme string
}

func NewServer(defaultDarkTheme, defaultLightTheme string) Server {
	return &server{
		defaultDarkTheme:  defaultDarkTheme,
		defaultLightTheme: defaultLightTheme,
	}
}

func (s *server) Start(socketPath string) error {
	socket, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Println("Error listening: ", err)
		panic(err)
	}
	planForDeath(socket)

	currentlyDark, err := isCurrentlyDark()
	if err != nil {
		fmt.Println("Error checking current state: ", err)
		panic(err)
	}
	_ = currentlyDark

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

func (s *server) talkToClient(conn net.Conn, clients *[]net.Conn, mutex *sync.Mutex) {
	fmt.Println("new client connected: ", conn.RemoteAddr())
	defer conn.Close()

	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			log.Println("ERROR: client disconnect: ", err)
			unsubscribe(mutex, clients, conn)
			return
		}

		parts := strings.Split(msg, ":")
		fmt.Printf("Received: %s\n", msg)
		verb := parts[0]
		switch verb {
		case "subscribe":
			subscribe(mutex, clients, conn)
		case "unsubscribe":
			unsubscribe(mutex, clients, conn)
		case "propose":
			proposedTheme := parts[1]
			// darkProposed := proposedTheme == "dark"
			// currentlyDark, err := isCurrentlyDark()
			// if err != nil {
			// 	continue
			// }

			broadcast(mutex, clients, protocol.Set(proposedTheme))
		case "get":
			currentlyDark, err := isCurrentlyDark()
			if err != nil {
				continue
			}

			theme := s.defaultLightTheme
			if currentlyDark {
				theme = s.defaultDarkTheme
			}

			whisper(mutex, conn, protocol.Set(theme))
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
	mutex.Lock()
	defer mutex.Unlock()
	fmt.Printf("broadcasting to %d clients: %s\n", len(*clients), string(msg))
	for _, c := range *clients {
		_, err := c.Write(msg)
		if err != nil {
			log.Println("Write error: ", err)
		}
	}
}

// whisper sends a message to just one client
func whisper(mutex *sync.Mutex, conn net.Conn, msg []byte) {
	mutex.Lock()
	defer mutex.Unlock()
	_, err := conn.Write(msg)
	if err != nil {
		log.Println("Write error: ", err)
	}
}

func isCurrentlyDark() (bool, error) {
	script := `tell application "System Events" to tell appearance preferences to get dark mode`

	output, err := exec.Command("osascript", "-e", script).Output()
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
		fmt.Println("Received an interrupt, stopping services...")
		socket.Close()
		os.Exit(1)
	}()
}
