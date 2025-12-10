package tui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/epos-eu/epos-opensource/command"
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App holds the main TUI application state.
// All screens access shared state through this struct.
type App struct {
	tview *tview.Application
	pages *tview.Pages
	frame *tview.Frame

	// Home screen components
	docker      *tview.List
	dockerEmpty *tview.TextView
	dockerFlex  *tview.Flex
	dockerEnvs  []string // Environment names for docker list (lookup by index)
	k8s         *tview.List
	k8sEmpty    *tview.TextView
	k8sFlex     *tview.Flex
	k8sEnvs     []string // Environment names for k8s list (lookup by index)
	details     *tview.Box
	currentEnv  tview.Primitive

	// Background tasks
	refreshTicker *time.Ticker
	refreshMutex  sync.Mutex

	// Output capture for deploy/command screens
	outputWriter *OutputWriter
}

// Run initializes and starts the TUI application.
// This is the main entry point for the TUI.
func Run() error {
	app := &App{}
	app.init()
	return app.run()
}

// init sets up the application state and UI components.
func (a *App) init() {
	a.tview = tview.NewApplication()
	a.pages = tview.NewPages()
	a.outputWriter = &OutputWriter{}

	// Apply global styles
	InitStyles()

	a.pages.SetBackgroundColor(tcell.ColorDefault)

	// Redirect CLI output to TUI capture
	display.Stdout = a.outputWriter
	display.Stderr = a.outputWriter
	command.Stdout = a.outputWriter
	command.Stderr = a.outputWriter

	// Build home screen
	home := a.createHome()
	a.pages.AddPage("home", home, true, true)

	// Wrap in frame for footer
	a.frame = tview.NewFrame(a.pages).SetBorders(0, 0, 0, 0, 0, 0)
	a.frame.SetBackgroundColor(DefaultTheme.Primary)

	// Set initial footer
	a.UpdateFooter("[Docker Environments]", KeyDescriptions["docker"])

	// Start background refresh
	a.startRefreshTicker()
}

// run starts the tview event loop.
func (a *App) run() error {
	return a.tview.SetRoot(a.frame, true).Run()
}

// startRefreshTicker starts background list refresh every second.
func (a *App) startRefreshTicker() {
	a.refreshTicker = time.NewTicker(1 * time.Second)
	go func() {
		for range a.refreshTicker.C {
			a.tview.QueueUpdateDraw(func() {
				a.refreshLists()
			})
		}
	}()
}

// UpdateFooter updates the frame footer with section text and available keys.
// Called by screens when they become active.
func (a *App) UpdateFooter(section string, keys []string) {
	section = tview.Escape(section)
	keyString := tview.Escape(strings.Join(keys, ", "))
	a.frame.Clear()
	a.frame.AddText("[::b]"+section, false, tview.AlignLeft, DefaultTheme.OnPrimary)
	a.frame.AddText("[::b]"+keyString, false, tview.AlignCenter, DefaultTheme.OnPrimary)

	version := fmt.Sprintf("EPOS Open source [%s]", common.GetVersion())
	gradient := CreateGradient(version, DefaultTheme.Secondary, DefaultTheme.OnSecondary)
	a.frame.AddText(gradient, false, tview.AlignRight, DefaultTheme.OnBackground)
}

// ShowError displays an error modal with a message.
// Press OK or ESC to dismiss.
func (a *App) ShowError(message string) {
	modal := tview.NewModal().
		SetText(DefaultTheme.DestructiveTag("b") + message + "[-]").
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("error")
		})

	modal.SetBackgroundColor(DefaultTheme.Background)
	modal.Box.SetBackgroundColor(DefaultTheme.Surface)
	modal.SetBorderColor(DefaultTheme.Destructive)
	modal.SetTitle(" [::b]Error ")
	modal.SetTitleColor(DefaultTheme.Destructive)
	modal.SetButtonActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))

	a.pages.AddPage("error", modal, true, true)
	a.tview.SetFocus(modal)
}

// CenterPrimitive wraps a primitive in a flex layout that centers it.
// Use width/height as proportions (1 = minimal, higher = more space).
func CenterPrimitive(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, 0, max(1, height), true).
			AddItem(nil, 0, 1, false), 0, max(1, width), true).
		AddItem(nil, 0, 1, false)
}

// Quit stops the application.
func (a *App) Quit() {
	if a.refreshTicker != nil {
		a.refreshTicker.Stop()
	}
	a.tview.Stop()
}
