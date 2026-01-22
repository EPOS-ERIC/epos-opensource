package dockercore

import (
	"fmt"
	"path/filepath"

	"github.com/EPOS-ERIC/epos-opensource/cmd/docker/dockercore/config"
	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
)

type ExportOpts struct {
	// Required. Path to export the default environment config
	Path string
}

func Export(opts ExportOpts) error {
	path, err := filepath.Abs(opts.Path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for export path %s: %w", opts.Path, err)
	}

	cfg := config.GetDefaultConfigBytes()

	err = common.Export(opts.Path, "config.yaml", cfg)
	if err != nil {
		return err
	}

	exportedPath := filepath.Join(path, "config.yaml")

	display.Done("Exported config: %s", exportedPath)

	display.Info("You can now use the config to deploy the environment with 'epos-opensource docker deploy --config %s'", exportedPath)

	return nil
}
