package sample

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// HomeScreen represents the home screen with a simple text view.
// It demonstrates a basic screen implementation, handling navigation to other screens.
type HomeScreen struct {
	textView *tview.TextView
}

// Name returns the screen's identifier.
// Used for registration and switching.
func (h *HomeScreen) Name() string {
	return "home"
}

// Init sets up the screen's UI.
// Builds primitives and adds to app.pages during registration.
// Uses type assertion to access *App for tview operations, as AppInterface hides tview details for testability.
func (h *HomeScreen) Init(app AppInterface) {
	h.textView = tview.NewTextView().
		SetText("Welcome to the Sample TUI Home Screen!\n\nPress 'd' to go to Deploy, 'q' to quit.").
		SetTextAlign(tview.AlignCenter)
	// Type assert to *App for tview-specific operations; mocks in tests won't match, so this is skipped in testing.
	if realApp, ok := app.(*App); ok {
		realApp.pages.AddPage("home", h.textView, true, true)
	}
}

// HandleEvent processes key events for this screen.
// Returns command strings for navigation, enabling App to handle switches.
func (h *HomeScreen) HandleEvent(event *tcell.EventKey, app AppInterface) (bool, Event) {
	if event.Rune() == 'd' {
		return true, EventSwitchDeploy
	}
	log.Printf("tui: home screen unhandled key %v", event.Rune())
	return false, ""
}

// UpdateFooter sets screen-specific footer text.
// Called by App on switch, using injected AppInterface to update global footer.
func (h *HomeScreen) UpdateFooter(app AppInterface) {
	app.UpdateFooter("Home Screen", []string{"d: deploy", "q: quit"})
}
