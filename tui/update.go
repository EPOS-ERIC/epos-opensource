package tui

import (
	"fmt"
	"strings"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// updateFormData holds the form field values.
type updateFormData struct {
	name        string
	envFile     string
	composeFile string
	customHost  string
	pullImages  bool
	force       bool
	reset       bool
}

// showUpdateForm displays the Docker environment update form.
func (a *App) showUpdateForm() {
	a.UpdateFooter("[Update Docker Environment]", KeyDescriptions["update-form"])

	envName := a.SelectedDockerEnv()
	if envName == "" {
		return
	}

	data := &updateFormData{
		name: envName,
	}

	form := tview.NewForm().
		AddInputField("Env File", "", 0, nil, func(text string) { data.envFile = text }).
		AddInputField("Compose File", "", 0, nil, func(text string) { data.composeFile = text }).
		AddInputField("Custom Host", "", 0, nil, func(text string) { data.customHost = text }).
		AddCheckbox("Pull Images", false, func(checked bool) { data.pullImages = checked }).
		AddCheckbox("Force (reset DB)", false, func(checked bool) { data.force = checked }).
		AddCheckbox("Reset Config", false, func(checked bool) { data.reset = checked }).
		AddButton("Update", func() { a.handleUpdate(data) }).
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
		SetTitle(" [::b]Update Docker Environment ").
		SetTitleColor(DefaultTheme.Secondary)

	a.pages.AddPage("update", CenterPrimitive(content, 2, 3), true, true)
	a.tview.SetFocus(form)
}

// handleUpdate validates the form and starts update.
func (a *App) handleUpdate(data *updateFormData) {
	// Basic validation - prevent reset with custom files
	if data.reset && (strings.TrimSpace(data.envFile) != "" || strings.TrimSpace(data.composeFile) != "") {
		a.ShowError("Cannot specify custom files when Reset Config is checked!")
		return
	}

	a.showUpdateProgress(data)
}

// showUpdateProgress displays the update progress with live output.
func (a *App) showUpdateProgress(data *updateFormData) {
	outputView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() { a.tview.Draw() })
	outputView.SetBorder(true).
		SetTitle(fmt.Sprintf(" [::b]Updating: %s ", data.name)).
		SetTitleColor(DefaultTheme.Secondary).
		SetBorderColor(DefaultTheme.Primary).
		SetBorderPadding(0, 0, 2, 2)

	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(DefaultTheme.SecondaryTag("") + "Updating... Please wait" + "[-]")

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(outputView, 0, 1, true).
		AddItem(statusBar, 1, 0, false)
	layout.SetBackgroundColor(DefaultTheme.Background)

	// Connect output writer
	a.outputWriter.ClearBuffer()
	a.outputWriter.SetView(a.tview, outputView)

	a.pages.AddAndSwitchToPage("update-progress", layout, true)
	a.UpdateFooter("[Updating]", KeyDescriptions["updating"])

	// Run update in background
	go func() {
		_, err := dockercore.Update(dockercore.UpdateOpts{
			Name:        data.name,
			EnvFile:     data.envFile,
			ComposeFile: data.composeFile,
			PullImages:  data.pullImages,
			Force:       data.force,
			CustomHost:  data.customHost,
			Reset:       data.reset,
		})

		a.tview.QueueUpdateDraw(func() {
			if err != nil {
				statusBar.SetText(fmt.Sprintf("%sUpdate failed: %v[-]", DefaultTheme.ErrorTag(""), err))
			} else {
				statusBar.SetText(DefaultTheme.SuccessTag("") + "Environment updated successfully!" + "[-]")
			}

			layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
					a.outputWriter.ClearView()
					a.returnFromUpdate()
					return nil
				}
				return event
			})
			a.UpdateFooter("[Update Complete]", KeyDescriptions["update-complete"])
		})
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
