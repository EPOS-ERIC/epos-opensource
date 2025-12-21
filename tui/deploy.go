package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/cmd/k8s/k8score"
	"github.com/epos-eu/epos-opensource/command"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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
}

// showDeployForm displays the deployment form for Docker or K8s.
func (a *App) showDeployForm() {
	a.PushFocus()
	isDocker := a.envList.IsDockerActive()
	title := "New Docker Environment"
	footer := "deploy-form"
	if !isDocker {
		title = "New Kubernetes Environment"
	}
	a.UpdateFooter(fmt.Sprintf("[%s]", title), KeyDescriptions[footer])

	data := &deployFormData{
		protocol: "http",
	}

	form := NewStyledForm()
	form.AddInputField("Name *", "", 0, nil, func(text string) { data.name = text })

	if isDocker {
		form.AddInputField("Env File", "", 0, nil, func(text string) { data.envFile = text }).
			AddInputField("Compose File", "", 0, nil, func(text string) { data.composeFile = text }).
			AddInputField("Path", "", 0, nil, func(text string) { data.path = text }).
			AddInputField("Host", "", 0, nil, func(text string) { data.host = text }).
			AddCheckbox("Update Images", false, func(checked bool) { data.pullImages = checked })
	} else {
		currentContext := ""
		if out, err := command.RunCommand(exec.Command("kubectl", "config", "current-context"), true); err == nil {
			currentContext = strings.TrimSpace(string(out))
		}
		data.context = currentContext

		form.AddInputField("Context", currentContext, 0, nil, func(text string) { data.context = text }).
			AddInputField("Env File", "", 0, nil, func(text string) { data.envFile = text }).
			AddInputField("Manifest Dir", "", 0, nil, func(text string) { data.manifestDir = text }).
			AddInputField("Path", "", 0, nil, func(text string) { data.path = text }).
			AddInputField("Host", "", 0, nil, func(text string) { data.host = text }).
			AddDropDown("Protocol", []string{"http", "https"}, 0, func(option string, optionIndex int) {
				data.protocol = option
			})
	}

	form.AddButton("Deploy", func() { a.handleDeploy(data, isDocker) }).
		AddButton("Cancel", func() { a.returnFromDeploy() })

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			a.returnFromDeploy()
			return nil
		}
		return event
	})

	content := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true)

	content.SetBorder(true).
		SetBorderColor(DefaultTheme.Primary).
		SetTitle(fmt.Sprintf(" [::b]%s ", title)).
		SetTitleColor(DefaultTheme.Secondary)

	a.pages.AddPage("deploy", CenterPrimitive(content, 2, 3), true, true)
	a.tview.SetFocus(form)
}

// handleDeploy validates the form and starts deployment.
func (a *App) handleDeploy(data *deployFormData, isDocker bool) {
	if strings.TrimSpace(data.name) == "" {
		a.ShowError("Name is required!")
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
				docker, derr := dockercore.Deploy(dockercore.DeployOpts{
					Name:        data.name,
					EnvFile:     data.envFile,
					ComposeFile: data.composeFile,
					Path:        data.path,
					PullImages:  data.pullImages,
					CustomHost:  data.host,
				})
				err = derr
				if docker != nil {
					guiURL = docker.GuiUrl
				}
			} else {
				k8s, kerr := k8score.Deploy(k8score.DeployOpts{
					Name:        data.name,
					EnvFile:     data.envFile,
					ManifestDir: data.manifestDir,
					Path:        data.path,
					Context:     data.context,
					Protocol:    data.protocol,
					CustomHost:  data.host,
				})
				err = kerr
				if k8s != nil {
					guiURL = k8s.GuiUrl
				}
			}

			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Deployment complete! GUI: %s", guiURL), nil
		},
	})
}

// returnFromDeploy cleans up and returns to the home screen.
func (a *App) returnFromDeploy() {
	a.ResetToHome(ResetOptions{
		PageNames:    []string{"deploy"},
		RestoreFocus: true,
	})
}
