package tui

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/db"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// DetailRow represents a row in the details grid.
type DetailRow struct {
	Label       string
	Value       string
	IncludeOpen bool
}

// DetailsPanel manages the right-side information and action panel.
type DetailsPanel struct {
	app                   *App
	details               *tview.Flex
	detailsGrid           *tview.Grid
	detailsButtons        []*tview.Button
	nameDirGrid           *tview.Grid
	nameDirButtons        []*tview.Button
	buttonsFlex           *tview.Flex
	deleteButton          *tview.Button
	cleanButton           *tview.Button
	updateButton          *tview.Button
	populateButton        *tview.Button
	detailsList           *tview.List
	detailsListEmpty      *tview.TextView
	detailsListFlex       *tview.Flex
	detailsEmpty          *tview.TextView
	currentDetailsName    string
	currentDetailsType    string
	currentDetailsContext string
	currentDetailsRows    []DetailRow
	currentDirectory      string
	detailsShown          bool
}

// NewDetailsPanel creates a new DetailsPanel component.
func NewDetailsPanel(app *App) *DetailsPanel {
	dp := &DetailsPanel{app: app}
	dp.buildUI()
	return dp
}

// GetFlex returns the main flex for this component.
func (dp *DetailsPanel) GetFlex() *tview.Flex {
	return dp.details
}

// IsShown returns true if the details panel is currently showing details.
func (dp *DetailsPanel) IsShown() bool {
	return dp.detailsShown
}

// GetCurrentDetailsName returns the current details environment name.
func (dp *DetailsPanel) GetCurrentDetailsName() string {
	return dp.currentDetailsName
}

// GetCurrentDetailsType returns the current details environment type.
func (dp *DetailsPanel) GetCurrentDetailsType() string {
	return dp.currentDetailsType
}

// GetCurrentDetailsContext returns the current details environment context.
func (dp *DetailsPanel) GetCurrentDetailsContext() string {
	return dp.currentDetailsContext
}

// buildUI constructs the component layout.
func (dp *DetailsPanel) buildUI() {
	dp.detailsGrid = tview.NewGrid()

	dp.nameDirGrid = tview.NewGrid()
	dp.nameDirGrid.SetBorderPadding(1, 0, 0, 0)

	dp.nameDirButtons = []*tview.Button{}

	dp.deleteButton = NewStyledButton("Delete", func() {
		dp.app.showDeleteConfirm()
	})

	dp.cleanButton = NewStyledButton("Clean", func() {
		dp.app.showCleanConfirm()
	})

	dp.updateButton = NewStyledButton("Update", func() {
		dp.app.showUpdateForm()
	})

	dp.populateButton = NewStyledButton("Populate", func() {
		dp.app.showPopulateForm()
	})

	dp.buttonsFlex = tview.NewFlex().SetDirection(tview.FlexColumn)
	dp.buttonsFlex.AddItem(tview.NewBox(), 0, 1, false)
	dp.buttonsFlex.AddItem(dp.populateButton, 14, 0, true)
	dp.buttonsFlex.AddItem(tview.NewBox(), 0, 1, false)
	dp.buttonsFlex.AddItem(dp.updateButton, 12, 0, false)
	dp.buttonsFlex.AddItem(tview.NewBox(), 0, 1, false)
	dp.buttonsFlex.AddItem(dp.cleanButton, 11, 0, false)
	dp.buttonsFlex.AddItem(tview.NewBox(), 0, 1, false)
	dp.buttonsFlex.AddItem(dp.deleteButton, 12, 0, false)
	dp.buttonsFlex.AddItem(tview.NewBox(), 0, 1, false)

	dp.detailsList = NewStyledList()
	dp.detailsList.SetBorder(true)
	dp.detailsList.SetTitle(" [::b]Ingested Files ")
	dp.detailsList.SetTitleColor(DefaultTheme.Secondary)
	dp.detailsList.SetBorderPadding(1, 0, 1, 1)

	dp.detailsListEmpty = NewStyledTextView()
	dp.detailsListEmpty.SetBorder(true)
	dp.detailsListEmpty.SetTitle(" [::b]Ingested Files ")
	dp.detailsListEmpty.SetTitleColor(DefaultTheme.Secondary)
	dp.detailsListEmpty.SetTextAlign(tview.AlignCenter)
	dp.detailsListEmpty.SetText("\n" + DefaultTheme.MutedTag("i") + "No ingested files yet")

	dp.detailsListFlex = tview.NewFlex().SetDirection(tview.FlexRow)
	dp.detailsListFlex.AddItem(dp.detailsList, 0, 1, false)

	dp.detailsEmpty = NewStyledTextView()
	dp.detailsEmpty.SetText(DefaultTheme.MutedTag("i") + "\nSelect an environment to view details")
	dp.detailsEmpty.SetTextAlign(tview.AlignCenter)
	dp.detailsEmpty.SetTextColor(DefaultTheme.OnSurface)

	dp.details = tview.NewFlex().SetDirection(tview.FlexRow)
	dp.details.SetBorder(true)
	dp.details.SetBorderColor(DefaultTheme.Surface)
	dp.details.SetTitle(" [::b]Environment Details ")
	dp.details.SetTitleColor(DefaultTheme.Secondary)
	dp.details.SetBorderPadding(1, 0, 1, 1)
	dp.details.AddItem(dp.detailsEmpty, 0, 1, true)
}

// Update fetches and displays environment details in the panel.
func (dp *DetailsPanel) Update(name, envType, context string, focus bool) {
	dp.currentDetailsName = name
	dp.currentDetailsType = envType
	dp.currentDetailsContext = context

	nameDirGridCount := 2
	detailsGridCount := 2

	switch envType {
	case "docker":
		dp.currentDetailsContext = ""
		if d, err := db.GetDockerByName(name); err != nil {
			dp.detailsGrid.Clear()
			dp.detailsButtons = nil
			dp.detailsGrid.SetRows(1)
			dp.detailsGrid.SetColumns(1)
			errorTV := tview.NewTextView().SetText(fmt.Sprintf("Error fetching details for %s: %v", name, err)).SetTextColor(DefaultTheme.Destructive)
			dp.detailsGrid.AddItem(errorTV, 0, 0, 1, 1, 0, 0, false)
		} else {
			apiURL, err := url.JoinPath(d.ApiUrl, "ui")
			if err != nil {
				dp.app.ShowError(fmt.Sprintf("Error joining Docker API URL: %v", err))
				return
			}
			var backofficeURL string
			if d.BackofficeUrl != nil {
				u, err := url.JoinPath(*d.BackofficeUrl, "home")
				if err != nil {
					dp.app.ShowError(fmt.Sprintf("Error joining Docker backoffice URL: %v", err))
					return
				}
				backofficeURL = u
			}
			nameDirRows := []DetailRow{
				{Label: "Name", Value: d.Name, IncludeOpen: false},
				{Label: "Directory", Value: d.Directory, IncludeOpen: true},
			}
			nameDirGridCount = len(nameDirRows)
			dp.createGridRows(dp.nameDirGrid, nameDirRows, &dp.nameDirButtons, "Basic Information")
			for _, row := range nameDirRows {
				if row.Label == "Directory" {
					dp.currentDirectory = row.Value
					break
				}
			}

			rows := []DetailRow{
				{Label: "GUI", Value: d.GuiUrl, IncludeOpen: true},
			}
			if backofficeURL != "" {
				rows = append(rows, DetailRow{Label: "Backoffice", Value: backofficeURL, IncludeOpen: true})
			}
			rows = append(rows, DetailRow{Label: "API", Value: apiURL, IncludeOpen: true})
			detailsGridCount = len(rows)
			dp.currentDetailsRows = rows
			dp.createDetailsRows(rows)
		}
	case string(K8sKey):
		env, err := k8s.GetEnv(name, context)
		if err != nil {
			dp.detailsGrid.Clear()
			dp.detailsButtons = nil
			dp.detailsGrid.SetRows(1)
			dp.detailsGrid.SetColumns(1)
			errorTV := tview.NewTextView().SetText(fmt.Sprintf("Error fetching details for %s: %v", name, err)).SetTextColor(DefaultTheme.Destructive)
			dp.detailsGrid.AddItem(errorTV, 0, 0, 1, 1, 0, 0, false)
		} else {
			urls, err := env.BuildEnvURLs()
			if err != nil {
				dp.detailsGrid.Clear()
				dp.detailsButtons = nil
				dp.detailsGrid.SetRows(1)
				dp.detailsGrid.SetColumns(1)
				errorTV := tview.NewTextView().SetText(fmt.Sprintf("Error building URLs for %s: %v", name, err)).SetTextColor(DefaultTheme.Destructive)
				dp.detailsGrid.AddItem(errorTV, 0, 0, 1, 1, 0, 0, false)
				break
			}

			dp.currentDetailsContext = env.Context
			dp.currentDirectory = ""

			nameDirRows := []DetailRow{
				{Label: "Name", Value: env.Name, IncludeOpen: false},
				{Label: "Context", Value: env.Context, IncludeOpen: false},
			}
			nameDirGridCount = len(nameDirRows)
			dp.createGridRows(dp.nameDirGrid, nameDirRows, &dp.nameDirButtons, "Basic Information")

			rows := []DetailRow{
				{Label: "GUI", Value: urls.GUIURL, IncludeOpen: true},
			}
			if urls.BackofficeURL != nil {
				rows = append(rows, DetailRow{Label: "Backoffice", Value: *urls.BackofficeURL, IncludeOpen: true})
			}
			rows = append(rows, DetailRow{Label: "API", Value: urls.APIURL, IncludeOpen: true})
			detailsGridCount = len(rows)
			dp.currentDetailsRows = rows
			dp.createDetailsRows(rows)
		}
	}

	if !dp.detailsShown {
		// calculate the height of the grids based on the number of rows + header + padding
		nameDirGridSize := (nameDirGridCount * 2) + 3 + 1
		detailsGridSize := (detailsGridCount * 2) + 3
		dp.details.Clear()
		dp.details.AddItem(dp.buttonsFlex, 1, 0, true)
		dp.details.AddItem(dp.nameDirGrid, nameDirGridSize, 0, false)
		dp.details.AddItem(dp.detailsGrid, detailsGridSize, 0, false)
		dp.details.AddItem(dp.detailsListFlex, 0, 1, false)
		dp.detailsShown = true
		updateBoxStyle(dp.details, true)
	}

	dp.RefreshFiles()

	if focus {
		dp.app.tview.SetFocus(dp.details)
		dp.app.UpdateFooter(getDetailsKey(envType))
	}
}

// Clear shows the placeholder text in the details panel.
func (dp *DetailsPanel) Clear() {
	if dp.detailsShown {
		dp.details.Clear()
		dp.details.AddItem(dp.detailsEmpty, 0, 1, true)
		dp.detailsShown = false
		updateBoxStyle(dp.details, false)
		dp.nameDirGrid.Clear()
		dp.nameDirButtons = nil
		dp.currentDetailsName = ""
		dp.currentDetailsType = ""
		dp.currentDetailsContext = ""
		dp.currentDirectory = ""
	}
}

// RefreshFiles refreshes the ingested files list if details are currently shown.
func (dp *DetailsPanel) RefreshFiles() {
	if dp.detailsShown {
		dp.populateIngestedFilesList()
	}
}

// populateIngestedFilesList populates the ingested files list.
func (dp *DetailsPanel) populateIngestedFilesList() {
	dp.detailsList.Clear()
	dp.detailsListFlex.Clear()
	dp.detailsList.SetSelectedFunc(nil)
	dp.detailsList.SetTitle(" [::b]Ingested Files ")
	dp.detailsListEmpty.SetTitle(" [::b]Ingested Files ")

	if dp.currentDetailsType == string(K8sKey) {
		dp.detailsListEmpty.SetText("\n" + DefaultTheme.MutedTag("i") + "Tracking is available only for Docker environments")
		dp.detailsListFlex.AddItem(dp.detailsListEmpty, 0, 1, true)
		dp.syncIngestedFilesFocus()

		return
	}

	dp.detailsListEmpty.SetText("\n" + DefaultTheme.MutedTag("i") + "No ingested files yet")

	if ingestedFiles, err := db.GetIngestedFilesByEnvironment(dp.currentDetailsName); err != nil {
		dp.detailsListEmpty.SetText("\n" + DefaultTheme.DestructiveTag("i") + fmt.Sprintf("Error loading files: %v", err))
		dp.detailsListFlex.AddItem(dp.detailsListEmpty, 0, 1, true)
	} else {
		count := len(ingestedFiles)
		if count > 0 {
			dp.detailsList.SetTitle(fmt.Sprintf(" [::b]Ingested Files (%d) ", count))
			for i, file := range ingestedFiles {
				itemText := fmt.Sprintf("%d. %s", i+1, file.FilePath)
				dp.detailsList.AddItem(itemText, "", 0, nil)
			}
			dp.detailsList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
				parts := strings.SplitN(mainText, ". ", 2)
				if len(parts) == 2 {
					filepath := parts[1]
					dp.openValue(filepath)
				}
			})
			dp.detailsListFlex.AddItem(dp.detailsList, 0, 1, true)
		} else {
			dp.detailsListFlex.AddItem(dp.detailsListEmpty, 0, 1, true)
		}
	}

	dp.syncIngestedFilesFocus()
}

func (dp *DetailsPanel) focusIngestedFiles() {
	if dp.detailsListFlex.GetItemCount() > 0 {
		target := dp.detailsListFlex.GetItem(0)
		if target != nil {
			dp.app.tview.SetFocus(target)

			return
		}
	}

	dp.app.tview.SetFocus(dp.detailsListFlex)
}

func (dp *DetailsPanel) syncIngestedFilesFocus() {
	focus := dp.app.tview.GetFocus()
	if focus == dp.detailsList || focus == dp.detailsListEmpty || focus == dp.detailsListFlex {
		dp.focusIngestedFiles()
	}
}

func (dp *DetailsPanel) focusActiveEnvList() {
	if dp.app.envList.IsDockerActive() {
		if dp.app.envList.docker.GetItemCount() > 0 {
			dp.app.tview.SetFocus(dp.app.envList.docker)
		} else {
			dp.app.tview.SetFocus(dp.app.envList.dockerEmpty)
		}

		return
	}

	if dp.app.envList.k8s.GetItemCount() > 0 {
		dp.app.tview.SetFocus(dp.app.envList.k8s)
	} else {
		dp.app.tview.SetFocus(dp.app.envList.k8sEmpty)
	}
}

// CycleFocus cycles focus between buttons, grid, and list in the details view.
func (dp *DetailsPanel) CycleFocus() {
	focus := dp.app.tview.GetFocus()
	switch focus {
	case dp.deleteButton:
		switch {
		case len(dp.nameDirButtons) > 0:
			dp.app.tview.SetFocus(dp.nameDirButtons[0])
		case len(dp.detailsButtons) > 0:
			dp.app.tview.SetFocus(dp.detailsButtons[0])
		default:
			dp.focusIngestedFiles()
		}
	case dp.cleanButton:
		dp.app.tview.SetFocus(dp.deleteButton)
	case dp.updateButton:
		dp.app.tview.SetFocus(dp.cleanButton)
	case dp.populateButton:
		dp.app.tview.SetFocus(dp.updateButton)
	case dp.detailsListFlex, dp.detailsList, dp.detailsListEmpty:
		dp.app.tview.SetFocus(dp.populateButton)
	default:
		// Check if it's a details button
		for i, btn := range dp.detailsButtons {
			if focus == btn {
				if i+1 < len(dp.detailsButtons) {
					dp.app.tview.SetFocus(dp.detailsButtons[i+1])
				} else {
					dp.focusIngestedFiles()
				}
				return
			}
		}
		// Check if it's a nameDir button
		for i, btn := range dp.nameDirButtons {
			if focus == btn {
				if i+1 < len(dp.nameDirButtons) {
					dp.app.tview.SetFocus(dp.nameDirButtons[i+1])
				} else {
					if len(dp.detailsButtons) > 0 {
						dp.app.tview.SetFocus(dp.detailsButtons[0])
					} else {
						dp.focusIngestedFiles()
					}
				}
				return
			}
		}
		// If not, start at the top
		dp.app.tview.SetFocus(dp.populateButton)
	}
}

// CycleFocusBackward cycles focus backward between buttons, grid, and list in the details view.
func (dp *DetailsPanel) CycleFocusBackward() {
	focus := dp.app.tview.GetFocus()
	switch focus {
	case dp.detailsList, dp.detailsListEmpty, dp.detailsListFlex:
		switch {
		case len(dp.detailsButtons) > 0:
			dp.app.tview.SetFocus(dp.detailsButtons[len(dp.detailsButtons)-1])
		case len(dp.nameDirButtons) > 0:
			dp.app.tview.SetFocus(dp.nameDirButtons[len(dp.nameDirButtons)-1])
		default:
			dp.app.tview.SetFocus(dp.deleteButton)
		}
	case dp.deleteButton:
		dp.app.tview.SetFocus(dp.cleanButton)
	case dp.cleanButton:
		dp.app.tview.SetFocus(dp.updateButton)
	case dp.updateButton:
		dp.app.tview.SetFocus(dp.populateButton)
	case dp.populateButton:
		dp.focusIngestedFiles()
	default:
		// Check if it's a details button
		for i, btn := range dp.detailsButtons {
			if focus == btn {
				if i > 0 {
					dp.app.tview.SetFocus(dp.detailsButtons[i-1])
				} else {
					if len(dp.nameDirButtons) > 0 {
						dp.app.tview.SetFocus(dp.nameDirButtons[len(dp.nameDirButtons)-1])
					} else {
						dp.app.tview.SetFocus(dp.deleteButton)
					}
				}
				return
			}
		}
		// Check if it's a nameDir button
		for i, btn := range dp.nameDirButtons {
			if focus == btn {
				if i > 0 {
					dp.app.tview.SetFocus(dp.nameDirButtons[i-1])
				} else {
					dp.app.tview.SetFocus(dp.deleteButton)
				}
				return
			}
		}
		// If not, start at the end
		dp.focusIngestedFiles()
	}
}

// SetupInput configures key handlers for the details panel.
func (dp *DetailsPanel) SetupInput() {
	handler := func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEsc:
			dp.Clear()
			dp.focusActiveEnvList()
			return nil
		case event.Key() == tcell.KeyTab:
			dp.CycleFocus()
			return nil
		case event.Key() == tcell.KeyBacktab:
			dp.CycleFocusBackward()
			return nil
		case event.Key() == tcell.KeyEnter:
			return event // Let the table handle via SetSelectedFunc
		case event.Rune() == 'd':
			dp.app.showDeleteConfirm()
			return nil
		case event.Rune() == 'c':
			dp.app.showCleanConfirm()
			return nil
		case event.Rune() == 'u':
			dp.app.showUpdateForm()
			return nil
		case event.Rune() == 'p':
			dp.app.showPopulateForm()
			return nil
		case event.Rune() == 'g':
			if dp.detailsShown && len(dp.currentDetailsRows) > 0 {
				dp.openValue(dp.currentDetailsRows[0].Value)
				return nil
			}
		case event.Rune() == 'b':
			if dp.detailsShown && len(dp.currentDetailsRows) > 1 {
				dp.openValue(dp.currentDetailsRows[1].Value)
				return nil
			}
			if dp.detailsShown && len(dp.currentDetailsRows) > 2 {
				dp.openValue(dp.currentDetailsRows[2].Value)
				return nil
			}
		case event.Rune() == 'G':
			if dp.detailsShown && len(dp.currentDetailsRows) > 0 {
				dp.app.copyToClipboardWithFeedback(dp.currentDetailsRows[0].Value, "Copied to clipboard", "Failed to copy to clipboard")
				return nil
			}
		case event.Rune() == 'B':
			if dp.detailsShown && len(dp.currentDetailsRows) > 1 {
				dp.app.copyToClipboardWithFeedback(dp.currentDetailsRows[1].Value, "Copied to clipboard", "Failed to copy to clipboard")
				return nil
			}
		case event.Rune() == 'A':
			if dp.detailsShown && len(dp.currentDetailsRows) > 2 {
				dp.app.copyToClipboardWithFeedback(dp.currentDetailsRows[2].Value, "Copied to clipboard", "Failed to copy to clipboard")
				return nil
			}
		case event.Rune() == 'e':
			if dp.detailsShown && dp.currentDirectory != "" {
				dp.openValue(dp.currentDirectory)
				return nil
			}
		case event.Rune() == 'E':
			if dp.detailsShown && dp.currentDirectory != "" {
				dp.app.copyToClipboardWithFeedback(dp.currentDirectory, "Copied to clipboard", "Failed to copy to clipboard")
				return nil
			}
		case event.Rune() == 'y':
			if dp.app.tview.GetFocus() == dp.detailsList {
				index := dp.detailsList.GetCurrentItem()
				mainText, _ := dp.detailsList.GetItemText(index)
				parts := strings.SplitN(mainText, ". ", 2)
				if len(parts) == 2 {
					filepath := parts[1]
					dp.app.copyToClipboardWithFeedback(filepath, "Copied to clipboard", "Failed to copy to clipboard")
					return nil
				}
			}
		}
		return event
	}
	dp.details.SetInputCapture(handler)
	dp.detailsList.SetInputCapture(handler)
	dp.detailsListEmpty.SetInputCapture(handler)
	dp.detailsListFlex.SetInputCapture(handler)
}

// setupFocusHandlers configures visual feedback when components gain/lose focus.
func (dp *DetailsPanel) setupFocusHandlers() {
	dp.detailsList.SetFocusFunc(func() {
		updateListStyle(dp.detailsList, true)
	})
	dp.detailsList.SetBlurFunc(func() {
		updateListStyle(dp.detailsList, false)
	})
}

// createDetailsRows creates the grid rows for details.
func (dp *DetailsPanel) createDetailsRows(rows []DetailRow) {
	dp.createGridRows(dp.detailsGrid, rows, &dp.detailsButtons, "Environment URLs")
}

// openValue opens the given value (URL, directory, or file) using the appropriate command.
func (dp *DetailsPanel) openValue(value string) {
	var cmd string
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		cmd = dp.app.config.TUI.OpenURLCommand
	} else {
		if info, err := os.Stat(value); err == nil && info.IsDir() {
			cmd = dp.app.config.TUI.OpenDirectoryCommand
		} else {
			cmd = dp.app.config.TUI.OpenFileCommand
		}
	}
	if cmd != "" {
		dp.app.tview.Suspend(func() {
			if err := common.OpenWithCommand(cmd, value); err != nil {
				dp.app.ShowError(fmt.Sprintf("Failed to open: %v", err))
			}
		})
	} else {
		dp.app.ShowError("Failed to open")
	}
}

// createGridRows creates the grid rows for details or name/dir
func (dp *DetailsPanel) createGridRows(grid *tview.Grid, rows []DetailRow, buttons *[]*tview.Button, header string) {
	grid.Clear()
	*buttons = nil

	numColumns := 3
	hasOpenButtons := false
	for _, row := range rows {
		if row.IncludeOpen {
			numColumns = 4
			hasOpenButtons = true
			break
		}
	}

	// Calculate total rows: header + separator + (data rows + spacing rows)
	totalRows := 0
	if header != "" {
		totalRows += 2 // Header + separator
	}
	totalRows += len(rows)
	if len(rows) > 0 {
		totalRows += len(rows) - 1 // Add spacing rows between data rows (not after last)
	}

	rowHeights := make([]int, totalRows)
	for i := range rowHeights {
		rowHeights[i] = 1
	}
	grid.SetRows(rowHeights...)

	if hasOpenButtons {
		grid.SetColumns(15, 0, 10, 10)
	} else {
		grid.SetColumns(15, 0, 10)
	}

	rowIndex := 0
	if header != "" {
		headerTV := tview.NewTextView().
			SetDynamicColors(true).
			SetText(DefaultTheme.SecondaryTag("b") + header)
		headerTV.SetBorderPadding(0, 0, 1, 1)

		grid.AddItem(headerTV, rowIndex, 0, 1, numColumns, 0, 0, false)
		rowIndex++

		// use Box that fills available width for the separator
		separatorBox := tview.NewBox().
			SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
				// Draw separator line across full width
				style := tcell.StyleDefault.Foreground(DefaultTheme.Muted).Background(DefaultTheme.Background)
				for i := range width {
					screen.SetContent(x+i, y, '─', nil, style)
				}
				return x, y, width, height
			})

		grid.AddItem(separatorBox, rowIndex, 0, 1, numColumns, 0, 0, false)
		rowIndex++
	}

	for i, row := range rows {
		labelTV := tview.NewTextView().
			SetDynamicColors(true).
			SetText(DefaultTheme.PrimaryTag("b") + row.Label)
		labelTV.SetBorderPadding(0, 0, 1, 1)

		valueTV := tview.NewTextView().
			SetText(row.Value).
			SetTextColor(DefaultTheme.OnSurface)
		valueTV.SetBorderPadding(0, 0, 1, 1)

		// add spacing around buttons
		copyBtn := NewStyledButton("Copy", func() {
			dp.app.copyToClipboardWithFeedback(row.Value, "Copied to clipboard", "Failed to copy to clipboard")
		})

		// wrap button in a box with padding
		copyBtnBox := tview.NewFlex().SetDirection(tview.FlexColumn)
		copyBtnBox.AddItem(tview.NewBox(), 1, 0, false)
		copyBtnBox.AddItem(copyBtn, 0, 1, false)
		copyBtnBox.AddItem(tview.NewBox(), 1, 0, false)

		grid.AddItem(labelTV, rowIndex, 0, 1, 1, 0, 0, false)
		grid.AddItem(valueTV, rowIndex, 1, 1, 1, 0, 0, false)
		grid.AddItem(copyBtnBox, rowIndex, 2, 1, 1, 0, 0, false)

		*buttons = append(*buttons, copyBtn)

		if row.IncludeOpen {
			openBtn := NewStyledButton("Open", func() {
				dp.openValue(row.Value)
			})

			// wrap button in a box with padding
			openBtnBox := tview.NewFlex().SetDirection(tview.FlexColumn)
			openBtnBox.AddItem(tview.NewBox(), 1, 0, false)
			openBtnBox.AddItem(openBtn, 0, 1, false)
			openBtnBox.AddItem(tview.NewBox(), 1, 0, false)

			grid.AddItem(openBtnBox, rowIndex, 3, 1, 1, 0, 0, false)
			*buttons = append(*buttons, openBtn)
		}

		rowIndex++

		// add spacing row between items but not after the last item
		if i < len(rows)-1 {
			spacingBox := tview.NewBox()
			grid.AddItem(spacingBox, rowIndex, 0, 1, numColumns, 0, 0, false)
			rowIndex++
		}
	}
}
