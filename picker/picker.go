package picker

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/brianmargolis/shades/client"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// The picker shells back out to shades for these, so they're commands, but
// they're internal to fzf's bindings and left out of the help.
const (
	ListCommand           = "_picker-list"
	ToggleFavoriteCommand = "_toggle-favorite"
)

const favoriteKey = "ctrl-f"

const favoriteMarker = "\x1b[33m★\x1b[0m"

type PickerOpts struct {
	client.Filter
	SocketPath string
	UseTmux    bool
}

type Picker interface {
	Start(PickerOpts) (result string, err error)
}

type picker struct{}

func NewPicker() Picker {
	zap.S().Info("NewPicker")
	return &picker{}
}

func (p *picker) Start(opts PickerOpts) (result string, err error) {
	logger := zap.S()
	logger.Debug("Start")
	// TODO:
	// first, get the current theme - if the user bails without picking a theme,
	// we want to restore that theme as the previewer will have changed it

	result, err = p.pick(logger, opts)
	if err != nil {
		err = errors.Wrap(err, "failed to pick")
		logger.Error(err.Error())
		return
	}

	err = client.ChangerClient{Theme: result}.Start(context.Background(), opts.SocketPath)
	if err != nil {
		err = errors.Wrap(err, "failed to start ChangerClient")
		logger.Error(err.Error())
	}
	return
}

func (p *picker) pick(
	logger *zap.SugaredLogger,
	opts PickerOpts,
) (result string, err error) {

	config, err := client.GetConfig()
	if err != nil {
		err = errors.Wrap(err, "failed to get config")
		return
	}
	logger.Debugw("config", "config", config)

	pickerOptions := Lines(config, opts.Filter)
	logger.Debugw("options", "options", pickerOptions)
	if len(pickerOptions) == 0 {
		if opts.OnlyFavorites {
			err = errors.New("no favorite theme variants match; mark one with " + favoriteKey + " in 'shades i', or 'favorite: true' in your config")
		} else {
			err = errors.New("no theme variants match")
		}
		return
	}

	fzfPath, err := client.LookPath(p.getCommand(opts))
	if err != nil {
		err = errors.Wrap(err, "failed to get fzf executable path")
		return
	}
	logger.Debugw("fzfPath", "fzfPath", fzfPath)

	fzfOptions := []string{
		// A bare number is a line count to fzf, not a percentage, so the
		// percent sign is what makes the picker fill the terminal.
		"--height=100%",
		// Each line is "theme;variant<tab>display". fzf shows and searches the
		// display, while the actions and the final output use the bare name.
		"--ansi",
		"--delimiter=\t",
		"--with-nth=2",
		"--accept-nth=1",
		// keep the cursor on the same theme when a favorite toggle reloads the list
		"--track",
		"--id-nth=1",
		"--header=" + favoriteKey + ": toggle favorite",
		// save an enter once we've narrowed it down to one
		"--bind=one:accept",
		// live preview. execute would switch to the alternate screen on every
		// focus change, flashing the whole picker once per keypress.
		"--bind=focus:execute-silent(shades set {1})",
		fmt.Sprintf(
			"--bind=%s:execute-silent(shades %s {1})+reload(shades %s)",
			favoriteKey,
			ToggleFavoriteCommand,
			strings.Join(append([]string{ListCommand}, opts.Flags()...), " "),
		),
		"--preview=shades preview {1}",
		"--no-scrollbar",
		"--preview-window",
		// The preview is one swatch line per distinct palette color, so sizing
		// the window to that count shows the whole palette with no dead space
		// and leaves every other line to the theme list. fzf clamps this if the
		// terminal is too short to honor it.
		fmt.Sprintf("up,%d,border-none", len(client.AllColors)),
		"--cycle",
	}

	if opts.UseTmux {
		// floating window. fzf-tmux appends --no-height after our arguments, so
		// --height above does nothing here and the popup geometry is the only
		// height control on this path.
		fzfOptions = append([]string{
			"-w 50%",
		}, fzfOptions...)
	}
	cmd := exec.Command(fzfPath, fzfOptions...)
	pipeIn, err := cmd.StdinPipe()
	if err != nil {
		err = errors.Wrap(err, "failed to create pipe into fzf")
		return
	}
	pipeIn.Write([]byte(strings.Join(pickerOptions, "\n")))
	pipeIn.Close()

	pipeOut, err := cmd.StdoutPipe()
	if err != nil {
		err = errors.Wrap(err, "failed to get stdout pipe")
		return
	}
	cmd.Stderr = os.Stderr

	resultBytes := []byte{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		resultBytes, err = io.ReadAll(pipeOut)
		if err != nil {
			err = errors.Wrap(err, "failed to read pipe output")
		}
	}()

	err = cmd.Run()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			err = errors.Wrapf(err, "fzf exited with status %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return
	}

	wg.Wait()
	result = string(resultBytes)

	return
}

func (p *picker) getCommand(opts PickerOpts) string {
	if opts.UseTmux {
		return "fzf-tmux"
	}
	return "fzf"
}

// Lines builds the fzf input, one "theme;variant<tab>display" line per variant,
// with favorites starred in the terminal's yellow. The order has to be stable
// because fzf reloads the list after every favorite toggle.
func Lines(config client.ConfigModel, filter client.Filter) []string {
	lines := []string{}
	for _, name := range config.Themes.Names(filter) {
		variant, err := config.Themes.GetVariant(name)
		if err != nil {
			continue
		}

		marker := " "
		if variant.Favorite {
			marker = favoriteMarker
		}
		lines = append(lines, fmt.Sprintf("%s\t%s %s", name, marker, name))
	}
	return lines
}
