// Package sample demonstrates a screen-based TUI architecture.
package sample

import "github.com/gdamore/tcell/v2"

// Event represents actions that screens can request from the app.
// Using a type instead of raw strings improves maintainability and type safety.
type Event string

const (
	EventQuit         Event = "quit"
	EventSwitchHome   Event = "switch:home"
	EventSwitchDeploy Event = "switch:deploy"
)

// AppInterface defines the methods screens can call on the app.
// This interface enables dependency injection for testing, allowing screens
// to be tested with mock implementations instead of the full tview-based App.
type AppInterface interface {
	RegisterScreen(screen Screen)
	SwitchTo(screenName string)
	UpdateFooter(sectionText string, keys []string)
}

// Screen defines the interface for TUI screens.
// This enables polymorphism: App can hold screens as interfaces,
// delegating events and updates without knowing concrete types.
type Screen interface {
	Name() string
	Init(app AppInterface)
	HandleEvent(event *tcell.EventKey, app AppInterface) (handled bool, command Event)
	UpdateFooter(app AppInterface)
}
