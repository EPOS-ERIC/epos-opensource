package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
	"github.com/EPOS-ERIC/epos-opensource/validate"
)

// RenderOpts defines inputs for Render.
type RenderOpts struct {
	// Optional. Name to give to the environment. If not set it will use the name set in the passed config. If set and also set in the config, this has precedence
	Name string
	// Optional. Custom config for the environment.
	Config *config.Config
	// Optional. path to export the embedded k8s-compose.yaml and .env files
	OutputPath string
}

// Render materializes rendered K8s manifests from configuration and returns created file paths.
func Render(opts RenderOpts) ([]string, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid render parameters: %w", err)
	}

	display.Debug("rendering k8s templates")
	files, err := opts.Config.Render()
	if err != nil {
		return nil, fmt.Errorf("failed to render templates: %w", err)
	}
	display.Debug("rendered k8s templates: %d", len(files))

	if opts.OutputPath == "" {
		display.Debug("output path not set, using current working directory")

		opts.OutputPath, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
	} else {
		display.Debug("resolving output path")

		opts.OutputPath, err = filepath.Abs(opts.OutputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for output path: %w", err)
		}

		if _, err := os.Stat(opts.OutputPath); err == nil {
			return nil, fmt.Errorf("environment directory %s already exists", opts.OutputPath)
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to check directory %s: %w", opts.OutputPath, err)
		}

		if err := os.MkdirAll(opts.OutputPath, 0o750); err != nil {
			return nil, fmt.Errorf("failed to create environment directory %s: %w", opts.OutputPath, err)
		}

		display.Debug("created output directory: %s", opts.OutputPath)
	}

	outputPaths := []string{}
	for name, out := range files {
		// don't save empty files
		out = strings.TrimSpace(out)
		if out == "" || out == "---" {
			continue
		}

		filePath := filepath.Join(opts.OutputPath, name)

		display.Debug("writing file: %s", filePath)

		if err := os.MkdirAll(filepath.Dir(filePath), 0o750); err != nil {
			return nil, fmt.Errorf("failed to create parent directory for %s file: %w", name, err)
		}

		err := common.CreateFileWithContent(filePath, out, true)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s file: %w", name, err)
		}

		outputPaths = append(outputPaths, filePath)
	}

	display.Done("Rendered K8s config in: %s", opts.OutputPath)

	return outputPaths, nil
}

// Validate checks RenderOpts and applies defaults needed before rendering.
func (r *RenderOpts) Validate() error {
	display.Debug("name: %s", r.Name)
	display.Debug("outputPath: %s", r.OutputPath)
	display.Debug("config: %+v", r.Config)

	if r.Config == nil {
		r.Config = config.GetDefaultConfig()

		display.Debug("config not provided, using default config")
	}

	if r.Name == "" {
		r.Name = r.Config.Name
	}

	if r.Name == "" {
		return fmt.Errorf("environment name is required")
	}

	if err := validate.Name(r.Name); err != nil {
		return fmt.Errorf("invalid name for environment: %w", err)
	}

	r.Config.Name = r.Name

	display.Debug("validated render config name: %s", r.Config.Name)

	return nil
}
