package tui

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showPopulateForm displays a confirmation dialog for populating a Docker environment.
func (a *App) showPopulateForm() {
	a.actionFromDetails = a.detailsShown
	envName := a.SelectedDockerEnv()
	if envName == "" {
		return
	}

	// Create text view for message
	textView := tview.NewTextView().
		SetText("This will ingest example data into the environment.\n\nThis may take some time depending on the data size.").
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	textView.SetBorderPadding(1, 0, 1, 1)

	// Create explicit buttons with styling
	populateBtn := tview.NewButton("Populate").SetSelectedFunc(func() {
		a.pages.RemovePage("populate-confirm")
		a.showPopulateProgress(envName)
	})
	populateBtn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
	populateBtn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))

	cancelBtn := tview.NewButton("Cancel").SetSelectedFunc(func() {
		a.returnFromPopulate()
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
				a.returnFromPopulate()
				return nil
			}
			return event
		}
	}
	populateBtn.SetInputCapture(buttonInputCapture(populateBtn, cancelBtn))
	cancelBtn.SetInputCapture(buttonInputCapture(populateBtn, cancelBtn))

	buttonContainer := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(populateBtn, 10, 0, true).
		AddItem(nil, 2, 0, false).
		AddItem(cancelBtn, 10, 0, true).
		AddItem(nil, 0, 1, false)
	buttonContainer.SetBackgroundColor(tcell.ColorDefault)

	// Main layout
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(buttonContainer, 1, 0, true)
	layout.SetBorder(true).
		SetTitle(" [::b]Populate Environment ").
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

	a.pages.AddPage("populate-confirm", outerLayout, true, true)
	a.tview.SetFocus(populateBtn)
	a.UpdateFooter("[Populate Environment]", KeyDescriptions["populate-confirm"])
}

// showPopulateProgress displays the populate progress with live output.
func (a *App) showPopulateProgress(envName string) {
	outputView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() { a.tview.Draw() })
	outputView.SetBorder(true).
		SetTitle(fmt.Sprintf(" [::b]Populating: %s ", envName)).
		SetTitleColor(DefaultTheme.Secondary).
		SetBorderColor(DefaultTheme.Primary).
		SetBorderPadding(0, 0, 2, 2)

	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(DefaultTheme.SecondaryTag("") + "Populating... Please wait" + "[-]")

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(outputView, 0, 1, true).
		AddItem(statusBar, 1, 0, false)
	layout.SetBackgroundColor(DefaultTheme.Background)

	// Connect output writer
	a.outputWriter.ClearBuffer()
	a.outputWriter.SetView(a.tview, outputView)

	a.pages.AddAndSwitchToPage("populate-progress", layout, true)
	a.UpdateFooter("[Populating]", KeyDescriptions["populating"])

	// Run populate in background
	go func() {
		_, err := dockercore.Populate(dockercore.PopulateOpts{
			Name:             envName,
			PopulateExamples: true,
		})

		a.tview.QueueUpdateDraw(func() {
			if err != nil {
				statusBar.SetText(fmt.Sprintf("%sPopulate failed: %v[-]", DefaultTheme.ErrorTag(""), err))
			} else {
				statusBar.SetText(DefaultTheme.SuccessTag("") + "Environment populated successfully!" + "[-]")
			}

			layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
					a.outputWriter.ClearView()
					a.returnFromPopulate()
					return nil
				}
				return event
			})
			a.UpdateFooter("[Populate Complete]", KeyDescriptions["populate-complete"])
		})
	}()
}

// returnFromPopulate cleans up and returns to the home screen.
func (a *App) returnFromPopulate() {
	a.pages.RemovePage("populate-confirm")
	a.pages.RemovePage("populate-progress")
	a.pages.SwitchToPage("home")
	a.refreshLists()

	if a.actionFromDetails {
		a.tview.SetFocus(a.details)
		a.UpdateFooter("[Environment Details]", KeyDescriptions["details"])
	} else {
		if a.docker.GetItemCount() > 0 {
			a.tview.SetFocus(a.docker)
		} else {
			a.tview.SetFocus(a.dockerEmpty)
		}
		a.UpdateFooter("[Docker Environments]", KeyDescriptions["docker"])
	}
	a.actionFromDetails = false
}
