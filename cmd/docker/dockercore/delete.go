package dockercore

import (
	_ "embed"
	"fmt"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
)

func Delete(name string) error {
	common.PrintStep("Deleting environment: %s", name)

	env, err := db.GetDockerByName(name)
	if err != nil {
		return fmt.Errorf("error getting docker environment from db called '%s': %w", name, err)
	}

	common.PrintStep("Stopping stack")

	if err := downStack(env.Directory, true); err != nil {
		return fmt.Errorf("docker compose down failed: %w", err)
	}

	common.PrintDone("Stopped environment: %s", name)

	if err := common.RemoveEnvDir(env.Directory); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", env.Directory, err)
	}

	err = db.DeleteDocker(name)
	if err != nil {
		return fmt.Errorf("failed to delete docker %s (dir: %s) in db: %w", name, env.Directory, err)
	}

	common.PrintDone("Deleted environment: %s", name)

	return nil
}
