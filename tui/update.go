package tui

import (
	"fmt"
	"strings"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/cmd/k8s/k8score"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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
	a.UpdateFooter(fmt.Sprintf("[%s]", title), "update-form")

	if envName == "" {
		return
	}

	data := &updateFormData{
		name: envName,
	}

	form := NewStyledForm()
	form.AddInputField("Env File", "", 0, nil, func(text string) { data.envFile = text })

	if isDocker {
		form.AddInputField("Compose File", "", 0, nil, func(text string) { data.composeFile = text }).
			AddInputField("Custom Host", "", 0, nil, func(text string) { data.customHost = text }).
			AddCheckbox("Pull Images", false, func(checked bool) { data.pullImages = checked }).
			AddCheckbox("Force (reset DB)", false, func(checked bool) { data.force = checked })
	} else {
		form.AddInputField("Manifest Dir", "", 0, nil, func(text string) { data.manifestDir = text }).
			AddInputField("Custom Host", "", 0, nil, func(text string) { data.customHost = text }).
			AddCheckbox("Force (reset namespace)", false, func(checked bool) { data.force = checked })
	}

	form.AddCheckbox("Reset Config", false, func(checked bool) { data.reset = checked }).
		AddButton("Update", func() { a.handleUpdate(data, isDocker) }).
		AddButton("Cancel", func() {
			a.ResetToHome(ResetOptions{
				PageNames:    []string{"update"},
				RefreshFiles: true,
				RestoreFocus: true,
			})
		})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			a.ResetToHome(ResetOptions{
				PageNames:    []string{"update"},
				RefreshFiles: true,
				RestoreFocus: true,
			})
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

	a.pages.AddPage("update", CenterPrimitive(content, 2, 3), true, true)
	a.currentPage = "update"
	a.tview.SetFocus(form)
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
					CustomHost:  data.customHost,
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
