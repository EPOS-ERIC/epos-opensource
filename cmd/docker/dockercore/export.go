package dockercore

import (
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/display"
)

type ExportOpts struct {
	// Required. Path to export the embedded docker-compose.yaml and .env files
	Path string
}

func Export(opts ExportOpts) error {
	err := common.Export(opts.Path, ".env", []byte(EnvFile))
	if err != nil {
		return err
	}
	display.Done("Exported %s", ".env")

	err = common.Export(opts.Path, "docker-compose.yaml", []byte(ComposeFile))
	if err != nil {
		return err
	}
	display.Done("Exported %s", "docker-compose.yaml")

	display.Done("Successfully exported default environment files in %s", opts)
	return nil
}
