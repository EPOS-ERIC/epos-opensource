package tui

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
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
	app                *App
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
	currentDetailsName string
	currentDetailsType string
	currentDetailsRows []DetailRow
	currentDirectory   string
	detailsShown       bool
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

// buildUI constructs the component layout.
func (dp *DetailsPanel) buildUI() {
	dp.detailsGrid = tview.NewGrid()
	dp.detailsGrid.SetBorders(true)

	dp.nameDirGrid = tview.NewGrid()
	dp.nameDirGrid.SetBorders(true)
	dp.nameDirGrid.SetBordersColor(DefaultTheme.Secondary)
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
func (dp *DetailsPanel) Update(name, envType string, focus bool) {
	dp.currentDetailsName = name
	dp.currentDetailsType = envType

	nameDirGridCount := 2
	detailsGridCount := 2

	switch envType {
	case "docker":
		if d, err := db.GetDockerByName(name); err == nil {
			apiURL, err := url.JoinPath(d.ApiUrl, "ui")
			if err != nil {
				dp.app.ShowError("error joining docker api url")
				log.Printf("error joining docker api url: %v", err)
				return
			}
			backofficeURL, err := url.JoinPath(d.BackofficeUrl, "home")
			if err != nil {
				dp.app.ShowError("error joining docker backoffice url")
				log.Printf("error joining docker backoffice url: %v", err)
				return
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
				{Label: "Backoffice", Value: backofficeURL, IncludeOpen: true},
				{Label: "API", Value: apiURL, IncludeOpen: true},
			}
			detailsGridCount = len(rows)
			dp.currentDetailsRows = rows
			dp.createDetailsRows(rows)
		} else {
			dp.detailsGrid.Clear()
			dp.detailsButtons = nil
			dp.detailsGrid.SetRows(1)
			dp.detailsGrid.SetColumns(1)
			errorTV := tview.NewTextView().SetText(fmt.Sprintf("Error fetching details for %s: %v", name, err)).SetTextColor(DefaultTheme.Destructive)
			dp.detailsGrid.AddItem(errorTV, 0, 0, 1, 1, 0, 0, false)
		}
	case "k8s":
		if k, err := db.GetK8sByName(name); err == nil {
			nameDirRows := []DetailRow{
				{Label: "Name", Value: k.Name, IncludeOpen: false},
				{Label: "Context", Value: k.Context, IncludeOpen: false},
				{Label: "Directory", Value: k.Directory, IncludeOpen: true},
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
				{Label: "GUI", Value: k.GuiUrl, IncludeOpen: true},
				{Label: "Backoffice", Value: k.BackofficeUrl, IncludeOpen: true},
				{Label: "API", Value: k.ApiUrl, IncludeOpen: true},
			}
			detailsGridCount = len(rows)
			dp.currentDetailsRows = rows
			dp.createDetailsRows(rows)
		} else {
			dp.detailsGrid.Clear()
			dp.detailsButtons = nil
			dp.detailsGrid.SetRows(1)
			dp.detailsGrid.SetColumns(1)
			errorTV := tview.NewTextView().SetText(fmt.Sprintf("Error fetching details for %s: %v", name, err)).SetTextColor(DefaultTheme.Destructive)
			dp.detailsGrid.AddItem(errorTV, 0, 0, 1, 1, 0, 0, false)
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

	dp.RefreshFiles()
	if focus {
		dp.app.tview.SetFocus(dp.details)
		dp.app.UpdateFooter("[Environment Details]", "details-"+envType)
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

	if ingestedFiles, err := db.GetIngestedFilesByEnvironment(dp.currentDetailsType, dp.currentDetailsName); err == nil {
		count := len(ingestedFiles)
		if count > 0 {
			dp.detailsList.SetTitle(fmt.Sprintf(" [::b]Ingested Files (%d) ", count))
			for i, file := range ingestedFiles {
				itemText := fmt.Sprintf("%d. %s", i+1, file.FilePath)
				dp.detailsList.AddItem(itemText, "", 0, nil)
			}
			dp.detailsListFlex.AddItem(dp.detailsList, 0, 1, true)
		} else {
			dp.detailsListFlex.AddItem(dp.detailsListEmpty, 0, 1, true)
		}
	} else {
		dp.detailsListEmpty.SetText("\n" + DefaultTheme.DestructiveTag("i") + fmt.Sprintf("Error loading files: %v", err))
		dp.detailsListFlex.AddItem(dp.detailsListEmpty, 0, 1, true)
	}
}

// CycleFocus cycles focus between buttons, grid, and list in the details view.
func (dp *DetailsPanel) CycleFocus() {
	focus := dp.app.tview.GetFocus()
	switch focus {
	case dp.deleteButton:
		if len(dp.nameDirButtons) > 0 {
			dp.app.tview.SetFocus(dp.nameDirButtons[0])
		} else if len(dp.detailsButtons) > 0 {
			dp.app.tview.SetFocus(dp.detailsButtons[0])
		} else {
			dp.app.tview.SetFocus(dp.detailsListFlex)
		}
	case dp.cleanButton:
		dp.app.tview.SetFocus(dp.deleteButton)
	case dp.updateButton:
		dp.app.tview.SetFocus(dp.cleanButton)
	case dp.populateButton:
		dp.app.tview.SetFocus(dp.updateButton)
	case dp.detailsListFlex:
		dp.app.tview.SetFocus(dp.populateButton)
	default:
		// Check if it's a details button
		for i, btn := range dp.detailsButtons {
			if focus == btn {
				if i+1 < len(dp.detailsButtons) {
					dp.app.tview.SetFocus(dp.detailsButtons[i+1])
				} else {
					dp.app.tview.SetFocus(dp.detailsListFlex)
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
						dp.app.tview.SetFocus(dp.detailsListFlex)
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
	case dp.detailsListFlex:
		dp.app.tview.SetFocus(dp.populateButton)
	case dp.deleteButton:
		dp.app.tview.SetFocus(dp.cleanButton)
	case dp.cleanButton:
		dp.app.tview.SetFocus(dp.updateButton)
	case dp.updateButton:
		dp.app.tview.SetFocus(dp.populateButton)
	case dp.populateButton:
		dp.app.tview.SetFocus(dp.detailsListFlex)
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
		dp.app.tview.SetFocus(dp.detailsListFlex)
	}
}

// SetupInput configures key handlers for the details panel.
func (dp *DetailsPanel) SetupInput() {
	handler := func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEsc:
			dp.Clear()
			if dp.app.envList.IsDockerActive() {
				if dp.app.envList.docker.GetItemCount() > 0 {
					dp.app.tview.SetFocus(dp.app.envList.docker)
				} else {
					dp.app.tview.SetFocus(dp.app.envList.dockerEmpty)
				}
			} else {
				if dp.app.envList.k8s.GetItemCount() > 0 {
					dp.app.tview.SetFocus(dp.app.envList.k8s)
				} else {
					dp.app.tview.SetFocus(dp.app.envList.k8sEmpty)
				}
			}
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
			if dp.app.envList.IsDockerActive() {
				dp.app.showCleanConfirm()
				return nil
			}
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
				go func() {
					if err := common.CopyToClipboard(dp.currentDetailsRows[0].Value); err != nil {
						dp.app.tview.QueueUpdateDraw(func() {
							dp.app.ShowError("Failed to copy to clipboard")
						})
					} else {
						dp.app.FlashMessage("Copied to clipboard", 2*time.Second)
					}
				}()
				return nil
			}
		case event.Rune() == 'B':
			if dp.detailsShown && len(dp.currentDetailsRows) > 1 {
				go func() {
					if err := common.CopyToClipboard(dp.currentDetailsRows[1].Value); err != nil {
						dp.app.tview.QueueUpdateDraw(func() {
							dp.app.ShowError("Failed to copy to clipboard")
						})
					} else {
						dp.app.FlashMessage("Copied to clipboard", 2*time.Second)
					}
				}()
				return nil
			}
		case event.Rune() == 'A':
			if dp.detailsShown && len(dp.currentDetailsRows) > 2 {
				go func() {
					if err := common.CopyToClipboard(dp.currentDetailsRows[2].Value); err != nil {
						dp.app.tview.QueueUpdateDraw(func() {
							dp.app.ShowError("Failed to copy to clipboard")
						})
					} else {
						dp.app.FlashMessage("Copied to clipboard", 2*time.Second)
					}
				}()
				return nil
			}
		case event.Rune() == 'e':
			if dp.detailsShown && dp.currentDirectory != "" {
				dp.openValue(dp.currentDirectory)
				return nil
			}
		case event.Rune() == 'E':
			if dp.detailsShown && dp.currentDirectory != "" {
				go func() {
					if err := common.CopyToClipboard(dp.currentDirectory); err != nil {
						dp.app.tview.QueueUpdateDraw(func() {
							dp.app.ShowError("Failed to copy to clipboard")
						})
					} else {
						dp.app.FlashMessage("Copied to clipboard", 2*time.Second)
					}
				}()
				return nil
			}
		}
		return event
	}
	dp.details.SetInputCapture(handler)
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
				log.Printf("Failed to open: %v", err)
			}
		})
	} else {
		dp.app.ShowError("Failed to open")
	}
}

// createGridRows creates the grid rows for details or name/dir
// createGridRows creates the grid rows for details or name/dir
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

	// Remove grid borders for cleaner look
	grid.SetBorders(false)

	rowIndex := 0
	if header != "" {
		// Header
		headerTV := tview.NewTextView().
			SetDynamicColors(true).
			SetText(DefaultTheme.SecondaryTag("b") + header)
		headerTV.SetBorderPadding(0, 0, 1, 1)

		grid.AddItem(headerTV, rowIndex, 0, 1, numColumns, 0, 0, false)
		rowIndex++

		// Separator - use Box that fills available width
		separatorBox := tview.NewBox().
			SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
				// Draw separator line across full width
				style := tcell.StyleDefault.Foreground(DefaultTheme.Muted)
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

		// Add spacing around buttons
		copyBtn := NewStyledButton("Copy", func() {
			go func() {
				if err := common.CopyToClipboard(row.Value); err != nil {
					dp.app.tview.QueueUpdateDraw(func() {
						dp.app.ShowError("Failed to copy to clipboard")
					})
				} else {
					dp.app.FlashMessage("Copied to clipboard", 2*time.Second)
				}
			}()
		})

		// Wrap button in a box with padding to create spacing
		copyBtnBox := tview.NewFlex().SetDirection(tview.FlexColumn)
		copyBtnBox.AddItem(tview.NewBox(), 1, 0, false) // Left padding
		copyBtnBox.AddItem(copyBtn, 0, 1, false)
		copyBtnBox.AddItem(tview.NewBox(), 1, 0, false) // Right padding

		grid.AddItem(labelTV, rowIndex, 0, 1, 1, 0, 0, false)
		grid.AddItem(valueTV, rowIndex, 1, 1, 1, 0, 0, false)
		grid.AddItem(copyBtnBox, rowIndex, 2, 1, 1, 0, 0, false)

		*buttons = append(*buttons, copyBtn)

		if row.IncludeOpen {
			openBtn := NewStyledButton("Open", func() {
				dp.openValue(row.Value)
			})

			// Wrap button in a box with padding
			openBtnBox := tview.NewFlex().SetDirection(tview.FlexColumn)
			openBtnBox.AddItem(tview.NewBox(), 1, 0, false) // Left padding
			openBtnBox.AddItem(openBtn, 0, 1, false)
			openBtnBox.AddItem(tview.NewBox(), 1, 0, false) // Right padding

			grid.AddItem(openBtnBox, rowIndex, 3, 1, 1, 0, 0, false)
			*buttons = append(*buttons, openBtn)
		}

		rowIndex++

		// Add spacing row between items (but not after the last item)
		if i < len(rows)-1 {
			// Add empty box for spacing
			spacingBox := tview.NewBox()
			grid.AddItem(spacingBox, rowIndex, 0, 1, numColumns, 0, 0, false)
			rowIndex++
		}
	}
}
