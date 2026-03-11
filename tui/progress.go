package tui

import (
	"fmt"
	"regexp"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	StateRunning = "running"
	StateSuccess = "success"
	StateError   = "error"
)

const (
	opDelete   = "Delete"
	opDeploy   = "Deploy"
	opPopulate = "Populate"
	opClean    = "Clean"
	opUpdate   = "Update"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

// stripANSI removes ANSI escape codes from a string.
func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// OperationProgress manages the UI for operation progress screens.
type OperationProgress struct {
	app       *App
	operation string
	envName   string
	startTime time.Time
	state     string
	errorMsg  string

	layout   *tview.Grid
	header   *tview.Flex
	title    *tview.TextView
	clock    *tview.TextView
	logsView *tview.TextView
	overlay  tview.Primitive

	wasInDetails        bool
	savedDetailsName    string
	savedDetailsType    string
	savedDetailsContext string

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
	op.wasInDetails = app.detailsPanel.IsShown()
	op.savedDetailsName = app.detailsPanel.GetCurrentDetailsName()
	op.savedDetailsType = app.detailsPanel.GetCurrentDetailsType()
	op.savedDetailsContext = app.detailsPanel.GetCurrentDetailsContext()
	op.rebuildUI()
	return op
}

// rebuildUI constructs the component layout.
func (op *OperationProgress) rebuildUI() {
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

	op.logsView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() { op.app.tview.Draw() })

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

// Start begins the progress display and ticker.
func (op *OperationProgress) Start() {
	footerTitle := fmt.Sprintf("[%s Progress]", op.operation)
	op.app.UpdateFooterCustom(footerTitle, []string{op.getProgressStatus()})

	op.app.outputWriter.ClearBuffer()
	op.app.outputWriter.SetView(op.app.tview, op.logsView)

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

	op.layout.SetInputCapture(op.handleInput)

	pageName := fmt.Sprintf("%s-progress", op.operation)
	op.app.pages.AddAndSwitchToPage(pageName, op.layout, true)
	op.app.currentPage = pageName
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

// Complete marks the operation as finished and shows the completion overlay.
func (op *OperationProgress) Complete(success bool, errorMsg string) {
	if success {
		op.state = StateSuccess
	} else {
		op.state = StateError
		op.errorMsg = errorMsg
	}

	close(op.done)

	op.app.tview.QueueUpdateDraw(func() {
		borderColor := DefaultTheme.Success
		if !success {
			borderColor = DefaultTheme.Error
		}
		op.layout.SetBordersColor(borderColor)

		footerTitle := fmt.Sprintf("[%s Progress]", op.operation)
		op.app.UpdateFooterCustom(footerTitle, []string{op.getProgressStatus()})
	})

	op.showCompletionOverlay()
}

// showCompletionOverlay displays the completion modal with results and actions.
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
			op.app.UpdateFooterCustom(footerTitle, []string{"Esc: back", "c: copy logs"})
		}
	}

	msgView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("\n" + message)
	msgView.SetBackgroundColor(DefaultTheme.Background).SetBorderPadding(1, 0, 1, 1)

	closeBtn := tview.NewButton("Close & Return").SetSelectedFunc(func() { doneFunc("Close & Return") })
	viewBtn := tview.NewButton("View Logs").SetSelectedFunc(func() { doneFunc("View Logs") })

	ApplyButtonStyle(closeBtn)
	ApplyButtonStyle(viewBtn)

	buttonInputCapture := func(other *tview.Button) func(*tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyLeft, tcell.KeyRight, tcell.KeyTab, tcell.KeyBacktab:
				op.app.tview.SetFocus(other)
				return nil
			case tcell.KeyEsc:
				// Hide overlay and show logs
				op.app.pages.HidePage("completion-overlay")
				op.app.tview.SetFocus(op.layout)
				footerTitle := fmt.Sprintf("[%s Progress]", op.operation)
				op.app.UpdateFooterCustom(footerTitle, []string{"Esc: back", "c: copy logs"})
				return nil
			}
			return event
		}
	}
	closeBtn.SetInputCapture(buttonInputCapture(viewBtn))
	viewBtn.SetInputCapture(buttonInputCapture(closeBtn))

	buttonContainer := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(closeBtn, 18, 1, true).
		AddItem(tview.NewBox(), 2, 0, false).
		AddItem(viewBtn, 13, 1, false).
		AddItem(tview.NewBox(), 0, 1, false)

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

	modalLayout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			op.app.pages.HidePage("completion-overlay")
			op.app.tview.SetFocus(op.layout)
			footerTitle := fmt.Sprintf("[%s Progress]", op.operation)
			op.app.UpdateFooterCustom(footerTitle, []string{"Esc: back", "c: copy logs"})
			return nil
		}
		return event
	})

	op.overlay = modalLayout
	op.app.pages.AddPage("completion-overlay", CenterPrimitiveFixed(modalLayout, 50, 10), true, true)
	op.app.currentPage = "completion-overlay"
	op.app.tview.SetFocus(closeBtn)
}

// returnToHome cleans up and returns to home screen.
func (op *OperationProgress) returnToHome() {
	pageName := fmt.Sprintf("%s-progress", op.operation)
	op.app.outputWriter.ClearView()

	op.app.ResetToHome(ResetOptions{
		PageNames:      []string{pageName, "completion-overlay"},
		ClearDetails:   op.operation == opDelete && op.state == StateSuccess,
		RefreshFiles:   (op.operation == opPopulate || op.operation == opClean || op.operation == opUpdate) && op.state == StateSuccess,
		RestoreFocus:   op.operation != opDeploy || op.state != StateSuccess,
		ForceEnvFocus:  op.operation == opDeploy && op.state == StateSuccess,
		SyncEnvRefresh: (op.operation == opDelete || op.operation == opDeploy) && op.state == StateSuccess,
	})

	// If we were in details and it wasn't a delete, we might need a full update
	// to show changed information (like new GUIs or updated config).
	if op.wasInDetails && op.operation != opDelete && op.state == StateSuccess {
		op.app.detailsPanel.Update(op.savedDetailsName, op.savedDetailsType, op.savedDetailsContext, true)
	}
}

// handleInput handles key events for the progress screen.
func (op *OperationProgress) handleInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEsc {
		if op.state == StateRunning {
			return nil
		}
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
	logs := op.app.outputWriter.GetBuffer()
	cleanLogs := stripANSI(logs)
	op.app.copyToClipboardWithFeedback(cleanLogs, "Logs copied to clipboard!", "Failed to copy logs")
}
