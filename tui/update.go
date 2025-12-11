package tui

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showUpdateForm displays a confirmation dialog for updating a Docker environment.
func (a *App) showUpdateForm() {
	a.previousFocus = a.tview.GetFocus()
	envName := a.SelectedDockerEnv()
	if envName == "" {
		return
	}

	// Create text view for message
	textView := tview.NewTextView().
		SetText("\nThis will recreate the environment with new settings.\n\nAny unsaved changes may be lost.").
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	// Create explicit buttons with styling
	updateBtn := tview.NewButton("Update").SetSelectedFunc(func() {
		a.pages.RemovePage("update-confirm")
		a.showUpdateProgress(envName)
	})
	updateBtn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
	updateBtn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))

	cancelBtn := tview.NewButton("Cancel").SetSelectedFunc(func() {
		a.returnFromUpdate()
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
				a.returnFromUpdate()
				return nil
			}
			return event
		}
	}
	updateBtn.SetInputCapture(buttonInputCapture(updateBtn, cancelBtn))
	cancelBtn.SetInputCapture(buttonInputCapture(updateBtn, cancelBtn))

	buttonContainer := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(updateBtn, 10, 0, true).
		AddItem(tview.NewBox(), 2, 0, false).
		AddItem(cancelBtn, 10, 0, true).
		AddItem(tview.NewBox(), 0, 1, false)
	buttonContainer.SetBackgroundColor(tcell.ColorDefault)

	// Main layout
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(buttonContainer, 1, 0, true)
	layout.SetBorder(true).
		SetTitle(" [::b]Update Environment ").
		SetTitleColor(DefaultTheme.Secondary).
		SetBorderColor(DefaultTheme.Primary).
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

	a.pages.AddPage("update-confirm", outerLayout, true, true)
	a.tview.SetFocus(updateBtn)
	a.UpdateFooter("[Update Environment]", KeyDescriptions["update-confirm"])
}

// showUpdateProgress displays the update progress with live output.
func (a *App) showUpdateProgress(envName string) {
	outputView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() { a.tview.Draw() })
	outputView.SetBorder(true).
		SetTitle(fmt.Sprintf(" [::b]Updating: %s ", envName)).
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
			Name: envName,
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
	a.pages.RemovePage("update-confirm")
	a.pages.RemovePage("update-progress")
	a.pages.SwitchToPage("home")
	a.refreshLists()

	if a.previousFocus != nil {
		a.tview.SetFocus(a.previousFocus)
	}
}
