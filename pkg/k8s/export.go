package k8s

import (
	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
)

type ExportOpts struct {
	// Required. Path to export the embedded manifests and .env file
	Path string
}

func Export(opts ExportOpts) error {
	cfg := config.GetDefaultConfigBytes()

	path, err := common.Export(opts.Path, "k8s-config.yaml", cfg)
	if err != nil {
		return err
	}

	display.Done("Exported config: %s", path)

	display.Info("You can now use the config to deploy the environment with 'epos-opensource k8s deploy --config %s'", path)

	return nil
}
