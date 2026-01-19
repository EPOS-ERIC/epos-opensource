package config

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/EPOS-ERIC/epos-opensource/common"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

func (e *EnvConfig) Render() (map[string]string, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}

	results := map[string]string{}
	var compose bytes.Buffer
	var env bytes.Buffer

	err = tmpl.ExecuteTemplate(&compose, "docker-compose.yaml.tmpl", e)
	if err != nil {
		return nil, fmt.Errorf("error executing template: %w", err)
	}

	err = tmpl.ExecuteTemplate(&env, ".env.tmpl", e)
	if err != nil {
		return nil, fmt.Errorf("error executing template: %w", err)
	}

	results["docker-compose.yaml"] = common.GenerateExportHeader() + compose.String()
	results[".env"] = common.GenerateExportHeader() + env.String()

	return results, nil
}
