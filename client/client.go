package client

import (
	"brianmargolis/shades/protocol"
	"bufio"
	"context"
	"net"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Client interface {
	Start(csocket string) error
}

type ClientConstructor func() Client

// SocketAsChannel takes a socket path and opens up a channel on top of it.
func SocketAsChannel(socket string) (chan string, chan string, error) {
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, nil, err
	}

	// set up a channel that pulls out new line-delimited messages from the
	// socket
	readChan := make(chan string)
	go func() {
		defer func() {
			conn.Close()
			close(readChan)
		}()

		r := bufio.NewReader(conn)
		for {
			result, err := r.ReadString('\n')
			if err != nil {
				return
			}
			readChan <- result
		}
	}()

	// set up a channel that dumps messages back into the socket (callers need to
	// handle delimitation)
	writeChan := make(chan string)
	go func() {
		for message := range writeChan {
			conn.Write([]byte(message))
		}
	}()

	return readChan, writeChan, nil
}

type loggerKey struct{}

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// LoggerFromContext extracts the logger from context, and if it doesn't exist, returns a noop logger. Will never return nil.
func LoggerFromContext(ctx context.Context) *zap.SugaredLogger {
	if logger, ok := ctx.Value(loggerKey{}).(*zap.SugaredLogger); ok {
		return logger
	}
	return zap.NewNop().Sugar()
}

// SetterWithContext wraps a theme setter function with context-based logging
func SetterWithContext(
	setter func(ThemeVariant) error,
	clientName string,
) func(ThemeVariant) error {
	return func(theme ThemeVariant) error {
		logger := zap.S()

		logger = logger.With("client", clientName, "theme", theme.ThemeName, "variant", theme.VariantName)

		err := setter(theme)
		if err != nil {
			err = errors.Wrapf(err, "failed to set %s theme", clientName)
			logger.Errorw("Error setting theme", "err", err)
		}
		return err
	}
}

// SubscribeToSocket is a simple way to build a client if you don't need the
// propose or get functionalities.
func SubscribeToSocket(
	setter func(theme ThemeVariant) error,
) func(string) error {
	return func(socketName string) error {
		read, write, err := SocketAsChannel(socketName)
		if err != nil {
			return err
		}

		write <- string(protocol.Subscribe("client"))

		for message := range read {
			verb, noun, err := protocol.Parse(message)
			if err != nil {
				return err
			}

			if verb == "set" {
				config, err := GetConfig()
				if err != nil {
					return err
				}

				themeAndVariant := strings.TrimSpace(noun)
				themeVariant, err := config.Themes.GetVariant(themeAndVariant)
				if err != nil {
					return err
				}

				if err = setter(themeVariant); err != nil {
					return err
				}
			}
		}

		return nil
	}
}
