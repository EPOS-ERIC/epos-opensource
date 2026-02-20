package docker

import (
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
)

type ExportOpts struct {
	// Required. Path to export the default environment config
	Path string
}

func Export(opts ExportOpts) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid export parameters: %w", err)
	}

	cfg := config.GetDefaultConfigBytes()

	display.Debug("loaded default docker config bytes")

	path, err := common.Export(opts.Path, "docker-config.yaml", cfg)
	if err != nil {
		return fmt.Errorf("failed to export config: %w", err)
	}

	display.Debug("exported docker config path: %s", path)
	display.Done("Exported config: %s", path)
	display.Info("You can now use the config to deploy the environment with 'epos-opensource docker deploy --config %s'", path)

	return nil
}

func (d *ExportOpts) Validate() error {
	display.Debug("path: %s", d.Path)

	if d.Path == "" {
		return fmt.Errorf("path is required")
	}

	return nil
}
