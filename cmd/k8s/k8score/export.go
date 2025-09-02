package k8score

import (
	"fmt"
	"os"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/display"
)

type ExportOpts struct {
	// Required. Path to export the embedded manifests and .env file
	Path string
}

func Export(opts ExportOpts) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid export parameters: %w", err)
	}
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid export parameters: %w", err)
	}
	err := common.Export(opts.Path, ".env", []byte(EnvFile))
	if err != nil {
		return err
	}

	for name, content := range EmbeddedManifestContents {
		err = common.Export(opts.Path, name, []byte(content))
		if err != nil {
			return err
		}
		display.Done("Exported %s", name)
	}

	display.Done("Successfully exported default environment files in %s", opts.Path)
	return nil
}
func (e *ExportOpts) Validate() error {
	// TODO: check if its a path
	info, err := os.Stat(e.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory %s does not exist", e.Path)
		}
		return fmt.Errorf("cannot stat directory %s: %w", e.Path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", e.Path)
	}

	return nil
}
