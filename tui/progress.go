package tui

import (
	"fmt"
	"regexp"
	"time"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	StateRunning = "running"
	StateSuccess = "success"
	StateError   = "error"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

// stripANSI removes ANSI escape codes from a string.
func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// OperationProgress manages the UI for operation progress screens.
type OperationProgress struct {
	app       *App
	operation string // "Deploy", "Populate", etc.
	envName   string
	startTime time.Time
	state     string // StateRunning, StateSuccess, StateError
	errorMsg  string

	// UI Components
	layout   *tview.Grid     // Main grid layout
	header   *tview.Flex     // Top flex for title and clock
	title    *tview.TextView // Operation: Env name
	clock    *tview.TextView // Elapsed time
	logsView *tview.TextView // Live output
	overlay  tview.Primitive // Completion modal

	// Navigation state
	wasInDetails     bool
	savedDetailsName string
	savedDetailsType string

	// Internal
	ticker *time.Ticker
	done   chan bool
}

// NewOperationProgress creates a new OperationProgress instance.
func NewOperationProgress(app *App, operation, envName string) *OperationProgress {
	op := &OperationProgress{
		app:       app,
		operation: operation,
		envName:   envName,
		startTime: time.Now(),
		state:     StateRunning,
		done:      make(chan bool),
	}
	// Save current navigation state
	op.wasInDetails = app.detailsShown
	op.savedDetailsName = app.currentDetailsName
	op.savedDetailsType = app.currentDetailsType
	op.rebuildUI()
	return op
}

// rebuildUI builds the progress UI components.
func (op *OperationProgress) rebuildUI() {
	// Create title and clock for header flex
	titleText := fmt.Sprintf("  [%s::b]%s: %s[-]", DefaultTheme.Hex(DefaultTheme.OnSurface), op.operation, op.envName)
	op.title = tview.NewTextView().
		SetDynamicColors(true).
		SetText(titleText)
	op.title.SetBackgroundColor(DefaultTheme.HeaderBackground)

	op.clock = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight)
	op.clock.SetBackgroundColor(DefaultTheme.HeaderBackground)

	op.header = tview.NewFlex().
		AddItem(op.title, 0, 1, false).
		AddItem(op.clock, 10, 0, false)
	op.header.SetBackgroundColor(DefaultTheme.HeaderBackground)

	// Create logs view
	op.logsView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() { op.app.tview.Draw() })

	// Create grid layout
	op.layout = tview.NewGrid().
		SetRows(1, 0).
		SetColumns(0).
		SetBorders(true).
		SetBordersColor(DefaultTheme.Secondary)

	op.layout.AddItem(op.header, 0, 0, 1, 1, 0, 0, false).
		AddItem(op.logsView, 1, 0, 1, 1, 0, 0, true)

	op.layout.SetTitle(fmt.Sprintf(" [::b]%s Progress ", op.operation)).
		SetTitleColor(DefaultTheme.Secondary)
}

// getProgressStatus returns status message for the footer.
func (op *OperationProgress) getProgressStatus() string {
	switch op.state {
	case StateRunning:
		return fmt.Sprintf("%s Operation in progress...", op.operation)
	case StateSuccess:
		return fmt.Sprintf("✓ %s Complete", op.operation)
	case StateError:
		return fmt.Sprintf("✕ %s Failed", op.operation)
	default:
		return ""
	}
}

// Start begins the progress display and timer updates.
func (op *OperationProgress) Start() {
	// Update global footer
	footerTitle := fmt.Sprintf("[%s Progress]", op.operation)
	op.app.UpdateFooter(footerTitle, []string{op.getProgressStatus()})

	// Connect output writer
	op.app.outputWriter.ClearBuffer()
	op.app.outputWriter.SetView(op.app.tview, op.logsView)

	// Start timer for header updates
	op.ticker = time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-op.ticker.C:
				op.updateHeader()
			case <-op.done:
				op.ticker.Stop()
				return
			}
		}
	}()

	// Set input capture
	op.layout.SetInputCapture(op.handleInput)

	// Add to pages and switch
	pageName := fmt.Sprintf("%s-progress", op.operation)
	op.app.pages.AddAndSwitchToPage(pageName, op.layout, true)
}

// updateHeader updates the header with current elapsed time.
func (op *OperationProgress) updateHeader() {
	elapsed := time.Since(op.startTime)
	elapsedStr := op.formatElapsedTime(elapsed)
	clockText := fmt.Sprintf("[%s]%s[-]  ",
		DefaultTheme.Hex(DefaultTheme.OnSurface),
		elapsedStr)

	op.app.tview.QueueUpdateDraw(func() {
		op.clock.SetText(clockText)
	})
}

// formatElapsedTime formats duration as MM:SS.
func (op *OperationProgress) formatElapsedTime(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// Complete marks the operation as complete and shows overlay.
func (op *OperationProgress) Complete(success bool, errorMsg string) {
	if success {
		op.state = StateSuccess
	} else {
		op.state = StateError
		op.errorMsg = errorMsg
	}

	// Stop ticker
	close(op.done)

	// Update global footer and border color
	op.app.tview.QueueUpdateDraw(func() {
		borderColor := DefaultTheme.Success
		if !success {
			borderColor = DefaultTheme.Error
		}
		op.layout.SetBordersColor(borderColor)

		footerTitle := fmt.Sprintf("[%s Progress]", op.operation)
		op.app.UpdateFooter(footerTitle, []string{op.getProgressStatus()})
	})

	// Show completion overlay
	op.showCompletionOverlay()
}

// showCompletionOverlay displays the completion modal.
func (op *OperationProgress) showCompletionOverlay() {
	var title, message string

	elapsed := time.Since(op.startTime)
	timeStr := op.formatElapsedTime(elapsed)

	if op.state == StateSuccess {
		title = fmt.Sprintf("✓ %s Complete", op.operation)
		message = fmt.Sprintf("%s operation completed successfully.\n\n\nTime: [::b]%s[-]", op.operation, timeStr)
	} else {
		title = fmt.Sprintf("✕ %s Failed", op.operation)
		errMsg := op.errorMsg
		if len(errMsg) > 200 {
			errMsg = "Operation failed. Check the logs below for full details."
		}
		message = fmt.Sprintf("%s\n\nTime: [::b]%s[-]", errMsg, timeStr)
	}

	doneFunc := func(buttonLabel string) {
		switch buttonLabel {
		case "Close & Return":
			op.returnToHome()
		case "Copy Logs":
			op.copyLogs()
		case "View Logs":
			op.app.pages.HidePage("completion-overlay")
			op.app.tview.SetFocus(op.layout)
			footerTitle := fmt.Sprintf("[%s Progress]", op.operation)
			op.app.UpdateFooter(footerTitle, []string{"Esc: back", "c: copy logs"})
		}
	}

	// Create message view
	msgView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("\n" + message)
	msgView.SetBackgroundColor(DefaultTheme.Background).SetBorderPadding(1, 0, 1, 1)

	// Create buttons
	closeBtn := tview.NewButton("Close & Return").SetSelectedFunc(func() { doneFunc("Close & Return") })
	viewBtn := tview.NewButton("View Logs").SetSelectedFunc(func() { doneFunc("View Logs") })

	// Style buttons
	styleBtn := func(btn *tview.Button) {
		btn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
		btn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))
	}
	styleBtn(closeBtn)
	styleBtn(viewBtn)

	// button navigation
	buttonInputCapture := func(prev, next *tview.Button) func(*tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyLeft, tcell.KeyBacktab:
				op.app.tview.SetFocus(prev)
				return nil
			case tcell.KeyRight, tcell.KeyTab:
				op.app.tview.SetFocus(next)
				return nil
			case tcell.KeyEsc:
				// Hide overlay and show logs instead of returning home
				op.app.pages.HidePage("completion-overlay")
				op.app.tview.SetFocus(op.layout)
				footerTitle := fmt.Sprintf("[%s Progress]", op.operation)
				op.app.UpdateFooter(footerTitle, []string{"Esc: back", "c: copy logs"})
				return nil
			}
			return event
		}
	}
	closeBtn.SetInputCapture(buttonInputCapture(viewBtn, closeBtn))
	viewBtn.SetInputCapture(buttonInputCapture(viewBtn, closeBtn))

	buttonContainer := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(closeBtn, 18, 1, true).
		AddItem(tview.NewBox(), 2, 0, false).
		AddItem(viewBtn, 13, 1, false).
		AddItem(tview.NewBox(), 0, 1, false)

	// Main modal layout
	modalLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(msgView, 0, 1, false).
		AddItem(buttonContainer, 1, 0, true)

	borderColor := DefaultTheme.Success
	if op.state == StateError {
		borderColor = DefaultTheme.Error
	}

	modalLayout.SetBorder(true).
		SetTitle(fmt.Sprintf(" [::b]%s ", title)).
		SetTitleColor(borderColor).
		SetBorderColor(borderColor).
		SetBackgroundColor(DefaultTheme.Background)
	// modalLayout.SetBorderPadding(1, 1, 2, 2)

	// Overlay handles Esc to close
	modalLayout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			op.app.pages.HidePage("completion-overlay")
			op.app.tview.SetFocus(op.layout)
			footerTitle := fmt.Sprintf("[%s Progress]", op.operation)
			op.app.UpdateFooter(footerTitle, []string{"Esc: back", "c: copy logs"})
			return nil
		}
		return event
	})

	op.overlay = modalLayout
	op.app.pages.AddPage("completion-overlay", CenterPrimitive(modalLayout, 0, 1), true, true)
	op.app.tview.SetFocus(closeBtn)
}

// returnToHome cleans up and returns to home screen.
func (op *OperationProgress) returnToHome() {
	pageName := fmt.Sprintf("%s-progress", op.operation)
	op.app.pages.RemovePage(pageName)
	op.app.pages.RemovePage("completion-overlay")
	op.app.pages.SwitchToPage("home")
	op.app.refreshLists()

	// Special handling for delete operation
	if op.operation == "Delete" && op.state == StateSuccess {
		op.app.clearDetailsPanel()
	}

	if op.wasInDetails && op.operation != "Delete" {
		// Return to details view
		op.app.showDetails(op.savedDetailsName, op.savedDetailsType)
	} else {
		// Return to envlist
		if op.app.docker.GetItemCount() > 0 {
			op.app.tview.SetFocus(op.app.docker)
		} else {
			op.app.tview.SetFocus(op.app.createNewButton)
		}
		op.app.UpdateFooter("[Docker Environments]", KeyDescriptions["docker"])
	}
}

// handleInput handles key events for the progress screen.
func (op *OperationProgress) handleInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEsc {
		if op.state == StateRunning {
			return nil
		}
		op.app.outputWriter.ClearView()
		op.returnToHome()
		return nil
	}
	if op.state != StateRunning && (event.Rune() == 'c' || event.Rune() == 'C') {
		op.copyLogs()
		return nil
	}
	return event
}

// copyLogs copies the clean log output to the clipboard and updates the footer.
func (op *OperationProgress) copyLogs() {
	go func() {
		logs := op.app.outputWriter.GetBuffer()
		cleanLogs := stripANSI(logs)
		if err := common.CopyToClipboard(cleanLogs); err != nil {
			op.app.tview.QueueUpdateDraw(func() {
				op.app.ShowError(fmt.Sprintf("Failed to copy logs: %v", err))
			})
		} else {
			// Show success notification and restore footer after 2s
			op.app.FlashMessage("Logs copied to clipboard!", 2*time.Second)
		}
	}()
}
