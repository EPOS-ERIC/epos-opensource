package tui

import (
	"fmt"
	"time"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker"
	dockerconfig "github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"
	k8sconfig "github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
	"github.com/EPOS-ERIC/epos-opensource/validate"
)

// deployFormData holds the form field values.
type deployFormData struct {
	name          string
	pullImages    bool   // Docker only
	context       string // K8s only
	configSession *configEditSession
}

func (d *deployFormData) cleanupConfigSession() {
	if d.configSession != nil {
		_ = d.configSession.Cleanup()
		d.configSession = nil
	}
}

// showDeployForm displays the deployment form for Docker or K8s.
func (a *App) showDeployForm() {
	a.PushFocus()
	isDocker := a.envList.IsDockerActive()
	title := "New Docker Environment"
	if !isDocker {
		title = "New K8s Environment"
	}
	a.UpdateFooter(DeployFormKey)

	data := &deployFormData{}

	fields := []FormField{
		{Type: "input", Label: "Name *", InputChangedFunc: func(text string) { data.name = text }},
	}

	if isDocker {
		fields = append(fields,
			FormField{
				Type:                "checkbox",
				Label:               "Update Images",
				CheckboxChangedFunc: func(checked bool) { data.pullImages = checked },
			},
		)
	} else {
		currentContext := ""
		if ctx, err := common.GetCurrentKubeContext(); err == nil {
			currentContext = ctx
		}
		data.context = currentContext

		contexts, err := common.GetKubeContexts()
		if err != nil {
			fields = append(fields, FormField{
				Type:             "input",
				Label:            "Context",
				Value:            currentContext,
				InputChangedFunc: func(text string) { data.context = text },
			})
		} else {
			fields = append(fields, FormField{
				Type:         "dropdown",
				Label:        "Context",
				Value:        currentContext,
				Options:      contexts,
				SelectedFunc: func(option string, index int) { data.context = option },
			})
		}
		// Keep K8s form minimal: most settings are edited via config file.
	}

	buttons := []FormButton{
		{Label: "Deploy", SelectedFunc: func() { a.handleDeploy(data, isDocker) }},
		{Label: "Cancel", SelectedFunc: func() {
			data.cleanupConfigSession()
			a.ResetToHome(ResetOptions{
				PageNames:    []string{"deploy"},
				RestoreFocus: true,
			})
		}},
	}

	height := 16
	if !isDocker {
		height = 20
	}
	opts := ModalFormOptions{
		PageName: "deploy",
		Title:    title,
		Fields:   fields,
		Buttons:  buttons,
		BottomButton: &FormButton{
			Label:        "Edit Config",
			SelectedFunc: func() { a.editDeployConfig(data, isDocker) },
		},
		OnCancel: func() {
			data.cleanupConfigSession()
			a.ResetToHome(ResetOptions{
				PageNames:    []string{"deploy"},
				RestoreFocus: true,
			})
		},
		Height: height,
	}

	a.ShowModalForm(opts)
}

// handleDeploy validates the form and starts deployment.
func (a *App) handleDeploy(data *deployFormData, isDocker bool) {
	if err := validate.Name(data.name); err != nil {
		a.ShowError(err.Error())
		return
	}

	dockerCfg, k8sCfg, err := a.buildDeployConfig(data, isDocker)
	if err != nil {
		a.ShowError(err.Error())
		return
	}

	data.cleanupConfigSession()
	a.showDeployProgress(data, isDocker, dockerCfg, k8sCfg)
}

// showDeployProgress displays the deployment progress with live output.
func (a *App) showDeployProgress(data *deployFormData, isDocker bool, dockerCfg *dockerconfig.EnvConfig, k8sCfg *k8sconfig.Config) {
	a.RunBackgroundTask(TaskOptions{
		Operation: "Deploy",
		EnvName:   data.name,
		IsDocker:  isDocker,
		Task: func() (string, error) {
			var err error
			var guiURL string

			if isDocker {
				env, derr := docker.Deploy(docker.DeployOpts{
					PullImages: data.pullImages,
					Config:     dockerCfg,
				})
				err = derr
				if env != nil {
					urls, uerr := env.BuildEnvURLs()
					if uerr != nil {
						return "", uerr
					}
					guiURL = urls.GUIURL
				}
			} else {
				env, kerr := k8s.Deploy(k8s.DeployOpts{
					Context: data.context,
					Config:  k8sCfg,
				})
				err = kerr
				if env != nil {
					urls, uerr := env.BuildEnvURLs()
					if uerr != nil {
						return "", uerr
					}
					guiURL = urls.GUIURL
				}
			}

			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Deployment complete! GUI: %s", guiURL), nil
		},
	})
}

func (a *App) editDeployConfig(data *deployFormData, isDocker bool) {
	if data.configSession == nil {
		fileName := "docker-config.yaml"
		if !isDocker {
			fileName = "k8s-config.yaml"
		}

		session, err := newConfigEditSession(fileName)
		if err != nil {
			a.ShowError(err.Error())
			return
		}

		if isDocker {
			seed := dockerconfig.GetDefaultConfig()
			if data.name != "" {
				seed.Name = data.name
			}

			if err := seed.Save(session.FilePath()); err != nil {
				_ = session.Cleanup()
				a.ShowError(err.Error())
				return
			}
		} else {
			seed := k8sconfig.GetDefaultConfig()
			if data.name != "" {
				seed.Name = data.name
			}

			if err := seed.Save(session.FilePath()); err != nil {
				_ = session.Cleanup()
				a.ShowError(err.Error())
				return
			}
		}

		data.configSession = session
	}

	if err := a.openConfigEditor(data.configSession.FilePath()); err != nil {
		a.ShowError(err.Error())
		return
	}

	a.FlashMessage("Config opened. Save changes before deploying.", 2*time.Second)
}

func (a *App) buildDeployConfig(data *deployFormData, isDocker bool) (*dockerconfig.EnvConfig, *k8sconfig.Config, error) {
	if isDocker {
		cfg := dockerconfig.GetDefaultConfig()
		if data.configSession != nil {
			loadedCfg, err := dockerconfig.LoadConfig(data.configSession.FilePath())
			if err != nil {
				return nil, nil, fmt.Errorf("failed to load docker config from %s: %w", data.configSession.FilePath(), err)
			}

			cfg = loadedCfg
		}

		cfg.Name = data.name
		if err := cfg.Validate(); err != nil {
			return nil, nil, fmt.Errorf("invalid docker config: %w", err)
		}

		return cfg, nil, nil
	}

	cfg := k8sconfig.GetDefaultConfig()
	if data.configSession != nil {
		loadedCfg, err := k8sconfig.LoadConfig(data.configSession.FilePath())
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load k8s config from %s: %w", data.configSession.FilePath(), err)
		}

		cfg = loadedCfg
	}

	cfg.Name = data.name
	if err := cfg.Validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid k8s config: %w", err)
	}

	return nil, cfg, nil
}
