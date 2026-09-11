package protocol

import (
	"errors"
	"strings"
)

func Parse(message string) (string, string, error) {
	// SplitN, not Split: a noun is free to contain colons of its own (the
	// palette payload is JSON), and only the first one delimits the verb.
	parts := strings.SplitN(message, ":", 2)
	if len(parts) != 2 {
		return "", "", errors.New("Invalid message format")
	}

	return parts[0], parts[1], nil
}
