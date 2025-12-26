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

const homePageName = "home"

// App manages the global TUI state, coordinator for components and navigation.
type App struct {
	tview  *tview.Application
	pages  *tview.Pages
	frame  *tview.Frame
	config config.Config

	envList      *EnvList
	detailsPanel *DetailsPanel
	homeFlex     *tview.Flex
	focusStack   []tview.Primitive

	refreshTicker *time.Ticker
	refreshMutex  sync.Mutex

	currentFooterSection string
	currentFooterKeys    []string
	currentContext       ScreenKey
	currentPage          string
	footerMutex          sync.Mutex

	outputWriter *OutputWriter
}

// ResetOptions configures the return to home screen.
type ResetOptions struct {
	PageNames     []string // Pages to remove
	ClearDetails  bool     // Clear details panel (e.g., after delete)
	RefreshFiles  bool     // Refresh files list (e.g., after populate)
	RestoreFocus  bool     // Restore focus from stack
	ForceEnvFocus bool     // Explicitly focus environment list
}

// ConfirmationOptions defines settings for the standardized confirmation modal.
type ConfirmationOptions struct {
	PageName           string
	Title              string
	Message            string
	ConfirmLabel       string
	CancelLabel        string
	OnConfirm          func()
	OnCancel           func()
	Destructive        bool // Use Red border
	ConfirmDestructive bool // Use Red confirm button
	Secondary          bool // Use Yellow border (ignored if Destructive)
	InputCapture       func(leftBtn, rightBtn *tview.Button) func(*tcell.EventKey) *tcell.EventKey
}

// TaskOptions defines settings for background operations with progress UI.
type TaskOptions struct {
	Operation    string
	EnvName      string
	IsDocker     bool
	Task         func() (string, error) // Returns success message or error
	OnSuccess    func()
	ClearDetails bool
}

// Run starts the TUI application.
func Run() error {
	app := &App{}
	app.init()
	return app.run()
}

func (a *App) init() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("config error: %v, using defaults", err)
		cfg = config.DefaultConfig()
	}
	a.config = cfg

	a.tview = tview.NewApplication()
	a.tview.EnableMouse(true)
	a.tview.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Don't handle global keys when focused on input components
		if current := a.tview.GetFocus(); current != nil {
			switch current.(type) {
			case *tview.InputField, *tview.DropDown, *tview.Form:
				return event
			}
		}

		switch event.Rune() {
		case 'q':
			a.Quit()
			return nil
		case '?':
			if (a.currentPage == homePageName || a.currentPage == "file-picker") && a.currentContext != "help" {
				a.showHelp()
			}
			return nil
		}
		return event
	})
	a.pages = tview.NewPages()
	a.outputWriter = &OutputWriter{}

	InitStyles()

	a.pages.SetBackgroundColor(tcell.ColorDefault)

	// Redirect command output to TUI
	display.Stdout = a.outputWriter
	display.Stderr = a.outputWriter
	command.Stdout = a.outputWriter
	command.Stderr = a.outputWriter

	a.envList = NewEnvList(a)
	a.detailsPanel = NewDetailsPanel(a)

	home := a.createHome()
	a.pages.AddPage(homePageName, home, true, true)
	a.currentPage = homePageName

	a.frame = tview.NewFrame(a.pages).SetBorders(0, 0, 0, 0, 0, 0)
	a.frame.SetBackgroundColor(DefaultTheme.Primary)

	a.startRefreshTicker()
}

// run starts the tview event loop.
func (a *App) run() error {
	a.tview.SetRoot(a.frame, true)
	a.envList.SetInitialFocus()
	return a.tview.Run()
}

// startRefreshTicker starts background list refresh every second.
func (a *App) startRefreshTicker() {
	a.refreshTicker = time.NewTicker(1 * time.Second)
	go func() {
		for range a.refreshTicker.C {
			a.tview.QueueUpdateDraw(func() {
				a.envList.Refresh()
			})
		}
	}()
}

// UpdateFooter updates the footer section text and shortcut keys.
func (a *App) UpdateFooter(section string, contextKey ScreenKey) {
	keys := getFooterHints(contextKey)
	a.footerMutex.Lock()
	a.currentFooterSection = section
	a.currentFooterKeys = keys
	a.setContext(contextKey)
	a.footerMutex.Unlock()

	a.drawFooter(section, keys)
}

// UpdateFooterCustom updates the footer with custom keys (not from context).
func (a *App) UpdateFooterCustom(section string, keys []string) {
	a.footerMutex.Lock()
	a.currentFooterSection = section
	a.currentFooterKeys = keys
	a.currentContext = ""
	a.footerMutex.Unlock()

	a.drawFooter(section, keys)
}

// setContext sets the current context with validation.
func (a *App) setContext(contextKey ScreenKey) {
	if _, ok := KeyHints[contextKey]; !ok {
		log.Printf("warning: invalid context key '%s', defaulting to 'general'", contextKey)
		contextKey = "general"
	}
	a.currentContext = contextKey
}

// drawFooter actually renders the footer components.
func (a *App) drawFooter(section string, keys []string) {
	section = tview.Escape(section)
	keyString := tview.Escape(strings.Join(keys, ", "))
	a.frame.Clear()
	a.frame.AddText("[::b]"+section, false, tview.AlignLeft, DefaultTheme.OnPrimary)
	a.frame.AddText("[::b]"+keyString, false, tview.AlignCenter, DefaultTheme.OnPrimary)

	version := fmt.Sprintf("epos-opensource [%s]", common.GetVersion())
	a.frame.AddText("[::b]"+tview.Escape(version), false, tview.AlignRight, DefaultTheme.OnPrimary)
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

// PushFocus saves the current focus to the stack.
func (a *App) PushFocus() {
	current := a.tview.GetFocus()
	if current != nil {
		a.focusStack = append(a.focusStack, current)
	}
}

// PopFocus restores the last saved focus from the stack.
// Returns the restored primitive or nil if stack was empty.
func (a *App) PopFocus() tview.Primitive {
	if len(a.focusStack) == 0 {
		return nil
	}
	lastIdx := len(a.focusStack) - 1
	p := a.focusStack[lastIdx]
	a.focusStack = a.focusStack[:lastIdx]
	a.tview.SetFocus(p)
	return p
}

// ShowError displays an error modal with a message.
// Press OK or ESC to dismiss.
func (a *App) ShowError(message string) {
	a.PushFocus()
	modal := tview.NewModal().
		SetText(DefaultTheme.DestructiveTag("b") + message + "[-]").
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("error")
			a.PopFocus()
		})

	modal.SetBackgroundColor(DefaultTheme.Background)
	modal.Box.SetBackgroundColor(DefaultTheme.Background)
	modal.SetBorderColor(DefaultTheme.Destructive)
	modal.SetTitle(" [::b]Error ")
	modal.SetTitleColor(DefaultTheme.Destructive)
	modal.SetButtonActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))

	a.pages.AddPage("error", modal, true, true)
	a.currentPage = "error"
	a.tview.SetFocus(modal)
}

// ResetToHome cleans up pages and returns to the home screen.
// Handles page removal, refreshing lists, and restoring focus.
func (a *App) ResetToHome(opts ResetOptions) {
	for _, page := range opts.PageNames {
		a.pages.RemovePage(page)
	}

	a.pages.SwitchToPage(homePageName)
	a.currentPage = homePageName
	a.envList.Refresh()

	if opts.RefreshFiles {
		a.detailsPanel.RefreshFiles()
	}

	if opts.ClearDetails {
		a.detailsPanel.Clear()
		a.envList.FocusActiveList()
	} else if opts.ForceEnvFocus {
		a.envList.FocusActiveList()
	} else if opts.RestoreFocus && len(a.focusStack) > 0 {
		a.PopFocus()
		if a.detailsPanel.IsShown() {
			key := getDetailsKey(a.detailsPanel.GetCurrentDetailsType())
			a.UpdateFooter(GetFooterText(key), key)
		}
	} else if a.detailsPanel.IsShown() {
		// If details are shown and we didn't force env or restore prev, focus details
		key := getDetailsKey(a.detailsPanel.GetCurrentDetailsType())
		a.UpdateFooter("[Environment Details]", key)
		a.tview.SetFocus(a.detailsPanel.GetFlex())
	} else {
		// Default fallback
		a.envList.FocusActiveList()
	}
}

// ShowConfirmation displays a standardized confirmation modal.
func (a *App) ShowConfirmation(opts ConfirmationOptions) {
	a.PushFocus()

	// Create text view for message
	textView := tview.NewTextView().
		SetText(opts.Message).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	textView.SetBorderPadding(1, 0, 1, 1)

	// Create styled buttons
	confirmBtn := tview.NewButton(opts.ConfirmLabel).SetSelectedFunc(func() {
		a.pages.RemovePage(opts.PageName)
		opts.OnConfirm()
	})

	if opts.ConfirmDestructive {
		confirmBtn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Destructive).Foreground(DefaultTheme.OnDestructive))
		confirmBtn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.OnSecondary))
	} else {
		confirmBtn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
		confirmBtn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.OnSecondary))
	}

	cancelBtn := tview.NewButton(opts.CancelLabel).SetSelectedFunc(func() {
		opts.OnCancel()
	})
	cancelBtn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
	cancelBtn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.OnSecondary))

	// Navigation
	buttonInputCapture := func(leftBtn, rightBtn *tview.Button) func(*tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyLeft, tcell.KeyBacktab:
				if a.tview.GetFocus() == leftBtn {
					a.tview.SetFocus(rightBtn)
				} else {
					a.tview.SetFocus(leftBtn)
				}
				return nil
			case tcell.KeyRight, tcell.KeyTab:
				if a.tview.GetFocus() == rightBtn {
					a.tview.SetFocus(leftBtn)
				} else {
					a.tview.SetFocus(rightBtn)
				}
				return nil
			case tcell.KeyEsc:
				opts.OnCancel()
				return nil
			}
			return event
		}
	}
	confirmBtn.SetInputCapture(buttonInputCapture(confirmBtn, cancelBtn))
	cancelBtn.SetInputCapture(buttonInputCapture(confirmBtn, cancelBtn))

	buttonContainer := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(confirmBtn, len(opts.ConfirmLabel)+4, 0, true).
		AddItem(tview.NewBox(), 2, 0, false).
		AddItem(cancelBtn, len(opts.CancelLabel)+4, 0, true).
		AddItem(tview.NewBox(), 0, 1, false)
	buttonContainer.SetBackgroundColor(tcell.ColorDefault)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(buttonContainer, 1, 0, true)

	borderColor := DefaultTheme.Secondary
	if opts.Destructive {
		borderColor = DefaultTheme.Destructive
	} else if opts.Secondary {
		borderColor = DefaultTheme.Secondary
	}

	layout.SetBorder(true).
		SetTitle(opts.Title).
		SetTitleColor(DefaultTheme.Secondary).
		SetBorderColor(borderColor).
		SetBackgroundColor(DefaultTheme.Background)

	// Center layout
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

	a.pages.AddPage(opts.PageName, outerLayout, true, true)
	a.currentPage = opts.PageName
	a.tview.SetFocus(confirmBtn)
}

// RunBackgroundTask runs an operation with a standard progress UI.
// It starts a new OperationProgress screen and executes the provided task in a goroutine.
// Handles updating the UI upon completion (success or error).
func (a *App) RunBackgroundTask(opts TaskOptions) {
	progress := NewOperationProgress(a, opts.Operation, opts.EnvName)
	progress.Start()

	go func() {
		msg, err := opts.Task()
		if err != nil {
			progress.Complete(false, err.Error())
		} else {
			if msg == "" {
				msg = fmt.Sprintf("%s completed successfully!", opts.Operation)
			}
			if opts.OnSuccess != nil {
				opts.OnSuccess()
			}
			progress.Complete(true, msg)
		}
	}()
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

// CenterPrimitiveFixed wraps a primitive in a flex layout that centers it with fixed dimensions.
func CenterPrimitiveFixed(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 0, true).                // height=fixed, proportion=0
			AddItem(nil, 0, 1, false), width, 0, true). // width=fixed, proportion=0
		AddItem(nil, 0, 1, false)
}

// Quit stops the application.
func (a *App) Quit() {
	if a.refreshTicker != nil {
		a.refreshTicker.Stop()
	}
	a.tview.Stop()
}
