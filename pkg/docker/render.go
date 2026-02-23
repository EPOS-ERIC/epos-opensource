package docker

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
	"github.com/EPOS-ERIC/epos-opensource/validate"
)

// RenderOpts defines inputs for Render.
type RenderOpts struct {
	// Optional. name to give to the environment. If not set it will use the name set in the passed config. If set and also set in the config, this has precedence
	Name string
	// Optional. custom config for the environment.
	Config *config.EnvConfig
	// Optional. path to export the embedded docker-compose.yaml and .env files
	OutputPath string
}

// Render materializes Docker runtime files from configuration and returns the created file paths.
func Render(opts RenderOpts) ([]string, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid render parameters: %w", err)
	}

	display.Debug("rendering docker templates")

	files, err := opts.Config.Render()
	if err != nil {
		return nil, fmt.Errorf("failed to render templates: %w", err)
	}

	display.Debug("rendered docker templates: %d", len(files))

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

	envFilePath := filepath.Join(opts.OutputPath, ".env")
	display.Debug("writing file: %s", envFilePath)

	if err := common.CreateFileWithContent(envFilePath, files[".env"], false); err != nil {
		return nil, fmt.Errorf("failed to create .env file: %w", err)
	}

	composeFilePath := filepath.Join(opts.OutputPath, "docker-compose.yaml")

	display.Debug("writing file: %s", composeFilePath)

	if err := common.CreateFileWithContent(composeFilePath, files["docker-compose.yaml"], false); err != nil {
		return nil, fmt.Errorf("failed to create docker-compose.yaml file: %w", err)
	}

	display.Done("Rendered Docker config in: %s", opts.OutputPath)

	return []string{
		envFilePath,
		composeFilePath,
	}, nil
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
