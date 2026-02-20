package k8s

import (
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
)

// ExportOpts defines inputs for Export.
type ExportOpts struct {
	// Required. Path to export the default K8s config file. If the path does not exist it will be created
	Path string
}

// Export writes the default K8s configuration file to the requested path.
func Export(opts ExportOpts) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid export parameters: %w", err)
	}

	cfg := config.GetDefaultConfigBytes()

	display.Debug("loaded default k8s config bytes")

	path, err := common.Export(opts.Path, "k8s-config.yaml", cfg)
	if err != nil {
		return fmt.Errorf("failed to export config: %w", err)
	}

	display.Debug("exported k8s config path: %s", path)
	display.Done("Exported config: %s", path)
	display.Info("You can now use the config to deploy the environment with 'epos-opensource k8s deploy --config %s'", path)

	return nil
}

// Validate checks ExportOpts for required values.
func (d *ExportOpts) Validate() error {
	display.Debug("path: %s", d.Path)

	if d.Path == "" {
		return fmt.Errorf("path is required")
	}

	return nil
}
