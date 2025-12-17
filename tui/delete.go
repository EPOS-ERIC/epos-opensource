package tui

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showDeleteConfirm displays a confirmation dialog for deleting a Docker environment.
func (a *App) showDeleteConfirm() {
	a.previousFocus = a.tview.GetFocus()
	envName := a.SelectedDockerEnv()
	if envName == "" {
		return
	}

	// Create text view for message
	textView := tview.NewTextView().
		SetText("This will permanently remove all containers, volumes, and associated resources.\n\n" + DefaultTheme.DestructiveTag("b") + "This action cannot be undone." + "[-]").
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	textView.SetBorderPadding(1, 0, 1, 1)

	// Create explicit buttons with styling
	deleteBtn := tview.NewButton("Delete").SetSelectedFunc(func() {
		a.pages.RemovePage("delete-confirm")
		a.showDeleteProgress(envName)
	})
	deleteBtn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Destructive).Foreground(DefaultTheme.OnDestructive))
	deleteBtn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Destructive))

	cancelBtn := tview.NewButton("Cancel").SetSelectedFunc(func() {
		a.returnFromDelete(false)
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
				a.returnFromDelete(false)
				return nil
			}
			return event
		}
	}
	deleteBtn.SetInputCapture(buttonInputCapture(deleteBtn, cancelBtn))
	cancelBtn.SetInputCapture(buttonInputCapture(deleteBtn, cancelBtn))

	buttonContainer := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(deleteBtn, 10, 0, true).
		AddItem(tview.NewBox(), 2, 0, false).
		AddItem(cancelBtn, 10, 0, true).
		AddItem(tview.NewBox(), 0, 1, false)
	buttonContainer.SetBackgroundColor(tcell.ColorDefault)

	// Main layout
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(buttonContainer, 1, 0, true)
	layout.SetBorder(true).
		SetTitle(" [::b]Delete Environment ").
		SetTitleColor(DefaultTheme.Secondary).
		SetBorderColor(DefaultTheme.Destructive).
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

	a.pages.AddPage("delete-confirm", outerLayout, true, true)
	a.tview.SetFocus(deleteBtn)
	a.UpdateFooter("[Delete Environment]", KeyDescriptions["delete-confirm"])
}

// showDeleteProgress displays the deletion progress with live output.
func (a *App) showDeleteProgress(envName string) {
	outputView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() { a.tview.Draw() })
	outputView.SetBorder(true).
		SetTitle(fmt.Sprintf(" [::b]Deleting: %s ", envName)).
		SetTitleColor(DefaultTheme.Secondary).
		SetBorderColor(DefaultTheme.Destructive).
		SetBorderPadding(0, 0, 2, 2)

	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(DefaultTheme.SecondaryTag("") + "Deleting... Please wait" + "[-]")

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(outputView, 0, 1, true).
		AddItem(statusBar, 1, 0, false)
	layout.SetBackgroundColor(DefaultTheme.Background)

	// Connect output writer
	a.outputWriter.ClearBuffer()
	a.outputWriter.SetView(a.tview, outputView)

	a.pages.AddAndSwitchToPage("delete-progress", layout, true)
	a.UpdateFooter("[Deleting]", KeyDescriptions["deleting"])

	// Run deletion in background
	go func() {
		err := dockercore.Delete(dockercore.DeleteOpts{
			Name: []string{envName},
		})

		a.tview.QueueUpdateDraw(func() {
			if err != nil {
				statusBar.SetText(fmt.Sprintf("%sDelete failed: %v[-]", DefaultTheme.ErrorTag(""), err))
				layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
					if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
						a.outputWriter.ClearView()
						a.returnFromDelete(false)
						return nil
					}
					return event
				})
			} else {
				statusBar.SetText(DefaultTheme.SuccessTag("") + "Environment deleted successfully!" + "[-]")
				layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
					if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
						a.outputWriter.ClearView()
						a.returnFromDelete(true)
						return nil
					}
					return event
				})
			}
			a.UpdateFooter("[Delete Complete]", KeyDescriptions["delete-complete"])
		})
	}()
}

// returnFromDelete cleans up and returns to the home screen.
func (a *App) returnFromDelete(clearDetails bool) {
	a.pages.RemovePage("delete-confirm")
	a.pages.RemovePage("delete-progress")
	a.pages.SwitchToPage("home")
	a.refreshLists()

	if clearDetails {
		a.clearDetailsPanel()
		a.tview.SetFocus(a.currentEnv)
	} else {
		if a.previousFocus != nil {
			a.tview.SetFocus(a.previousFocus)
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
