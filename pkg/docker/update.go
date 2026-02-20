// Package docker contains the internal functions used by the docker cmd to manage the environments
package docker

import (
	_ "embed"
	"fmt"
	"log"
	"path/filepath"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/db"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/db/sqlc"
)

// UpdateOpts defines inputs for Update.
type UpdateOpts struct {
	// Pull images before deploying the updated environment
	PullImages bool
	// Stop and remove containers before updating. Useful to reset the database
	Force bool
	// Reset config to embedded defaults. Cannot be used with NewConfig
	Reset bool
	// Name of the environment to update (required)
	OldEnvName string
	// New configuration to apply. If nil, preserves existing config
	NewConfig *config.EnvConfig
}

// Update updates an existing Docker environment with new configuration.
// It performs a safe update with rollback capability:
//   - Creates a backup of the current environment in a temp directory
//   - Deploys the new configuration
//   - On failure, restores the backup automatically
//
// Options:
//   - PullImages: Pull images before deploying
//   - Force: Stop and reset the database before updating
//   - Reset: Reset config to embedded defaults (ignores NewConfig if set)
//   - OldEnvName: Name of the environment to update (required)
//   - NewConfig: New configuration to apply (optional - preserves existing if nil)
//
// Returns the updated Docker record on success.
// Returns an error if validation fails or update cannot complete.
func Update(opts UpdateOpts) (*sqlc.Docker, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters for update command: %w", err)
	}

	display.Debug("loading docker environment from db: %s", opts.OldEnvName)

	docker, err := db.GetDockerByName(opts.OldEnvName)
	if err != nil {
		return nil, fmt.Errorf("error getting docker environment from db called '%s': %w", opts.OldEnvName, err)
	}

	display.Debug("loaded docker environment directory: %s", docker.Directory)
	display.Debug("loading current config from: %s", filepath.Join(docker.Directory, "config.yaml"))

	oldConfig, err := config.LoadConfig(filepath.Join(docker.Directory, "config.yaml"))
	if err != nil {
		return nil, fmt.Errorf("error loading config from directory '%s': %w", docker.Directory, err)
	}

	if opts.Reset {
		display.Info("Reset flag enabled: using default config")

		opts.NewConfig = config.GetDefaultConfig()
	}

	if opts.NewConfig == nil {
		display.Debug("new config not provided, using existing environment config")

		opts.NewConfig = oldConfig
	}

	if opts.NewConfig.Name == "" {
		opts.NewConfig.Name = oldConfig.Name

		display.Debug("new config name not set, using old name: %s", opts.NewConfig.Name)
	}

	display.Step("Updating environment: %s", opts.OldEnvName)

	if !opts.PullImages {
		updates, err := opts.NewConfig.CheckForUpdates()
		if err != nil {
			log.Printf("error checking for updates: %v", err)
		}
		display.ImageUpdatesAvailable(updates, opts.NewConfig.Name)
	}

	bkpedir, err := common.CreateTmpCopy(docker.Directory)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup copy: %w", err)
	}

	display.Debug("created backup directory: %s", bkpedir)

	handleFailure := func(msg string, mainErr error) (*sqlc.Docker, error) {
		display.Error("Failed to update environment: %v", mainErr)
		display.Step("Restoring environment from backup")
		display.Warn("update failed, rolling back from backup")

		if err := common.RemoveEnvDir(docker.Directory); err != nil {
			display.Error("Failed to remove corrupted directory: %v", err)
		}

		if err := common.RestoreTmpDir(bkpedir, docker.Directory); err != nil {
			display.Error("Failed to restore from backup: %v", err)
		} else {
			if err := deployStack(true, docker.Directory); err != nil {
				display.Error("Failed to deploy restored environment: %v", err)
			}
		}

		if cleanupErr := common.RemoveTmpDir(bkpedir); cleanupErr != nil {
			display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
		}

		return nil, fmt.Errorf(msg, mainErr)
	}

	display.Step("Updating stack")

	// If force is set do a docker compose down on the original env
	if opts.Force {
		display.Info("Force flag enabled: stopping old environment with volumes")
		display.Step("Stopping old environment")

		if err := downStack(docker.Directory, true); err != nil {
			return handleFailure("docker compose down failed: %w", err)
		}

		display.Done("Stopped environment: %s", oldConfig.Name)
	}

	display.Step("Removing old environment directory")

	// Remove the contents of the env dir and create the updated env file and docker-compose
	if err := common.RemoveEnvDir(docker.Directory); err != nil {
		return handleFailure("failed to remove directory: %w", fmt.Errorf("%s: %w", docker.Directory, err))
	}

	display.Debug("removed old environment directory: %s", docker.Directory)
	display.Step("Creating new environment directory")

	// create the directory in the parent
	dir, err := NewEnvDir(filepath.Dir(docker.Directory), opts.NewConfig)
	if err != nil {
		return handleFailure("failed to prepare environment directory: %w", err)
	}

	display.Debug("created updated environment directory: %s", dir)
	display.Done("Updated environment created in directory: %s", dir)

	// If pullImages is set before deploying pull the images for the updated env
	if opts.PullImages {
		display.Debug("pulling images before stack deployment")

		if err := pullEnvImages(dir, opts.NewConfig.Name); err != nil {
			display.Error("Pulling images failed: %v", err)
			return handleFailure("pulling images failed: %w", err)
		}
	}

	// Deploy the updated compose
	display.Debug("deploying updated stack")

	if err = deployStack(true, dir); err != nil {
		display.Error("Deploy failed: %v", err)
		return handleFailure("deploy failed: %w", err)
	}

	display.Debug("building environment URLs")

	urls, err := opts.NewConfig.BuildEnvURLs()
	if err != nil {
		display.Error("error building urls: %v", err)
		return handleFailure("error building urls: %w", err)
	}

	display.Debug("built urls: %+v", urls)

	// only repopulate the ontologies if the database has been cleaned
	if opts.Force {
		display.Debug("force update: repopulating base ontologies and clearing ingested tracking")

		if err := common.PopulateOntologies(urls.APIURL); err != nil {
			display.Error("error initializing the ontologies in the environment: %v", err)
			return handleFailure("error initializing the ontologies in the environment: %w", err)
		}

		if err := db.DeleteIngestedFilesByEnvironment(opts.NewConfig.Name); err != nil {
			return handleFailure("failed to clear ingested files tracking: %w", err)
		}
	}

	// If everything goes right, delete the tmp dir and finish
	if cleanupErr := common.RemoveTmpDir(bkpedir); cleanupErr != nil {
		display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
	} else {
		display.Debug("cleaned backup directory: %s", bkpedir)
	}

	display.Debug("deleting old docker record: %s", opts.OldEnvName)

	err = db.DeleteDocker(opts.OldEnvName)
	if err != nil {
		return handleFailure("failed to delete old docker in db: %w", fmt.Errorf("%s (dir: %s): %w", opts.OldEnvName, docker.Directory, err))
	}

	var backofficePort int64
	if opts.NewConfig.Components.Backoffice.Enabled {
		backofficePort = int64(opts.NewConfig.Components.Backoffice.GUI.Port)
	}

	newDocker, err := db.UpsertDocker(sqlc.Docker{
		Name:           opts.NewConfig.Name,
		Directory:      dir,
		ApiUrl:         urls.APIURL,
		GuiUrl:         urls.GUIURL,
		BackofficeUrl:  urls.BackofficeURL,
		ApiPort:        int64(opts.NewConfig.Components.Gateway.Port),
		GuiPort:        int64(opts.NewConfig.Components.PlatformGUI.Port),
		BackofficePort: &backofficePort,
	})
	if err != nil {
		return handleFailure("failed to insert docker in db: %w", fmt.Errorf("%s (dir: %s): %w", opts.NewConfig.Name, dir, err))
	}

	display.Done("Updated environment: %s", opts.NewConfig.Name)

	return newDocker, nil
}

// Validate checks UpdateOpts for consistency and verifies that the target environment exists.
func (u *UpdateOpts) Validate() error {
	display.Debug("oldEnvName: %s", u.OldEnvName)
	display.Debug("pullImages: %v", u.PullImages)
	display.Debug("force: %v", u.Force)
	display.Debug("reset: %v", u.Reset)
	display.Debug("newConfig: %+v", u.NewConfig)

	if u.OldEnvName == "" {
		return fmt.Errorf("name is required")
	}

	if err := EnsureEnvironmentExists(u.OldEnvName); err != nil {
		return fmt.Errorf("no environment with name '%s' exists: %w", u.OldEnvName, err)
	}

	if u.Reset && u.NewConfig != nil {
		return fmt.Errorf("cannot specify custom config when Reset is true")
	}

	if u.NewConfig != nil {
		if err := u.NewConfig.Validate(); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}
	}

	return nil
}
