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
	isDocker := a.currentEnv == a.dockerFlex
	title := "Update Docker Environment"
	if !isDocker {
		title = "Update Kubernetes Environment"
	}
	a.UpdateFooter(fmt.Sprintf("[%s]", title), KeyDescriptions["update-form"])

	envName := ""
	if isDocker {
		envName = a.SelectedDockerEnv()
	} else {
		envName = a.SelectedK8sEnv()
	}

	if envName == "" {
		return
	}

	data := &updateFormData{
		name: envName,
	}

	form := tview.NewForm().
		AddInputField("Env File", "", 0, nil, func(text string) { data.envFile = text })

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
		AddButton("Cancel", func() { a.returnFromUpdate() })

	form.SetFieldBackgroundColor(DefaultTheme.Surface)
	form.SetFieldTextColor(DefaultTheme.Secondary)
	form.SetLabelColor(DefaultTheme.Secondary)
	form.SetButtonBackgroundColor(DefaultTheme.Primary)
	form.SetButtonTextColor(DefaultTheme.OnPrimary)
	form.SetButtonActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))
	form.SetBorderPadding(1, 0, 2, 2)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			a.returnFromUpdate()
			return nil
		}
		return event
	})

	// Layout
	content := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true)

	content.SetBorder(true).
		SetBorderColor(DefaultTheme.Primary).
		SetTitle(fmt.Sprintf(" [::b]%s ", title)).
		SetTitleColor(DefaultTheme.Secondary)

	a.pages.AddPage("update", CenterPrimitive(content, 2, 3), true, true)
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
	progress := NewOperationProgress(a, "Update", data.name)
	progress.Start()

	// Run update in background
	go func() {
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

		if err != nil {
			progress.Complete(false, err.Error())
		} else {
			progress.Complete(true, "Environment updated successfully!")
		}
	}()
}

// returnFromUpdate cleans up and returns to the home screen.
func (a *App) returnFromUpdate() {
	a.pages.RemovePage("update")
	a.pages.RemovePage("update-progress")
	a.pages.SwitchToPage("home")
	a.refreshLists()
	a.refreshIngestedFiles()

	if a.detailsShown {
		a.tview.SetFocus(a.details)
	} else {
		if a.currentEnv == a.dockerFlex {
			a.tview.SetFocus(a.docker)
		} else {
			a.tview.SetFocus(a.k8s)
		}
	}
	if a.detailsShown {
		key := DetailsK8sKey
		if a.currentEnv == a.dockerFlex {
			key = DetailsDockerKey
		}
		a.UpdateFooter("[Environment Details]", KeyDescriptions[key])
	} else {
		if a.currentEnv == a.dockerFlex {
			a.UpdateFooter("[Docker Environments]", KeyDescriptions["docker"])
		} else {
			a.UpdateFooter("[K8s Environments]", KeyDescriptions["k8s"])
		}
	}
}
