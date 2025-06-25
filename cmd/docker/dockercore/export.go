package dockercore

import (
	"github.com/epos-eu/epos-opensource/common"
)

func Export(path string) error {
	err := common.Export(path, ".env", []byte(EnvFile))
	if err != nil {
		return err
	}
	common.PrintDone("Exported %s", ".env")

	err = common.Export(path, "docker-compose.yaml", []byte(ComposeFile))
	if err != nil {
		return err
	}
	common.PrintDone("Exported %s", "docker-compose.yaml")

	common.PrintDone("Successfully exported default environment files in %s", path)
	return nil
}
