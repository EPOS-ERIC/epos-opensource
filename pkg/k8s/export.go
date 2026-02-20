package k8s

import (
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
)

type ExportOpts struct {
	// Required. Path to export the embedded manifests and .env file. If the path does not exist it will be created
	Path string
}

func Export(opts ExportOpts) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid export parameters: %w", err)
	}

	cfg := config.GetDefaultConfigBytes()

	path, err := common.Export(opts.Path, "k8s-config.yaml", cfg)
	if err != nil {
		return fmt.Errorf("failed to export config: %w", err)
	}

	display.Done("Exported config: %s", path)

	display.Info("You can now use the config to deploy the environment with 'epos-opensource k8s deploy --config %s'", path)

	return nil
}

func (d *ExportOpts) Validate() error {
	if d.Path == "" {
		return fmt.Errorf("path is required")
	}

	return nil
}
