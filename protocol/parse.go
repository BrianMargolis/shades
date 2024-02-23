package protocol

import (
	"errors"
	"strings"
)

func Parse(message string) (string, string, error) {
	parts := strings.Split(message, ":")
	if len(parts) != 2 {
		return "", "", errors.New("Invalid message format")
	}

	return parts[0], parts[1], nil
}
