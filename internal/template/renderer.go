package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	gotemplate "text/template"

	"github.com/yourorg/patchwork-deploy/internal/config"
)

// Renderer renders deployment config templates using app and environment variables.
type Renderer struct {
	cfg *config.Config
}

// NewRenderer creates a new Renderer for the given config.
func NewRenderer(cfg *config.Config) *Renderer {
	return &Renderer{cfg: cfg}
}

// RenderFile reads a template file and renders it with the given app's variables.
// The rendered output is returned as a string.
func (r *Renderer) RenderFile(appName, templatePath string) (string, error) {
	app, ok := r.cfg.Apps[appName]
	if !ok {
		return "", fmt.Errorf("template: unknown app %q", appName)
	}

	raw, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("template: read file %q: %w", templatePath, err)
	}

	return r.render(appName, app, string(raw))
}

// RenderString renders a template string with the given app's variables.
func (r *Renderer) RenderString(appName, tmplStr string) (string, error) {
	app, ok := r.cfg.Apps[appName]
	if !ok {
		return "", fmt.Errorf("template: unknown app %q", appName)
	}

	return r.render(appName, app, tmplStr)
}

func (r *Renderer) render(appName string, app config.App, tmplStr string) (string, error) {
	data := buildTemplateData(appName, app, r.cfg.Global)

	tmpl, err := gotemplate.New(appName).Option("missingkey=error").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("template: parse: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template: execute: %w", err)
	}

	return buf.String(), nil
}

// buildTemplateData merges global and app-level env vars into a flat map.
func buildTemplateData(appName string, app config.App, global config.GlobalConfig) map[string]string {
	data := map[string]string{
		"AppName": appName,
		"Dir":     app.Dir,
		"Host":    app.Host,
	}

	for k, v := range global.Env {
		data[sanitizeKey(k)] = v
	}
	for k, v := range app.Env {
		data[sanitizeKey(k)] = v
	}

	return data
}

// sanitizeKey replaces characters invalid in Go template field names with underscores.
func sanitizeKey(k string) string {
	return strings.NewReplacer("-", "_", ".", "_", " ", "_").Replace(
		filepath.Base(k),
	)
}
