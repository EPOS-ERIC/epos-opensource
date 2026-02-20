package config

import (
	"fmt"

	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

// Render renders Helm templates using the current configuration values.
func (e *Config) Render() (map[string]string, error) {
	chart, err := GetChart()
	if err != nil {
		return nil, fmt.Errorf("load chart from embedded filesystem: %w", err)
	}

	values, err := e.AsValues()
	if err != nil {
		return nil, fmt.Errorf("build Helm values: %w", err)
	}

	renderValues, err := chartutil.ToRenderValues(chart, values.AsMap(), chartutil.ReleaseOptions{
		Name:      e.Name,
		Namespace: e.Name,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("build Helm render values: %w", err)
	}

	files, err := engine.Render(chart, renderValues)
	if err != nil {
		return nil, fmt.Errorf("render Helm templates: %w", err)
	}

	return files, nil
}
