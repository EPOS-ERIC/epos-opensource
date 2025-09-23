package k8score

import (
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/display"
)

type ExportOpts struct {
	// Required. Path to export the embedded manifests and .env file
	Path string
}

func Export(opts ExportOpts) error {
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
