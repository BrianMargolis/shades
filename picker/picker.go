package picker

import (
	"brianmargolis/shades/client"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type PickerOpts struct {
	SocketPath string
	UseTmux    bool
}

type Picker interface {
	Start(PickerOpts) (result string, err error)
}

type picker struct {
	logger *zap.SugaredLogger
}

func NewPicker(
	logger *zap.Logger,
) Picker {
	zap.S().Info("NewPicker")
	return &picker{
		logger: logger.Sugar(),
	}
}

func (p *picker) Start(opts PickerOpts) (result string, err error) {
	p.logger.Debug("Start")
	// TODO:
	// first, get the current theme - if the user bails without picking a theme,
	// we want to restore that theme as the previewer will have changed it

	result, err = p.pick(opts)
	if err != nil {
		err = errors.Wrap(err, "failed to pick")
		p.logger.Error(err.Error())
		return
	}

	err = client.ChangerClient{Theme: result}.Start(opts.SocketPath)
	if err != nil {
		err = errors.Wrap(err, "failed to start ChangerClient")
		p.logger.Error(err.Error())
	}
	return
}

func (p *picker) pick(opts PickerOpts) (result string, err error) {
	config, err := client.GetConfig()
	if err != nil {
		err = errors.Wrap(err, "failed to get config")
		return
	}
	p.logger.Debugw("config", "config", config)

	pickerOptions := p.getOptions(config)
	p.logger.Debugw("options", "options", pickerOptions)

	fzfPath, err := client.LookPath(p.getCommand(opts))
	if err != nil {
		err = errors.Wrap(err, "failed to get fzf executable path")
		return
	}
	p.logger.Debugw("fzfPath", "fzfPath", fzfPath)

	fzfOptions := []string{
		"--height=44",
		// save an enter once we've narrowed it down to one
		"--bind=one:accept",
		// live preview
		"--bind=focus:execute(shades set {})",
		"--preview=shades preview {}",
		"--no-scrollbar",
		"--preview-window",
		"up,70%,border-none",
		"--cycle",
	}

	if opts.UseTmux {
		// floating window
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

func (*picker) getOptions(config client.ConfigModel) []string {
	pickerOptions := []string{}
	for themeName, theme := range config.Themes {
		for variantName, _ := range theme.Variants {
			pickerOptions = append(pickerOptions, fmt.Sprintf("%s;%s", themeName, variantName))
		}
	}
	return pickerOptions
}
