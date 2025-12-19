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
	a.UpdateFooter("[New Docker Environment]", KeyDescriptions["deploy-form"])

	data := &deployFormData{}

	form := tview.NewForm().
		AddInputField("Name *", "", 0, nil, func(text string) { data.name = text }).
		AddInputField("Env File", "", 0, nil, func(text string) { data.envFile = text }).
		AddInputField("Compose File", "", 0, nil, func(text string) { data.composeFile = text }).
		AddInputField("Path", "", 0, nil, func(text string) { data.path = text }).
		AddInputField("Host", "", 0, nil, func(text string) { data.host = text }).
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

	a.pages.AddPage("deploy", CenterPrimitive(content, 2, 3), true, true)
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
	progress := NewOperationProgress(a, "Deploy", data.name)
	progress.Start()

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

		if err != nil {
			progress.Complete(false, err.Error())
		} else {
			successMsg := fmt.Sprintf("Deployment complete! GUI: %s", docker.GuiUrl)
			progress.Complete(true, successMsg)
		}
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
