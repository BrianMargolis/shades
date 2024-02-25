package client

import (
	"brianmargolis/shades/protocol"
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type AlacrittyClient struct{}

func (a AlacrittyClient) Start(socket string, config map[string]string) error {
	read, write, err := SocketAsChannel(socket)
	if err != nil {
		return err
	}

	write <- string(protocol.Subscribe("alacritty"))
	for message := range read {
		verb, noun, err := protocol.Parse(message)
		if err != nil {
			return err
		}

		if verb == "set" {
			theme := strings.TrimSpace(noun)
			err = a.set(theme)
			if err != nil {
				panic(err)
			}
		}
	}

	return nil
}

func (a AlacrittyClient) set(theme string) error {
	alacrittyConfigPath := os.Getenv("HOME") + "/.config/alacritty/alacritty.toml"
	themePath := fmt.Sprintf("\"~/.config/alacritty/themes/themes/everforest_%s.toml\"", theme)
	return a.replaceLineInFile(alacrittyConfigPath, 2, themePath)
}

// ReplaceLineInFile replaces a specific line in a file with a given replacement string.
// `filePath` is the path to the file, `lineNumber` is the 1-based line number to replace,
// and `replacement` is the new content for the specified line.
func (a AlacrittyClient) replaceLineInFile(filePath string, lineNumber int, replacement string) error {
	// Open the original file for reading.
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	// Create a temporary file where the modified content will be written.
	tempFile, err := os.CreateTemp("", "tempfile-")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer tempFile.Close()

	// Use bufio for efficient reading and writing.
	reader := bufio.NewReader(file)
	writer := bufio.NewWriter(tempFile)

	// Iterate over each line in the file.
	currentLine := 1
	for {
		line, err := reader.ReadString('\n')

		// When the target line number is reached, write the replacement string instead.
		if currentLine == lineNumber {
			if _, err := writer.WriteString(replacement + "\n"); err != nil {
				return fmt.Errorf("writing replacement line: %w", err)
			}
		} else {
			if _, err := writer.WriteString(line); err != nil {
				return fmt.Errorf("writing line: %w", err)
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("reading file: %w", err)
		}
		currentLine++
	}

	// Ensure all writes are flushed to the temporary file.
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flushing writes to temp file: %w", err)
	}

	// Close the original file (already deferred) and the temporary file.
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	// Replace the original file with the modified temporary file.
	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		return fmt.Errorf("replacing original file with temp file: %w", err)
	}

	return nil
}
