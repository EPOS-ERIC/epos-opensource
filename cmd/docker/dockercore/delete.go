package dockercore

import (
	_ "embed"
	"fmt"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
)

func Delete(name string) error {
	common.PrintStep("Deleting environment: %s", name)

	dir, err := common.GetEnvDir(name, platform)
	if err != nil {
		return fmt.Errorf("failed to resolve environment directory: %w", err)
	}

	common.PrintInfo("Environment directory: %s", dir)
	common.PrintStep("Stopping stack")

	if err := downStack(dir, true); err != nil {
		return fmt.Errorf("docker compose down failed: %w", err)
	}

	common.PrintDone("Stopped environment: %s", name)

	if err := common.RemoveEnvDir(dir); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", dir, err)
	}

	err = db.DeleteEnv(name, "docker")
	if err != nil {
		return fmt.Errorf("failed to delete env %s (dir: %s, platform: %s) in db: %w", name, dir, "docker", err)
	}

	common.PrintDone("Deleted environment: %s", name)

	return nil
}
