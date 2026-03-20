package client

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
)

type FirefoxClient struct{}

func NewFirefoxClient() Client {
	return FirefoxClient{}
}

func (f FirefoxClient) Start(socket string) error {
	return SubscribeToSocket(f.set)(socket)
}

func (f FirefoxClient) set(theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	logger := zap.S().With("theme", theme.ThemeName, "variant", theme.VariantName)

	themePath := ExpandTilde(config.Client["firefox"]["theme-path"])
	logger = logger.With("path", themePath)

	var sb strings.Builder
	sb.WriteString(":root {\n")
	for _, color := range AllColors {
		sb.WriteString(fmt.Sprintf("  --%s: %s;\n", string(color), theme.Colors[color]))
	}
	sb.WriteString("}\n")

	if err := os.WriteFile(themePath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write firefox theme: %w", err)
	}

	logger.Info("updated firefox theme")
	return nil
}
