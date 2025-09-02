package dockercore

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/epos-eu/epos-opensource/validate"
)

type ExportOpts struct {
	// Required. Path to export the embedded docker-compose.yaml and .env files
	Path string
}

func Export(opts ExportOpts) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid export parameters: %w", err)
	}
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
func (d *ExportOpts) Validate() error {
	if err := validate.PathExists(d.Path); err != nil {
		return fmt.Errorf("the path '%s' is not a valid path: %w", d.Path, err)
	}
	return nil
}
