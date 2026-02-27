package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/EPOS-ERIC/epos-opensource/command"
	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
)

type composeBundle struct {
	dir         string
	envFile     string
	composeFile string
}

func createComposeBundle(cfg *config.EnvConfig) (*composeBundle, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	files, err := cfg.Render()
	if err != nil {
		return nil, fmt.Errorf("failed to render docker templates: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "epos-docker-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary compose bundle directory: %w", err)
	}

	bundle := &composeBundle{
		dir:         tmpDir,
		envFile:     filepath.Join(tmpDir, ".env"),
		composeFile: filepath.Join(tmpDir, "docker-compose.yaml"),
	}

	if err := common.CreateFileWithContent(bundle.envFile, files[".env"], true); err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to write temporary .env file: %w", err)
	}

	if err := common.CreateFileWithContent(bundle.composeFile, files["docker-compose.yaml"], true); err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to write temporary docker-compose file: %w", err)
	}

	return bundle, nil
}

func withComposeBundle(cfg *config.EnvConfig, fn func(bundle *composeBundle) error) error {
	bundle, err := createComposeBundle(cfg)
	if err != nil {
		return err
	}

	defer func() {
		if err := os.RemoveAll(bundle.dir); err != nil {
			display.Warn("failed to cleanup temporary compose bundle %s: %v", bundle.dir, err)
		}
	}()

	if err := fn(bundle); err != nil {
		return err
	}

	return nil
}

// composeCommand creates a docker compose command configured with explicit files and project name.
func composeCommand(bundle *composeBundle, projectName string, args ...string) *exec.Cmd {
	baseArgs := []string{
		"compose",
		"-p",
		projectName,
		"--env-file",
		bundle.envFile,
		"-f",
		bundle.composeFile,
	}

	cmd := exec.Command("docker", append(baseArgs, args...)...)
	cmd.Dir = bundle.dir
	return cmd
}

// pullEnvImages pulls docker images for the environment with custom messages
// TODO make this take just the image struct and pull those, not using the env file
func pullEnvImages(cfg *config.EnvConfig) error {
	display.Step("Pulling images for environment: %s", cfg.Name)

	err := withComposeBundle(cfg, func(bundle *composeBundle) error {
		if _, err := command.RunCommand(composeCommand(bundle, cfg.Name, "pull"), false); err != nil {
			return fmt.Errorf("pull images failed: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	display.Done("Images pulled for environment: %s", cfg.Name)

	return nil
}

// deployStack deploys the stack using a rendered temporary compose bundle.
// The deployed stack will use the configured ports for the deployment of the services.
// It is the responsibility of the caller to ensure the ports are not used.
func deployStack(removeOrphans bool, cfg *config.EnvConfig) error {
	display.Step("Deploying stack")

	args := []string{"up", "-d"}
	if removeOrphans {
		args = append(args, "--remove-orphans")
	}

	err := withComposeBundle(cfg, func(bundle *composeBundle) error {
		if _, err := command.RunCommand(composeCommand(bundle, cfg.Name, args...), false); err != nil {
			return fmt.Errorf("deployment of stack failed: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	display.Done("Stack deployed successfully")

	return nil
}

// downStack stops the stack for the given environment configuration.
func downStack(cfg *config.EnvConfig, removeVolumes bool) error {
	err := withComposeBundle(cfg, func(bundle *composeBundle) error {
		args := []string{"down"}
		if removeVolumes {
			args = append(args, "-v")
		}

		if _, err := command.RunCommand(composeCommand(bundle, cfg.Name, args...), false); err != nil {
			return fmt.Errorf("docker compose down failed: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
