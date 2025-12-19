package tui

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showCleanConfirm displays a confirmation dialog for cleaning a Docker environment.
func (a *App) showCleanConfirm() {
	a.previousFocus = a.tview.GetFocus()
	envName := a.SelectedDockerEnv()
	if envName == "" {
		return
	}

	// Create text view for message
	textView := tview.NewTextView().
		SetText("This will permanently delete all data in environment '" + envName + "'.\n\n" + DefaultTheme.DestructiveTag("b") + "This action cannot be undone." + "[-]").
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	textView.SetBorderPadding(1, 0, 1, 1)

	// Create explicit buttons with styling
	cleanBtn := tview.NewButton("Clean").SetSelectedFunc(func() {
		a.pages.RemovePage("clean-confirm")
		a.showCleanProgress(envName)
	})
	cleanBtn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Destructive).Foreground(DefaultTheme.OnDestructive))
	cleanBtn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Destructive))

	cancelBtn := tview.NewButton("Cancel").SetSelectedFunc(func() {
		a.returnFromClean()
	})
	cancelBtn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
	cancelBtn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))

	// Helper to handle arrow key navigation between buttons
	buttonInputCapture := func(leftBtn, rightBtn *tview.Button) func(*tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyLeft:
				a.tview.SetFocus(leftBtn)
				return nil
			case tcell.KeyRight:
				a.tview.SetFocus(rightBtn)
				return nil
			case tcell.KeyEsc:
				a.returnFromClean()
				return nil
			}
			return event
		}
	}
	cleanBtn.SetInputCapture(buttonInputCapture(cleanBtn, cancelBtn))
	cancelBtn.SetInputCapture(buttonInputCapture(cleanBtn, cancelBtn))

	buttonContainer := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(cleanBtn, 9, 0, true). // "clean" + 4
		AddItem(tview.NewBox(), 2, 0, false).
		AddItem(cancelBtn, 10, 0, true).
		AddItem(tview.NewBox(), 0, 1, false)
	buttonContainer.SetBackgroundColor(tcell.ColorDefault)

	// Main layout
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(buttonContainer, 1, 0, true)
	layout.SetBorder(true).
		SetTitle(" [::b]Clean Environment ").
		SetTitleColor(DefaultTheme.Secondary).
		SetBorderColor(DefaultTheme.Secondary).
		SetBackgroundColor(DefaultTheme.Background)

	// Center the layout
	innerFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(layout, 11, 1, true).
		AddItem(nil, 0, 1, false)
	innerFlex.SetBackgroundColor(DefaultTheme.Background)

	outerLayout := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(innerFlex, 60, 1, true).
		AddItem(nil, 0, 1, false)
	outerLayout.SetBackgroundColor(DefaultTheme.Background)

	a.pages.AddPage("clean-confirm", outerLayout, true, true)
	a.tview.SetFocus(cleanBtn)
	a.UpdateFooter("[Clean Environment]", KeyDescriptions["clean-confirm"])
}

// showCleanProgress displays the cleaning progress with live output.
func (a *App) showCleanProgress(envName string) {
	progress := NewOperationProgress(a, "Clean", envName)
	progress.Start()

	// Run cleaning in background
	go func() {
		docker, err := dockercore.Clean(dockercore.CleanOpts{
			Name: envName,
		})

		if err != nil {
			progress.Complete(false, err.Error())
		} else {
			successMsg := fmt.Sprintf("Environment cleaned successfully! GUI: %s", docker.GuiUrl)
			progress.Complete(true, successMsg)
		}
	}()
}

// returnFromClean cleans up and returns to the home screen.
func (a *App) returnFromClean() {
	a.pages.RemovePage("clean-confirm")
	a.pages.RemovePage("clean-progress")
	a.pages.SwitchToPage("home")
	a.refreshLists()
	a.refreshIngestedFiles()

	if a.previousFocus != nil {
		a.tview.SetFocus(a.previousFocus)
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
