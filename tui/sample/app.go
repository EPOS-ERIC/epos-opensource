package sample

import (
	"log"

	"github.com/epos-eu/epos-opensource/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App holds the main TUI application state.
// It centralizes global elements like pages and frame,
// allowing screens to update shared state (e.g., footer) via dependency injection.
type App struct {
	app           *tview.Application
	pages         *tview.Pages
	frame         *tview.Frame
	currentScreen Screen
	screens       map[string]Screen
	config        *config.Config
}

// NewApp creates a new App instance.
// Initializes tview components and sets up the frame for footer rendering.
func NewApp(cfg *config.Config) *App {
	app := tview.NewApplication()
	pages := tview.NewPages()
	frame := tview.NewFrame(pages).SetBorders(0, 0, 0, 0, 0, 0)

	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Borders.HorizontalFocus = tview.Borders.Horizontal
	tview.Borders.VerticalFocus = tview.Borders.Vertical
	tview.Borders.TopLeft = '╭'
	tview.Borders.TopRight = '╮'
	tview.Borders.BottomLeft = '╰'
	tview.Borders.BottomRight = '╯'
	tview.Borders.TopLeftFocus = '╭'
	tview.Borders.TopRightFocus = '╮'
	tview.Borders.BottomLeftFocus = '╰'
	tview.Borders.BottomRightFocus = '╯'

	return &App{
		app:     app,
		pages:   pages,
		frame:   frame,
		screens: make(map[string]Screen),
		config:  cfg,
	}
}

// RegisterScreen adds a screen to the app.
// Calls screen.Init to build UI, enabling lazy initialization without cycles.
func (a *App) RegisterScreen(screen Screen) {
	a.screens[screen.Name()] = screen
	screen.Init(a)
}

// SwitchTo switches to a different screen.
// Updates pages, current screen, and footer, handling navigation commands.
func (a *App) SwitchTo(screenName string) {
	if screen, exists := a.screens[screenName]; exists {
		a.currentScreen = screen
		a.pages.SwitchToPage(screenName)
		screen.UpdateFooter(a)
	}
}

// UpdateFooter updates the frame footer.
// Screens call this via injected *App to set context-specific help text.
func (a *App) UpdateFooter(sectionText string, keys []string) {
	a.frame.Clear()
	a.frame.AddText("[::b]"+sectionText, false, tview.AlignLeft, tcell.ColorWhite)
	keysText := ""
	for i, key := range keys {
		if i > 0 {
			keysText += ", "
		}
		keysText += key
	}
	a.frame.AddText("[::b]"+keysText, false, tview.AlignCenter, tcell.ColorWhite)
	a.frame.AddText("[::b]Sample TUI", false, tview.AlignRight, tcell.ColorWhite)
}

// Run starts the TUI application.
// Sets initial screen, captures global events (delegating to screens),
// and handles commands like quit or switch.
func (a *App) Run() error {
	// Set initial screen
	if len(a.screens) > 0 {
		initial := "home" // default
		if a.config != nil && a.config.TUI.DefaultScreen != "" {
			initial = a.config.TUI.DefaultScreen
		}
		a.SwitchTo(initial)
	}

	// Global input capture
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if a.currentScreen != nil {
			if handled, command := a.currentScreen.HandleEvent(event, a); handled {
				if command == EventQuit {
					log.Printf("tui: quitting")
					a.app.Stop()
					return nil
				}
				if len(string(command)) > 7 && string(command)[:7] == "switch:" {
					target := string(command)[7:]
					log.Printf("tui: switching to %s", target)
					a.SwitchTo(target)
					return nil
				}
				return nil
			}
		}
		// Global quit
		if event.Rune() == 'q' {
			a.app.Stop()
			return nil
		}
		return event
	})

	return a.app.SetRoot(a.frame, true).Run()
}
