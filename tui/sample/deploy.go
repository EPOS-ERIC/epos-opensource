package sample

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// DeployScreen represents the deploy screen with a form.
// Shows how screens can have interactive elements and handle form submissions.
type DeployScreen struct {
	form *tview.Form
}

// Name returns the screen's identifier.
func (d *DeployScreen) Name() string {
	return "deploy"
}

// Init sets up the screen's UI.
// Creates form with buttons that call app.SwitchTo for navigation.
// Uses type assertion to access *App for tview operations, as AppInterface hides tview details for testability.
func (d *DeployScreen) Init(app AppInterface) {
	d.form = tview.NewForm().
		AddInputField("Name", "", 20, nil, nil).
		AddButton("Deploy", func() {
			// Simulate deploy action
			app.SwitchTo("home")
		}).
		AddButton("Cancel", func() {
			app.SwitchTo("home")
		})
	d.form.SetBorder(true).SetTitle("Deploy Form")
	// Type assert to *App for tview-specific operations; mocks in tests won't match, so this is skipped in testing.
	if realApp, ok := app.(*App); ok {
		realApp.pages.AddPage("deploy", d.form, true, true)
	}
}

// HandleEvent processes key events.
// Demonstrates custom key handling (e.g., ESC) alongside tview's form events.
func (d *DeployScreen) HandleEvent(event *tcell.EventKey, app AppInterface) (bool, Event) {
	// Form handles its own events, but we can add custom keys
	if event.Key() == tcell.KeyEsc {
		return true, EventSwitchHome
	}
	log.Printf("tui: deploy screen unhandled key %v", event.Key())
	return false, ""
}

// UpdateFooter sets screen-specific footer text.
func (d *DeployScreen) UpdateFooter(app AppInterface) {
	app.UpdateFooter("Deploy Screen", []string{"esc: back", "enter: submit"})
}
