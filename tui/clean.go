package tui

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showCleanConfirm displays a confirmation dialog for cleaning a Docker environment.
func (a *App) showCleanConfirm() {
	envName := a.SelectedDockerEnv()
	if envName == "" {
		return
	}

	// Create text view for message
	textView := tview.NewTextView().
		SetText("This will permanently delete all data in environment '" + envName + "'.\n\n" + DefaultTheme.DestructiveTag("b") + "This action cannot be undone." + "[-]").
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

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
		AddItem(nil, 0, 1, false).
		AddItem(cleanBtn, 9, 0, true). // "clean" + 4
		AddItem(nil, 2, 0, false).
		AddItem(cancelBtn, 10, 0, true).
		AddItem(nil, 0, 1, false)
	buttonContainer.SetBackgroundColor(tcell.ColorDefault)

	// Main layout
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(buttonContainer, 1, 0, true)
	layout.SetBorder(true).
		SetTitle(" [::b]Clean Environment ").
		SetTitleColor(DefaultTheme.Secondary).
		SetBorderColor(DefaultTheme.Secondary).
		SetBackgroundColor(DefaultTheme.Background).
		SetBorderPadding(1, 1, 1, 1)

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
	a.UpdateFooter("[Clean Environment]", []string{"←→: switch", "enter: confirm", "esc: cancel"})
}

// showCleanProgress displays the cleaning progress with live output.
func (a *App) showCleanProgress(envName string) {
	outputView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() { a.tview.Draw() })
	outputView.SetBorder(true).
		SetTitle(fmt.Sprintf(" [::b]Cleaning: %s ", envName)).
		SetTitleColor(DefaultTheme.Secondary).
		SetBorderColor(DefaultTheme.Secondary).
		SetBorderPadding(0, 0, 2, 2)

	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(DefaultTheme.SecondaryTag("") + "Cleaning... Please wait" + "[-]")

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(outputView, 0, 1, true).
		AddItem(statusBar, 1, 0, false)
	layout.SetBackgroundColor(DefaultTheme.Background)

	// Connect output writer
	a.outputWriter.ClearBuffer()
	a.outputWriter.SetView(a.tview, outputView)

	a.pages.AddAndSwitchToPage("clean-progress", layout, true)
	a.UpdateFooter("[Cleaning]", []string{"please wait..."})

	// Run cleaning in background
	go func() {
		docker, err := dockercore.Clean(dockercore.CleanOpts{
			Name: envName,
		})

		a.tview.QueueUpdateDraw(func() {
			if err != nil {
				statusBar.SetText(fmt.Sprintf("%sClean failed: %v[-]", DefaultTheme.ErrorTag(""), err))
			} else {
				statusBar.SetText(fmt.Sprintf("%sEnvironment cleaned successfully![-] GUI: %s", DefaultTheme.SuccessTag(""), docker.GuiUrl))
			}

			layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
					a.outputWriter.ClearView()
					a.returnFromClean()
					return nil
				}
				return event
			})
			a.UpdateFooter("[Clean Complete]", []string{"esc/enter: back to home"})
		})
	}()
}

// returnFromClean cleans up and returns to the home screen.
func (a *App) returnFromClean() {
	a.pages.RemovePage("clean-confirm")
	a.pages.RemovePage("clean-progress")
	a.pages.SwitchToPage("home")
	a.refreshLists()

	if a.docker.GetItemCount() > 0 {
		a.tview.SetFocus(a.docker)
	} else {
		a.tview.SetFocus(a.dockerEmpty)
	}
	a.UpdateFooter("[Docker Environments]", KeyDescriptions["docker"])
}
