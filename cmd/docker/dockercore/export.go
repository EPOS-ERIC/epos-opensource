package dockercore

import (
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/display"
)

func Export(path string) error {
	err := common.Export(path, ".env", []byte(EnvFile))
	if err != nil {
		return err
	}
	display.Done("Exported %s", ".env")

	err = common.Export(path, "docker-compose.yaml", []byte(ComposeFile))
	if err != nil {
		return err
	}
	display.Done("Exported %s", "docker-compose.yaml")

	display.Done("Successfully exported default environment files in %s", path)
	return nil
}
