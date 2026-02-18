package docker

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
)

type RenderOpts struct {
	// Optional. name to give to the environment. If not set it will use the name set in the passed config. If set and also set in the config, this has precedence
	Name string
	// Optional. custom config for the environment.
	Config *config.EnvConfig
	// Optional. path to export the embedded docker-compose.yaml and .env files
	OutputPath string
}

func Render(opts RenderOpts) ([]string, error) {
	cfg := opts.Config
	if opts.Config == nil {
		cfg = config.GetDefaultConfig()
	}

	if opts.Name != "" {
		cfg.Name = opts.Name
	} else if cfg.Name == "" {
		return nil, fmt.Errorf("environment name is required. Provide it as an argument or in the config file")
	}

	files, err := cfg.Render()
	if err != nil {
		return nil, fmt.Errorf("failed to render templates: %w", err)
	}

	if opts.OutputPath == "" {
		opts.OutputPath, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
	} else {
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
	}

	envFilePath := filepath.Join(opts.OutputPath, ".env")
	if err := common.CreateFileWithContent(envFilePath, files[".env"], true); err != nil {
		return nil, fmt.Errorf("failed to create .env file: %w", err)
	}

	composeFilePath := filepath.Join(opts.OutputPath, "docker-compose.yaml")
	if err := common.CreateFileWithContent(composeFilePath, files["docker-compose.yaml"], true); err != nil {
		return nil, fmt.Errorf("failed to create docker-compose.yaml file: %w", err)
	}

	return []string{
		envFilePath,
		composeFilePath,
	}, nil
}
