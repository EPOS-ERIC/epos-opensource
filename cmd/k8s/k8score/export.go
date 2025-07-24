package k8score

import (
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/display"
)

func Export(path string) error {
	err := common.Export(path, ".env", []byte(EnvFile))
	if err != nil {
		return err
	}

	for name, content := range EmbeddedManifestContents {
		err = common.Export(path, name, []byte(content))
		if err != nil {
			return err
		}
		display.Done("Exported %s", name)
	}

	display.Done("Successfully exported default environment files in %s", path)
	return nil
}
