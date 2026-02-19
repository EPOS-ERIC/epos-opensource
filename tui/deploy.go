package tui

import (
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"
	"github.com/EPOS-ERIC/epos-opensource/validate"
)

// deployFormData holds the form field values.
type deployFormData struct {
	name        string
	envFile     string
	composeFile string // Docker
	manifestDir string // K8s
	path        string
	host        string
	pullImages  bool   // Docker
	context     string // K8s
	protocol    string // K8s
	tlsEnabled  bool   // K8s
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

	data := &deployFormData{
		protocol: "http",
	}

	fields := []FormField{
		{Type: "input", Label: "Name *", InputChangedFunc: func(text string) { data.name = text }},
	}

	if isDocker {
		fields = append(fields,
			FormField{
				Type:             "input",
				Label:            "Env File",
				InputChangedFunc: func(text string) { data.envFile = text },
			},
			FormField{
				Type:             "input",
				Label:            "Compose File",
				InputChangedFunc: func(text string) { data.composeFile = text },
			},
			FormField{
				Type:             "input",
				Label:            "Path",
				InputChangedFunc: func(text string) { data.path = text },
			},
			FormField{
				Type:             "input",
				Label:            "Host",
				InputChangedFunc: func(text string) { data.host = text },
			},
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
		fields = append(fields,
			FormField{
				Type:             "input",
				Label:            "Env File",
				InputChangedFunc: func(text string) { data.envFile = text },
			},
			FormField{
				Type:             "input",
				Label:            "Manifest Dir",
				InputChangedFunc: func(text string) { data.manifestDir = text },
			},
			FormField{
				Type:             "input",
				Label:            "Path",
				InputChangedFunc: func(text string) { data.path = text },
			},
			FormField{
				Type:             "input",
				Label:            "Host",
				InputChangedFunc: func(text string) { data.host = text },
			},
			FormField{
				Type:         "dropdown",
				Label:        "Protocol",
				Value:        "http",
				Options:      []string{"http", "https"},
				SelectedFunc: func(option string, index int) { data.protocol = option },
			},
			FormField{
				Type:                "checkbox",
				Label:               "Use TLS Manifest",
				CheckboxChangedFunc: func(checked bool) { data.tlsEnabled = checked },
			},
		)
	}

	buttons := []FormButton{
		{Label: "Deploy", SelectedFunc: func() { a.handleDeploy(data, isDocker) }},
		{Label: "Cancel", SelectedFunc: func() {
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
		OnCancel: func() {
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
	a.showDeployProgress(data, isDocker)
}

// showDeployProgress displays the deployment progress with live output.
func (a *App) showDeployProgress(data *deployFormData, isDocker bool) {
	a.RunBackgroundTask(TaskOptions{
		Operation: "Deploy",
		EnvName:   data.name,
		IsDocker:  isDocker,
		Task: func() (string, error) {
			var err error
			var guiURL string

			if isDocker {
				env, derr := docker.Deploy(docker.DeployOpts{
					// Name:       data.name,
					Path:       data.path,
					PullImages: data.pullImages,
				})
				err = derr
				if env != nil {
					guiURL = env.GuiUrl
				}
			} else {
				env, kerr := k8s.Deploy(k8s.DeployOpts{
					// Name:        data.name,
					// EnvFile:     data.envFile,
					// ManifestDir: data.manifestDir,
					// Path:    data.path,
					Context: data.context,
					// Protocol:    data.protocol,
					// CustomHost:  data.host,
					// TLSEnabled:  data.tlsEnabled,
				})
				err = kerr
				if env != nil {
					// guiURL = env.GuiUrl
				}
			}

			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Deployment complete! GUI: %s", guiURL), nil
		},
	})
}
