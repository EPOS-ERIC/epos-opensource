package docker

import (
	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
)

type ExportOpts struct {
	// Required. Path to export the default environment config
	Path string
}

func Export(opts ExportOpts) error {
	cfg := config.GetDefaultConfigBytes()

	path, err := common.Export(opts.Path, "docker-config.yaml", cfg)
	if err != nil {
		return err
	}

	display.Done("Exported config: %s", path)

	display.Info("You can now use the config to deploy the environment with 'epos-opensource docker deploy --config %s'", path)

	return nil
}
