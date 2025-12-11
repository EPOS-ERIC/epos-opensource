package tui

import (
	"fmt"
	"strings"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// deployFormData holds the form field values.
type deployFormData struct {
	name        string
	envFile     string
	composeFile string
	path        string
	host        string
	pullImages  bool
}

// showDeployForm displays the Docker deployment form.
func (a *App) showDeployForm() {
	a.UpdateFooter("[New Docker Environment]", []string{"tab: next", "enter: submit", "esc: cancel"})

	data := &deployFormData{}

	form := tview.NewForm().
		AddInputField("Name *", "", 40, nil, func(text string) { data.name = text }).
		AddInputField("Env File", "", 40, nil, func(text string) { data.envFile = text }).
		AddInputField("Compose File", "", 40, nil, func(text string) { data.composeFile = text }).
		AddInputField("Path", "", 40, nil, func(text string) { data.path = text }).
		AddInputField("Host", "", 40, nil, func(text string) { data.host = text }).
		AddCheckbox("Update Images", false, func(checked bool) { data.pullImages = checked }).
		AddButton("Deploy", func() { a.handleDeploy(data) }).
		AddButton("Cancel", func() { a.returnFromDeploy() })

	form.SetFieldBackgroundColor(DefaultTheme.Surface)
	form.SetFieldTextColor(DefaultTheme.Secondary)
	form.SetLabelColor(DefaultTheme.Secondary)
	form.SetButtonBackgroundColor(DefaultTheme.Primary)
	form.SetButtonTextColor(DefaultTheme.OnPrimary)
	form.SetButtonActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))
	form.SetBorderPadding(1, 0, 2, 2)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			a.returnFromDeploy()
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
		SetTitle(" [::b]New Docker Environment ").
		SetTitleColor(DefaultTheme.Secondary)

	a.pages.AddPage("deploy", CenterPrimitive(content, 1, 3), true, true)
	a.tview.SetFocus(form)
}

// handleDeploy validates the form and starts deployment.
func (a *App) handleDeploy(data *deployFormData) {
	if strings.TrimSpace(data.name) == "" {
		a.ShowError("Name is required!")
		return
	}
	a.showDeployProgress(data)
}

// showDeployProgress displays the deployment progress with live output.
func (a *App) showDeployProgress(data *deployFormData) {
	outputView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() { a.tview.Draw() })
	outputView.SetBorder(true).
		SetTitle(fmt.Sprintf(" [::b]Deploying: %s ", data.name)).
		SetTitleColor(DefaultTheme.Secondary).
		SetBorderColor(DefaultTheme.Primary).
		SetBorderPadding(0, 0, 2, 2)

	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(DefaultTheme.SecondaryTag("") + "Deploying... Press ESC to go back" + "[-]")

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(outputView, 0, 1, true).
		AddItem(statusBar, 1, 0, false)

	// Connect output writer
	a.outputWriter.ClearBuffer()
	a.outputWriter.SetView(a.tview, outputView)

	layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			a.outputWriter.ClearView()
			a.returnFromDeploy()
			return nil
		}
		return event
	})

	a.pages.AddAndSwitchToPage("deploy-progress", layout, true)
	a.UpdateFooter("[Deploying]", []string{"esc: back (won't stop deployment)"})

	// Run deployment in background
	go func() {
		docker, err := dockercore.Deploy(dockercore.DeployOpts{
			Name:        data.name,
			EnvFile:     data.envFile,
			ComposeFile: data.composeFile,
			Path:        data.path,
			PullImages:  data.pullImages,
			CustomHost:  data.host,
		})

		a.tview.QueueUpdateDraw(func() {
			if err != nil {
				statusBar.SetText(fmt.Sprintf("%sDeployment failed: %v[-]", DefaultTheme.ErrorTag(""), err))
			} else {
				statusBar.SetText(fmt.Sprintf("%sDeployment complete![-] GUI: %s", DefaultTheme.SuccessTag(""), docker.GuiUrl))
			}

			layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
					a.outputWriter.ClearView()
					a.returnFromDeploy()
					return nil
				}
				return event
			})
			a.UpdateFooter("[Deploy Complete]", []string{"esc/enter: back to home"})
		})
	}()
}

// returnFromDeploy cleans up and returns to the home screen.
func (a *App) returnFromDeploy() {
	a.pages.RemovePage("deploy")
	a.pages.RemovePage("deploy-progress")
	a.pages.SwitchToPage("home")
	a.refreshLists()

	if a.docker.GetItemCount() > 0 {
		a.tview.SetFocus(a.docker)
	} else {
		a.tview.SetFocus(a.dockerEmpty)
	}
	a.UpdateFooter("[Docker Environments]", KeyDescriptions["docker"])
}
