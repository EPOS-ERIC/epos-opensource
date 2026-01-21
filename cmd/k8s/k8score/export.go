package k8score

import (
	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
)

type ExportOpts struct {
	// Required. Path to export the embedded manifests and .env file
	Path string
}

func Export(opts ExportOpts) error {
	header := common.GenerateExportHeader()

	envContent := header + EnvFile

	err := common.Export(opts.Path, ".env", []byte(envContent))
	if err != nil {
		return err
	}
	display.Done("Exported file: %s", ".env")

	for name, content := range EmbeddedManifestContents {
		fileContent := header + content
		err = common.Export(opts.Path, name, []byte(fileContent))
		if err != nil {
			return err
		}
		display.Done("Exported file: %s", name)
	}

	display.Done("All files exported to %s", opts.Path)
	return nil
}
