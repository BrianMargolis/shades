package client

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// RunApplescript runs an AppleScript command and returns the output.
func RunApplescript(script string) ([]byte, error) {
	return Run("osascript", "-e", script)
}

// Run returns a command's output and error (if any).
func Run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

func ReplaceAtTag(
	filePath string,
	replacement string,
	tag string,
) (int, error) {
	numReplacements := 0

	f, err := os.Open(filePath)
	if err != nil {
		return numReplacements, errors.Wrap(err, "opening file")
	}
	defer f.Close()

	// write the modified version to a temp file
	tmp, err := os.CreateTemp("", "tempfile-")
	if err != nil {
		return numReplacements, errors.Wrap(err, "creating temp file")
	}
	defer tmp.Close()

	reader := bufio.NewReader(f)
	writer := bufio.NewWriter(tmp)
	for {
		line, err := reader.ReadString('\n')

		if strings.Contains(line, tag) {
			numReplacements++
			if _, err := writer.WriteString(replacement + "\n"); err != nil {
				return numReplacements, errors.Wrap(err, "writing replacement line")
			}
		} else {
			if _, err := writer.WriteString(line); err != nil {
				return numReplacements, errors.Wrap(err, "writing line")
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return numReplacements, errors.Wrap(err, "reading line")
		}
	}

	if err := writer.Flush(); err != nil {
		return numReplacements, errors.Wrap(err, "flushing temp file")
	}

	// already deferred, but we want to close the file before renaming it
	if err := tmp.Close(); err != nil {
		return numReplacements, errors.Wrap(err, "closing temp file")
	}

	if err := os.Rename(tmp.Name(), filePath); err != nil {
		return numReplacements, errors.Wrap(err, "replacing original file")
	}

	return numReplacements, nil
}

func ExpandTilde(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home + path[1:]
	}
	return path
}

func DoTemplate(
	template string,
	variant ThemeVariant,
) string {
	for color, value := range variant.Colors {
		template = strings.ReplaceAll(template, fmt.Sprintf("{%s}", color), value)
	}
	return template
}

func LookPath(executableName string) (string, error) {
	path, err := exec.LookPath(executableName)
	if err != nil {
		return "", errors.Wrapf(err, "executable not found: %s", executableName)
	}
	return path, nil
}
