package client

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type GhosttyClient struct{}

func NewGhosttyClient() Client {
	return GhosttyClient{}
}

func (a GhosttyClient) Start(socket string) error {
	return SubscribeToSocket(SetterWithContext(a.set, "ghostty"))(socket)
}

func (a GhosttyClient) set(theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	path := ExpandTilde(config.Client["ghostty"]["path"])
	zap.S().Debugw("applying theme", "client", "ghostty", "theme", theme.ThemeName, "variant", theme.VariantName, "path", path)

	// clear out the file and replace it with `theme = <theme>`
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open ghostty config file: %w", err)
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("theme = \"%s-%s\"\n", theme.ThemeName, theme.VariantName)); err != nil {
		return fmt.Errorf("failed to write to ghostty config file: %w", err)
	}

	// Run ghostty-shader-manager integration if configured.
	// shader-manager-light-off is a comma-separated list of shaders to disable
	// when a light theme is active (and re-enable when a dark theme is active).
	if shaderList := config.Client["ghostty"]["shader-manager-light-off"]; shaderList != "" {
		action := "on"
		if theme.Light {
			action = "off"
		}
		for _, shader := range strings.Split(shaderList, ",") {
			shader = strings.TrimSpace(shader)
			if shader == "" {
				continue
			}
			zap.S().Debugw("ghostty-shader-manager", "action", action, "shader", shader)
			if _, runErr := Run("ghostty-shader-manager", action, shader); runErr != nil {
				zap.S().Warnw("ghostty-shader-manager failed", "action", action, "shader", shader, "err", runErr)
			}
		}
	}

	blurKey := "shader-manager-blur-dark"
	if theme.Light {
		blurKey = "shader-manager-blur-light"
	}

	if shouldBlurStr := config.Client["ghostty"][blurKey]; shouldBlurStr != "" {
		shouldBlur, err := strconv.ParseBool(shouldBlurStr)
		if err != nil {
			zap.S().Warnw("invalid value for "+blurKey, "value", shouldBlurStr, "err", err)
		}
		action := "on"
		if !shouldBlur {
			action = "off"
		}

		zap.S().Debugw("ghostty-shader-manager", "shouldBlur", shouldBlur, "action", action)
		if _, runErr := Run("ghostty-shader-manager", "blur", action); runErr != nil {
			zap.S().Warnw("ghostty-shader-manager failed", "action", "blur", "shouldBlur", shouldBlur, "err", runErr)
		}
	}

	// send USR2 signal to ghostty process to reload config
	_, err = Run("pkill", "-USR2", "ghostty")
	return err
}
