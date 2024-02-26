package client

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type AlacrittyClient struct{}

func (a AlacrittyClient) Start(socket string, config map[string]string) error {
	return SubscribeToSocket(a.set(config))(socket)
}

func (a AlacrittyClient) set(config map[string]string) func(string) error {
	return func(theme string) error {
		alacrittyConfigPath := config["alacritty-config-path"]
		if alacrittyConfigPath == "" {
			alacrittyConfigPath = os.Getenv("HOME") + "/.config/alacritty/alacritty.toml"
		}

		themePath := config["light-theme-path"]
		if theme == "dark" {
			themePath = config["dark-theme-path"]
		}

		n, err := ReplaceAtTag(
			alacrittyConfigPath,
			fmt.Sprintf("\"%s\", # shades-replace", themePath),
			"# shades-replace",
		)
		if err != nil {
			return errors.Wrap(err, "replacing alacritty theme path")
		}

		if n == 0 {
			fmt.Println("WARNING: no '# shades-replace' tag found in alacritty config: " + alacrittyConfigPath)
		}

		return nil
	}
}

func (a AlacrittyClient) replaceAtShadesTag(filePath string, replacement string, tag string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return errors.Wrap(err, "opening file")
	}
	defer f.Close()

	// write the modified version to a temp file
	tmp, err := os.CreateTemp("", "tempfile-")
	if err != nil {
		return errors.Wrap(err, "creating temp file")
	}
	defer tmp.Close()

	didReplacement := false
	reader := bufio.NewReader(f)
	writer := bufio.NewWriter(tmp)
	for {
		line, err := reader.ReadString('\n')

		if strings.Contains(line, tag) {
			didReplacement = true
			if _, err := writer.WriteString("\"" + replacement + "\"," + " # shades-replace" + "\n"); err != nil {
				return errors.Wrap(err, "writing replacement line")
			}
		} else {
			if _, err := writer.WriteString(line); err != nil {
				return errors.Wrap(err, "writing line")
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.Wrap(err, "reading line")
		}
	}

	if !didReplacement {
		fmt.Printf("WARNING: no '# shades-replace' tag found in %s", filePath)
	}

	if err := writer.Flush(); err != nil {
		return errors.Wrap(err, "flushing temp file")
	}

	// already deferred, but we want to close the file before renaming it
	if err := tmp.Close(); err != nil {
		return errors.Wrap(err, "closing temp file")
	}

	if err := os.Rename(tmp.Name(), filePath); err != nil {
		return errors.Wrap(err, "replacing original file")
	}

	return nil
}
