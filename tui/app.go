package tui

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/epos-eu/epos-opensource/command"
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/config"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App holds the main TUI application state.
// All screens access shared state through this struct.
type App struct {
	tview  *tview.Application
	pages  *tview.Pages
	frame  *tview.Frame
	config config.Config

	// Home screen components
	docker             *tview.List
	dockerEmpty        *tview.TextView
	dockerFlex         *tview.Flex
	dockerFlexInner    *tview.Flex
	dockerEnvs         []string // Environment names for docker list (lookup by index)
	k8s                *tview.List
	k8sEmpty           *tview.TextView
	k8sFlex            *tview.Flex
	k8sEnvs            []string // Environment names for k8s list (lookup by index)
	createNewButton    *tview.Button
	buttonFlex         *tview.Flex
	details            *tview.Flex
	detailsGrid        *tview.Grid
	detailsButtons     []*tview.Button
	nameDirGrid        *tview.Grid
	nameDirButtons     []*tview.Button
	buttonsFlex        *tview.Flex
	deleteButton       *tview.Button
	cleanButton        *tview.Button
	updateButton       *tview.Button
	populateButton     *tview.Button
	detailsList        *tview.List
	detailsListEmpty   *tview.TextView
	detailsListFlex    *tview.Flex
	detailsEmpty       *tview.TextView
	currentEnv         tview.Primitive
	homeFlex           *tview.Flex
	detailsShown       bool
	currentDetailsName string
	currentDetailsType string
	currentDetailsRows []DetailRow
	previousFocus      tview.Primitive

	// Background tasks
	refreshTicker *time.Ticker
	refreshMutex  sync.Mutex

	// Footer state tracking for FlashMessage
	currentFooterSection string
	currentFooterKeys    []string
	footerMutex          sync.Mutex

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
	cfg, err := config.LoadConfig()
	if err != nil {
		// Log error and use defaults
		log.Printf("config error: %v, using defaults", err)
		cfg = config.DefaultConfig()
	}
	a.config = cfg

	a.tview = tview.NewApplication()
	a.tview.EnableMouse(true)
	a.tview.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			a.Quit()
			return nil
		case '?':
			a.showHelp()
			return nil
		}
		return event
	})
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

	// Set initial focus
	if a.docker.GetItemCount() > 0 {
		a.tview.SetFocus(a.docker)
		a.UpdateFooter("[Docker Environments]", KeyDescriptions["docker"])
	} else {
		a.tview.SetFocus(a.createNewButton)
		a.UpdateFooter("[Docker Environments]", KeyDescriptions["docker"])
	}

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
	a.footerMutex.Lock()
	a.currentFooterSection = section
	a.currentFooterKeys = keys
	a.footerMutex.Unlock()

	a.drawFooter(section, keys)
}

// drawFooter actually renders the footer components.
func (a *App) drawFooter(section string, keys []string) {
	section = tview.Escape(section)
	keyString := tview.Escape(strings.Join(keys, ", "))
	a.frame.Clear()
	a.frame.AddText("[::b]"+section, false, tview.AlignLeft, DefaultTheme.OnPrimary)
	a.frame.AddText("[::b]"+keyString, false, tview.AlignCenter, DefaultTheme.OnPrimary)

	version := fmt.Sprintf("epos-opensource [%s]", common.GetVersion())
	gradient := CreateGradient(version, DefaultTheme.Secondary, DefaultTheme.OnSecondary)
	a.frame.AddText(gradient, false, tview.AlignRight, DefaultTheme.OnBackground)
}

// FlashMessage shows a temporary message in the footer for the specified duration.
func (a *App) FlashMessage(message string, duration time.Duration) {
	a.tview.QueueUpdateDraw(func() {
		a.footerMutex.Lock()
		section := a.currentFooterSection
		a.footerMutex.Unlock()

		a.drawFooter(section, []string{message})
	})

	go func() {
		time.Sleep(duration)
		a.tview.QueueUpdateDraw(func() {
			a.footerMutex.Lock()
			section := a.currentFooterSection
			keys := a.currentFooterKeys
			a.footerMutex.Unlock()

			a.drawFooter(section, keys)
		})
	}()
}

// ShowError displays an error modal with a message.
// Press OK or ESC to dismiss.
func (a *App) ShowError(message string) {
	a.previousFocus = a.tview.GetFocus()
	modal := tview.NewModal().
		SetText(DefaultTheme.DestructiveTag("b") + message + "[-]").
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("error")
			if a.previousFocus != nil {
				a.tview.SetFocus(a.previousFocus)
			}
		})

	modal.SetBackgroundColor(DefaultTheme.Background)
	modal.Box.SetBackgroundColor(DefaultTheme.Background)
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
