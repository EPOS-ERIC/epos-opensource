// Package docker contains the internal functions used by the docker cmd to manage the environments
package docker

import (
	"fmt"
	"log"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/db"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
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
func Update(opts UpdateOpts) (*Env, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters for update command: %w", err)
	}

	display.Debug("loading current environment: %s", opts.OldEnvName)

	oldEnv, err := GetEnv(opts.OldEnvName)
	if err != nil {
		return nil, fmt.Errorf("failed to load current environment: %w", err)
	}

	oldConfig := oldEnv.EnvConfig

	if opts.Reset {
		display.Info("Reset flag enabled: using default config")

		opts.NewConfig = config.GetDefaultConfig()
	}

	if opts.NewConfig == nil {
		display.Debug("new config not provided, using current environment config")

		opts.NewConfig = &oldConfig
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

	rollbackNeeded := false
	handleFailure := func(msg string, mainErr error) (*Env, error) {
		display.Error("Failed to update environment: %v", mainErr)

		if rollbackNeeded {
			display.Step("Restoring previous environment configuration")
			display.Warn("update failed, rolling back")

			rollbackConfig := oldConfig
			if err := deployStack(true, &rollbackConfig); err != nil {
				display.Error("Failed to rollback environment: %v", err)
			} else {
				display.Done("Rollback completed")
			}
		}

		return nil, fmt.Errorf(msg, mainErr)
	}

	display.Step("Updating stack")

	// If force is set do a docker compose down on the original env
	if opts.Force {
		display.Info("Force flag enabled: stopping old environment with volumes")
		display.Step("Stopping old environment")

		if err := downStack(&oldConfig, true); err != nil {
			return handleFailure("docker compose down failed: %w", err)
		}

		rollbackNeeded = true

		display.Done("Stopped environment: %s", oldConfig.Name)
	}

	// If pullImages is set before deploying pull the images for the updated env
	if opts.PullImages {
		display.Debug("pulling images before stack deployment")

		if err := pullEnvImages(opts.NewConfig); err != nil {
			display.Error("Pulling images failed: %v", err)
			return handleFailure("pulling images failed: %w", err)
		}
	}

	// Deploy the updated compose
	display.Debug("deploying updated stack")
	rollbackNeeded = true

	if err := deployStack(true, opts.NewConfig); err != nil {
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

	newEnv, err := upsertEnvConfig(opts.NewConfig)
	if err != nil {
		return handleFailure("failed to persist environment config: %w", err)
	}

	display.Done("Updated environment: %s", opts.NewConfig.Name)

	return newEnv, nil
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

	if u.Reset && u.NewConfig != nil {
		return fmt.Errorf("cannot specify custom config when Reset is true")
	}

	if u.NewConfig != nil {
		if u.NewConfig.Name != "" && u.NewConfig.Name != u.OldEnvName {
			return fmt.Errorf("config name %q must match environment name %q", u.NewConfig.Name, u.OldEnvName)
		}

		if err := u.NewConfig.Validate(); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}
	}

	if err := EnsureEnvironmentExists(u.OldEnvName); err != nil {
		return fmt.Errorf("no environment with name '%s' exists: %w", u.OldEnvName, err)
	}

	return nil
}
