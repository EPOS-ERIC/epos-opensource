package tui

import (
	"strings"

	"github.com/EPOS-ERIC/epos-opensource/cmd/docker/dockercore"
	"github.com/EPOS-ERIC/epos-opensource/cmd/k8s/k8score"
)

// updateFormData holds the form field values.
type updateFormData struct {
	name        string
	envFile     string
	composeFile string // Docker
	manifestDir string // K8s
	customHost  string
	pullImages  bool // Docker
	force       bool // Docker: reset DB; K8s: reset namespace
	reset       bool // Reset config to defaults
}

// showUpdateForm displays the environment update form for Docker or K8s.
func (a *App) showUpdateForm() {
	a.PushFocus()
	envName, isDocker := a.envList.GetSelected()
	title := "Update Docker Environment"
	if !isDocker {
		title = "Update K8s Environment"
	}
	a.UpdateFooter(UpdateFormKey)

	if envName == "" {
		return
	}

	data := &updateFormData{
		name: envName,
	}

	fields := []FormField{
		{Type: "input", Label: "Env File", InputChangedFunc: func(text string) { data.envFile = text }},
	}

	if isDocker {
		fields = append(fields,
			FormField{Type: "input", Label: "Compose File", InputChangedFunc: func(text string) { data.composeFile = text }},
			FormField{Type: "input", Label: "Custom Host", InputChangedFunc: func(text string) { data.customHost = text }},
			FormField{Type: "checkbox", Label: "Pull Images", CheckboxChangedFunc: func(checked bool) { data.pullImages = checked }},
			FormField{Type: "checkbox", Label: "Force (reset DB)", CheckboxChangedFunc: func(checked bool) { data.force = checked }},
		)
	} else {
		fields = append(fields,
			FormField{Type: "input", Label: "Manifest Dir", InputChangedFunc: func(text string) { data.manifestDir = text }},
			FormField{Type: "input", Label: "Custom Host", InputChangedFunc: func(text string) { data.customHost = text }},
			FormField{Type: "checkbox", Label: "Force (reset namespace)", CheckboxChangedFunc: func(checked bool) { data.force = checked }},
		)
	}

	fields = append(fields, FormField{Type: "checkbox", Label: "Reset Config", CheckboxChangedFunc: func(checked bool) { data.reset = checked }})

	buttons := []FormButton{
		{Label: "Update", SelectedFunc: func() { a.handleUpdate(data, isDocker) }},
		{Label: "Cancel", SelectedFunc: func() {
			a.ResetToHome(ResetOptions{
				PageNames:    []string{"update"},
				RefreshFiles: true,
				RestoreFocus: true,
			})
		}},
	}

	opts := ModalFormOptions{
		PageName: "update",
		Title:    title,
		Fields:   fields,
		Buttons:  buttons,
		OnCancel: func() {
			a.ResetToHome(ResetOptions{
				PageNames:    []string{"update"},
				RefreshFiles: true,
				RestoreFocus: true,
			})
		},
	}

	a.ShowModalForm(opts)
}

// handleUpdate validates the form and starts update.
func (a *App) handleUpdate(data *updateFormData, isDocker bool) {
	// Basic validation - prevent reset with custom files
	if data.reset && (strings.TrimSpace(data.envFile) != "" || strings.TrimSpace(data.composeFile) != "" || strings.TrimSpace(data.manifestDir) != "") {
		a.ShowError("Cannot specify custom files when Reset Config is checked!")
		return
	}

	a.showUpdateProgress(data, isDocker)
}

// showUpdateProgress displays the update progress with live output.
func (a *App) showUpdateProgress(data *updateFormData, isDocker bool) {
	a.RunBackgroundTask(TaskOptions{
		Operation: "Update",
		EnvName:   data.name,
		IsDocker:  isDocker,
		Task: func() (string, error) {
			var err error
			if isDocker {
				_, err = dockercore.Update(dockercore.UpdateOpts{
					Name:        data.name,
					EnvFile:     data.envFile,
					ComposeFile: data.composeFile,
					PullImages:  data.pullImages,
					Force:       data.force,
					Reset:       data.reset,
				})
			} else {
				_, err = k8score.Update(k8score.UpdateOpts{
					Name:        data.name,
					EnvFile:     data.envFile,
					ManifestDir: data.manifestDir,
					Force:       data.force,
					CustomHost:  data.customHost,
					Reset:       data.reset,
				})
			}
			return "", err
		},
	})
}
