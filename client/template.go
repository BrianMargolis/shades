package client

import (
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// TemplateClient renders arbitrary template files for apps that don't get
// their own dedicated client. Each entry in config.Templates names a template
// file, an output path, and the character that wraps color placeholders in
// the template (see RenderTemplate).
type TemplateClient struct{}

func NewTemplateClient() Client {
	return TemplateClient{}
}

func (t TemplateClient) Start(socket string) error {
	return SubscribeToSocket(SetterWithContext(t.set, "template"))(socket)
}

func (t TemplateClient) set(theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	for _, tmpl := range config.Templates {
		templatePath := ExpandTilde(tmpl.TemplatePath)
		outputPath := ExpandTilde(tmpl.OutputPath)

		zap.S().Debugw("applying theme", "client", "template", "theme", theme.ThemeName, "variant", theme.VariantName, "templatePath", templatePath, "outputPath", outputPath)

		templateContent, err := os.ReadFile(templatePath)
		if err != nil {
			return errors.Wrapf(err, "reading template file %s", templatePath)
		}

		rendered := RenderTemplate(string(templateContent), tmpl.EscapeChar, theme)

		if err := os.WriteFile(outputPath, []byte(rendered), 0644); err != nil {
			return errors.Wrapf(err, "writing rendered template to %s", outputPath)
		}
	}

	return nil
}
