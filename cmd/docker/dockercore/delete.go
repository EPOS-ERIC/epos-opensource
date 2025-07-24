package dockercore

import (
	_ "embed"
	"fmt"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/display"
)

func Delete(name string) error {
	display.Step("Deleting environment: %s", name)

	env, err := db.GetDockerByName(name)
	if err != nil {
		return fmt.Errorf("error getting docker environment from db called '%s': %w", name, err)
	}

	display.Step("Stopping stack")

	if err := downStack(env.Directory, true); err != nil {
		return fmt.Errorf("docker compose down failed: %w", err)
	}

	display.Done("Stopped environment: %s", name)

	if err := common.RemoveEnvDir(env.Directory); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", env.Directory, err)
	}

	err = db.DeleteDocker(name)
	if err != nil {
		return fmt.Errorf("failed to delete docker %s (dir: %s) in db: %w", name, env.Directory, err)
	}

	display.Done("Deleted environment: %s", name)

	return nil
}
