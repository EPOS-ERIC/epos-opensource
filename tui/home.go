package tui

import (
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strings"

	"github.com/epos-eu/epos-opensource/db"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// DetailRow represents a row in the details grid.
type DetailRow struct {
	Label string
	Value string
}

// createHome builds the home screen layout with environment lists and details panel.
func (a *App) createHome() *tview.Flex {
	envsFlex := a.createEnvLists()
	a.detailsGrid = tview.NewGrid()
	a.detailsGrid.SetBorders(true)

	a.nameDirTable = tview.NewTable()
	a.nameDirTable.SetBorders(true)
	a.nameDirTable.SetBordersColor(DefaultTheme.Secondary)
	a.nameDirTable.SetBorderPadding(1, 1, 0, 0)

	a.deleteButton = tview.NewButton("Delete")
	a.deleteButton.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
	a.deleteButton.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))
	a.deleteButton.SetSelectedFunc(func() {
		if a.currentEnv == a.dockerFlex && a.docker.GetItemCount() > 0 {
			a.showDeleteConfirm()
		}
	})

	a.cleanButton = tview.NewButton("Clean")
	a.cleanButton.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
	a.cleanButton.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))
	a.cleanButton.SetSelectedFunc(func() {
		if a.currentEnv == a.dockerFlex && a.docker.GetItemCount() > 0 {
			a.showCleanConfirm()
		}
	})

	a.updateButton = tview.NewButton("Update")
	a.updateButton.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
	a.updateButton.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))
	a.updateButton.SetSelectedFunc(func() {
		if a.currentEnv == a.dockerFlex && a.docker.GetItemCount() > 0 {
			a.showUpdateForm()
		}
	})

	a.populateButton = tview.NewButton("Populate")
	a.populateButton.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
	a.populateButton.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))
	a.populateButton.SetSelectedFunc(func() {
		if a.currentEnv == a.dockerFlex && a.docker.GetItemCount() > 0 {
			a.showPopulateForm()
		}
	})

	a.buttonsFlex = tview.NewFlex().SetDirection(tview.FlexColumn)
	a.buttonsFlex.AddItem(tview.NewBox(), 0, 1, false) // Spacer
	a.buttonsFlex.AddItem(a.populateButton, 14, 0, true)
	a.buttonsFlex.AddItem(tview.NewBox(), 0, 1, false) // Spacer
	a.buttonsFlex.AddItem(a.updateButton, 12, 0, false)
	a.buttonsFlex.AddItem(tview.NewBox(), 0, 1, false) // Spacer
	a.buttonsFlex.AddItem(a.cleanButton, 11, 0, false)
	a.buttonsFlex.AddItem(tview.NewBox(), 0, 1, false) // Spacer
	a.buttonsFlex.AddItem(a.deleteButton, 12, 0, false)
	a.buttonsFlex.AddItem(tview.NewBox(), 0, 1, false) // Spacer

	a.detailsList = tview.NewList()
	a.detailsList.SetBorder(true)
	a.detailsList.SetTitle(" [::b]Ingested Files ")
	a.detailsList.SetTitleColor(DefaultTheme.Secondary)
	updateListStyle(a.detailsList, false)
	a.detailsList.AddItem("/path/to/ttl1.ttl", "", 0, nil)
	a.detailsList.AddItem("/path/to/ttl2.ttl", "", 0, nil)
	a.detailsList.AddItem("/path/to/ttl3.ttl", "", 0, nil)

	a.detailsEmpty = tview.NewTextView()
	a.detailsEmpty.SetText(DefaultTheme.MutedTag("i") + "\nSelect an environment to view details")
	a.detailsEmpty.SetTextAlign(tview.AlignCenter)
	a.detailsEmpty.SetDynamicColors(true)
	a.detailsEmpty.SetTextColor(DefaultTheme.OnSurface)

	a.details = tview.NewFlex().SetDirection(tview.FlexRow)
	a.details.SetBorder(true)
	a.details.SetBorderColor(DefaultTheme.Surface)
	a.details.SetTitle(" [::b]Environment Details ")
	a.details.SetTitleColor(DefaultTheme.Secondary)
	a.details.SetBorderPadding(1, 0, 1, 1)
	a.details.AddItem(a.detailsEmpty, 0, 1, true)

	home := tview.NewFlex().
		AddItem(envsFlex, 0, 1, true).
		AddItem(a.details, 0, 4, false)

	a.homeFlex = home
	a.setupHomeInput(envsFlex)
	a.setupFocusHandlers()

	return home
}

// clearDetailsPanel shows the placeholder text in the details panel.
func (a *App) clearDetailsPanel() {
	if a.detailsShown {
		a.details.Clear()
		a.details.AddItem(a.detailsEmpty, 0, 1, true)
		a.detailsShown = false
		updateBoxStyle(a.details, false)
	}
}

// copyToClipboard copies the given text to the clipboard.
func (a *App) copyToClipboard(text string) {
	text = strings.Trim(text, " ")
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	err := cmd.Run()
	if err != nil {
		a.ShowError("Failed to copy to clipboard")
	}
}

// openInBrowser opens the given URL in the default browser.
func (a *App) openInBrowser(url string) {
	url = strings.Trim(url, " ")
	if err := exec.Command("open", url).Run(); err != nil {
		a.ShowError("Failed to open in browser")
		log.Printf("error opening in browser: %v", err)
	}
}

// createEnvLists creates the Docker and K8s environment list components.
func (a *App) createEnvLists() *tview.Flex {
	// Docker list
	a.docker = tview.NewList()
	a.docker.SetBorder(true)
	a.docker.SetBorderPadding(1, 1, 1, 1)
	a.docker.SetTitle(" [::b]Docker Environments ")
	a.docker.SetTitleColor(DefaultTheme.Secondary)
	updateListStyle(a.docker, false)

	a.dockerEmpty = tview.NewTextView()
	a.dockerEmpty.SetBorder(true)
	a.dockerEmpty.SetBorderPadding(1, 1, 1, 1)
	a.dockerEmpty.SetTitle(" [::b]Docker Environments ")
	a.dockerEmpty.SetTitleColor(DefaultTheme.Secondary)
	updateBoxStyle(a.dockerEmpty, false)
	a.dockerEmpty.SetTextAlign(tview.AlignCenter)
	a.dockerEmpty.SetDynamicColors(true)
	a.dockerEmpty.SetText(DefaultTheme.MutedTag("i") + "No Docker environments found")

	a.dockerFlex = tview.NewFlex()

	// K8s list
	a.k8s = tview.NewList()
	a.k8s.SetBorder(true)
	a.k8s.SetBorderPadding(1, 1, 1, 1)
	a.k8s.SetTitle(" [::b]K8s Environments ")
	a.k8s.SetTitleColor(DefaultTheme.Secondary)
	updateListStyle(a.k8s, false)

	a.k8sEmpty = tview.NewTextView()
	a.k8sEmpty.SetBorder(true)
	a.k8sEmpty.SetBorderPadding(1, 1, 1, 1)
	a.k8sEmpty.SetTitle(" [::b]K8s Environments ")
	a.k8sEmpty.SetTitleColor(DefaultTheme.Secondary)
	updateBoxStyle(a.k8sEmpty, false)
	a.k8sEmpty.SetTextAlign(tview.AlignCenter)
	a.k8sEmpty.SetDynamicColors(true)
	a.k8sEmpty.SetText(DefaultTheme.MutedTag("i") + "No Kubernetes environments found")

	a.k8sFlex = tview.NewFlex()

	// Initial data load
	a.refreshLists()

	// Layout
	envsFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.dockerFlex, 0, 1, true).
		AddItem(a.k8sFlex, 0, 1, false)

	a.currentEnv = a.dockerFlex
	return envsFlex
}

// refreshLists updates the Docker and K8s lists from the database.
func (a *App) refreshLists() {
	a.refreshMutex.Lock()
	defer a.refreshMutex.Unlock()

	// Preserve current selection
	dockerIndex := a.docker.GetCurrentItem()
	k8sIndex := a.k8s.GetCurrentItem()

	// Docker environments
	a.dockerFlex.Clear()
	a.docker.Clear()
	a.dockerEnvs = nil
	if dockers, err := db.GetAllDocker(); err == nil {
		if len(dockers) == 0 {
			a.dockerFlex.AddItem(a.dockerEmpty, 0, 1, true)
		} else {
			a.dockerFlex.AddItem(a.docker, 0, 1, true)
			for _, d := range dockers {
				a.docker.AddItem("[::b] • "+d.Name+"  ", "", 0, nil)
				a.dockerEnvs = append(a.dockerEnvs, d.Name)
			}
			if dockerIndex < a.docker.GetItemCount() {
				a.docker.SetCurrentItem(dockerIndex)
			}
		}
	}

	// K8s environments
	a.k8sFlex.Clear()
	a.k8s.Clear()
	a.k8sEnvs = nil
	if k8sEnvs, err := db.GetAllKubernetes(); err == nil {
		if len(k8sEnvs) == 0 {
			a.k8sFlex.AddItem(a.k8sEmpty, 0, 1, false)
		} else {
			a.k8sFlex.AddItem(a.k8s, 0, 1, false)
			for _, k := range k8sEnvs {
				a.k8s.AddItem("[::b] • "+k.Name+" ", "", 0, nil)
				a.k8sEnvs = append(a.k8sEnvs, k.Name)
			}
			if k8sIndex < a.k8s.GetItemCount() {
				a.k8s.SetCurrentItem(k8sIndex)
			}
		}
	}
}

// SelectedDockerEnv returns the currently selected Docker environment name, or empty if none.
func (a *App) SelectedDockerEnv() string {
	idx := a.docker.GetCurrentItem()
	if idx >= 0 && idx < len(a.dockerEnvs) {
		return a.dockerEnvs[idx]
	}
	return ""
}

// SelectedK8sEnv returns the currently selected K8s environment name, or empty if none.
func (a *App) SelectedK8sEnv() string {
	idx := a.k8s.GetCurrentItem()
	if idx >= 0 && idx < len(a.k8sEnvs) {
		return a.k8sEnvs[idx]
	}
	return ""
}

// setupHomeInput configures keyboard handlers for the home screen.
func (a *App) setupHomeInput(envsFlex *tview.Flex) {
	a.setupRootInput(envsFlex)
	a.setupListInput(a.docker, true)
	a.setupListInput(a.k8s, false)
	a.setupEmptyInput(a.dockerEmpty)
	a.setupEmptyInput(a.k8sEmpty)
	a.setupDetailsInput(a.details)
}

// switchEnvFocus toggles focus between Docker and K8s lists.
func (a *App) switchEnvFocus() {
	if a.currentEnv == a.dockerFlex {
		if a.k8s.GetItemCount() > 0 {
			a.tview.SetFocus(a.k8s)
		} else {
			a.tview.SetFocus(a.k8sEmpty)
		}
	} else {
		if a.docker.GetItemCount() > 0 {
			a.tview.SetFocus(a.docker)
		} else {
			a.tview.SetFocus(a.dockerEmpty)
		}
	}
}

// cycleDetailsFocus cycles focus between buttons, grid, and list in the details view.
func (a *App) cycleDetailsFocus() {
	focus := a.tview.GetFocus()
	switch focus {
	case a.deleteButton:
		a.tview.SetFocus(a.nameDirTable)
	case a.cleanButton:
		a.tview.SetFocus(a.deleteButton)
	case a.updateButton:
		a.tview.SetFocus(a.cleanButton)
	case a.populateButton:
		a.tview.SetFocus(a.updateButton)
	case a.nameDirTable:
		if len(a.detailsButtons) > 0 {
			a.tview.SetFocus(a.detailsButtons[0])
		} else {
			a.tview.SetFocus(a.detailsList)
		}
	case a.detailsList:
		a.tview.SetFocus(a.populateButton)
	default:
		// Check if it's a details button
		for i, btn := range a.detailsButtons {
			if focus == btn {
				if i+1 < len(a.detailsButtons) {
					a.tview.SetFocus(a.detailsButtons[i+1])
				} else {
					a.tview.SetFocus(a.detailsList)
				}
				return
			}
		}
		// If not, start at the top
		a.tview.SetFocus(a.populateButton)
	}
}

// cycleDetailsFocusBackward cycles focus backward between buttons, grid, and list in the details view.
func (a *App) cycleDetailsFocusBackward() {
	focus := a.tview.GetFocus()
	switch focus {
	case a.detailsList:
		if len(a.detailsButtons) > 0 {
			a.tview.SetFocus(a.detailsButtons[len(a.detailsButtons)-1])
		} else {
			a.tview.SetFocus(a.nameDirTable)
		}
	case a.nameDirTable:
		a.tview.SetFocus(a.deleteButton)
	case a.deleteButton:
		a.tview.SetFocus(a.cleanButton)
	case a.cleanButton:
		a.tview.SetFocus(a.updateButton)
	case a.updateButton:
		a.tview.SetFocus(a.populateButton)
	case a.populateButton:
		a.tview.SetFocus(a.detailsList)
	default:
		// Check if it's a details button
		for i, btn := range a.detailsButtons {
			if focus == btn {
				if i > 0 {
					a.tview.SetFocus(a.detailsButtons[i-1])
				} else {
					a.tview.SetFocus(a.nameDirTable)
				}
				return
			}
		}
		// If not, start at the end
		a.tview.SetFocus(a.detailsList)
	}
}

// createDetailsRows creates the grid rows for details.
func (a *App) createDetailsRows(rows []DetailRow) {
	a.detailsGrid.Clear()
	a.detailsButtons = nil

	// Set up rows: one row per detail item, each 1 cell tall (no border padding)
	rowHeights := make([]int, len(rows))
	for i := range rowHeights {
		rowHeights[i] = 1
	}
	a.detailsGrid.SetRows(rowHeights...)

	// Set up columns:
	// Col 0: Label (fixed width ~15 chars)
	// Col 1: Value (flexible, takes remaining space)
	// Col 2: Copy button (fixed width ~8 chars)
	// Col 3: Open button (fixed width ~8 chars)
	a.detailsGrid.SetColumns(15, 0, 8, 8)

	for i, row := range rows {
		// Create label with no extra padding
		labelTV := tview.NewTextView().
			SetText("[::b]" + row.Label).
			SetTextColor(DefaultTheme.Primary).
			SetDynamicColors(true)
		labelTV.SetBorderPadding(0, 0, 1, 1)

		// Create value with no extra padding
		valueTV := tview.NewTextView().
			SetText(row.Value).
			SetTextColor(DefaultTheme.OnSurface)
		valueTV.SetBorderPadding(0, 0, 1, 1)

		// Create buttons
		copyBtn := tview.NewButton("Copy").
			SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary)).
			SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))
		copyBtn.SetSelectedFunc(func() { a.copyToClipboard(row.Value) })

		openBtn := tview.NewButton("Open").
			SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary)).
			SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))
		openBtn.SetSelectedFunc(func() { a.openInBrowser(row.Value) })

		a.detailsGrid.AddItem(labelTV, i, 0, 1, 1, 0, 0, false)
		a.detailsGrid.AddItem(valueTV, i, 1, 1, 1, 0, 0, false)
		a.detailsGrid.AddItem(copyBtn, i, 2, 1, 1, 0, 0, false)
		a.detailsGrid.AddItem(openBtn, i, 3, 1, 1, 0, 0, false)

		a.detailsButtons = append(a.detailsButtons, copyBtn, openBtn)
	}
	// a.detailsGrid.SetBackgroundColor(DefaultTheme.OnSurface)
	a.detailsGrid.SetBordersColor(DefaultTheme.Secondary)
}

// showDetails fetches and displays environment details in a grid.
func (a *App) showDetails(name, envType string) {
	if !a.detailsShown {
		a.details.Clear()
		a.details.AddItem(a.buttonsFlex, 1, 0, true)
		a.details.AddItem(a.nameDirTable, 0, 1, false)
		a.details.AddItem(a.detailsGrid, 0, 1, false)
		a.details.AddItem(a.detailsList, 0, 1, false)
		a.detailsShown = true
		updateBoxStyle(a.details, true)
	}

	switch envType {
	case "docker":
		if d, err := db.GetDockerByName(name); err == nil {
			apiURL, err := url.JoinPath(d.ApiUrl, "ui")
			if err != nil {
				a.ShowError("error joining docker api url")
				log.Printf("error joining docker api url: %v", err)
				return
			}
			backofficeURL, err := url.JoinPath(d.BackofficeUrl, "home")
			if err != nil {
				a.ShowError("error joining docker backoffice url")
				log.Printf("error joining docker backoffice url: %v", err)
				return
			}
			a.nameDirTable.SetCell(0, 0, tview.NewTableCell(" Name ").SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.nameDirTable.SetCell(0, 1, tview.NewTableCell(" "+d.Name+" ").SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.nameDirTable.SetCell(1, 0, tview.NewTableCell(" Directory ").SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.nameDirTable.SetCell(1, 1, tview.NewTableCell(" "+d.Directory+" ").SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			rows := []DetailRow{
				{Label: "API URL", Value: apiURL},
				{Label: "GUI URL", Value: d.GuiUrl},
				{Label: "Backoffice URL", Value: backofficeURL},
			}
			a.createDetailsRows(rows)
		} else {
			a.detailsGrid.Clear()
			a.detailsButtons = nil
			a.detailsGrid.SetRows(1)
			a.detailsGrid.SetColumns(1)
			errorTV := tview.NewTextView().SetText(fmt.Sprintf("Error fetching details for %s: %v", name, err)).SetTextColor(DefaultTheme.Destructive)
			a.detailsGrid.AddItem(errorTV, 0, 0, 1, 1, 0, 0, false)
		}
	case "k8s":
		if k, err := db.GetKubernetesByName(name); err == nil {
			a.nameDirTable.SetCell(0, 0, tview.NewTableCell(" Name ").SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.nameDirTable.SetCell(0, 1, tview.NewTableCell(" "+k.Name+" ").SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.nameDirTable.SetCell(1, 0, tview.NewTableCell(" Directory ").SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.nameDirTable.SetCell(1, 1, tview.NewTableCell(" "+k.Directory+" ").SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			rows := []DetailRow{
				{Label: "API URL", Value: k.ApiUrl},
				{Label: "GUI URL", Value: k.GuiUrl},
				{Label: "Backoffice URL", Value: k.BackofficeUrl},
			}
			a.createDetailsRows(rows)
		} else {
			a.detailsGrid.Clear()
			a.detailsButtons = nil
			a.detailsGrid.SetRows(1)
			a.detailsGrid.SetColumns(1)
			errorTV := tview.NewTextView().SetText(fmt.Sprintf("Error fetching details for %s: %v", name, err)).SetTextColor(DefaultTheme.Destructive)
			a.detailsGrid.AddItem(errorTV, 0, 0, 1, 1, 0, 0, false)
		}
	}

	a.tview.SetFocus(a.details)
	a.UpdateFooter("[Environment Details]", KeyDescriptions["details-"+envType])
}

// setupRootInput configures global key handlers for the home screen root.
func (a *App) setupRootInput(envsFlex *tview.Flex) {
	handler := func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyTab, event.Key() == tcell.KeyBacktab:
			a.switchEnvFocus()
			return nil
		case event.Rune() == 'n':
			if a.currentEnv == a.dockerFlex {
				a.showDeployForm()
				return nil
			}
		case event.Rune() == 'd':
			if a.currentEnv == a.dockerFlex && a.docker.GetItemCount() > 0 {
				a.showDeleteConfirm()
				return nil
			}
		case event.Rune() == 'c':
			if a.currentEnv == a.dockerFlex && a.docker.GetItemCount() > 0 {
				a.showCleanConfirm()
				return nil
			}
		case event.Rune() == 'u':
			if a.currentEnv == a.dockerFlex && a.docker.GetItemCount() > 0 {
				a.showUpdateForm()
				return nil
			}
		case event.Rune() == 'p':
			if a.currentEnv == a.dockerFlex && a.docker.GetItemCount() > 0 {
				a.showPopulateForm()
				return nil
			}
		}
		return event
	}
	envsFlex.SetInputCapture(handler)
}

// setupDetailsInput configures key handlers for the details panel.
func (a *App) setupDetailsInput(details *tview.Flex) {
	handler := func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEsc:
			a.clearDetailsPanel()
			if a.currentEnv == a.dockerFlex {
				if a.docker.GetItemCount() > 0 {
					a.tview.SetFocus(a.docker)
				} else {
					a.tview.SetFocus(a.dockerEmpty)
				}
			} else {
				if a.k8s.GetItemCount() > 0 {
					a.tview.SetFocus(a.k8s)
				} else {
					a.tview.SetFocus(a.k8sEmpty)
				}
			}
			return nil
		case event.Key() == tcell.KeyTab:
			a.cycleDetailsFocus()
			return nil
		case event.Key() == tcell.KeyBacktab:
			a.cycleDetailsFocusBackward()
			return nil
		case event.Key() == tcell.KeyEnter:
			return event // Let the table handle via SetSelectedFunc
		case event.Rune() == 'd':
			if a.currentEnv == a.dockerFlex {
				a.showDeleteConfirm()
				return nil
			}
		case event.Rune() == 'c':
			if a.currentEnv == a.dockerFlex {
				a.showCleanConfirm()
				return nil
			}
		case event.Rune() == 'u':
			if a.currentEnv == a.dockerFlex {
				a.showUpdateForm()
				return nil
			}
		case event.Rune() == 'p':
			if a.currentEnv == a.dockerFlex {
				a.showPopulateForm()
				return nil
			}
		}
		return event
	}
	details.SetInputCapture(handler)
}

// setupListInput configures key handlers for environment lists.
func (a *App) setupListInput(list *tview.List, isDocker bool) {
	handler := func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEnter:
			if isDocker && list.GetItemCount() > 0 {
				name := a.SelectedDockerEnv()
				if name != "" {
					a.showDetails(name, "docker")
				}
				return nil
			}
			if !isDocker && a.k8s.GetItemCount() > 0 {
				name := a.SelectedK8sEnv()
				if name != "" {
					a.showDetails(name, "k8s")
				}
				return nil
			}
		}
		return event
	}
	list.SetInputCapture(handler)
}

// setupEmptyInput configures key handlers for empty state views (bubbles unhandled keys).
func (a *App) setupEmptyInput(empty *tview.TextView) {
	handler := func(event *tcell.EventKey) *tcell.EventKey {
		return event
	}
	empty.SetInputCapture(handler)
}

// setupFocusHandlers configures visual feedback when components gain/lose focus.
func (a *App) setupFocusHandlers() {
	// Docker List
	a.docker.SetFocusFunc(func() {
		a.currentEnv = a.dockerFlex
		updateListStyle(a.docker, true)
		a.UpdateFooter("[Docker Environments]", KeyDescriptions["docker"])
	})
	a.docker.SetBlurFunc(func() {
		updateListStyle(a.docker, false)
	})

	// Docker Empty
	a.dockerEmpty.SetFocusFunc(func() {
		a.currentEnv = a.dockerFlex
		updateBoxStyle(a.dockerEmpty, true)
		a.UpdateFooter("[Docker Environments]", KeyDescriptions["docker"])
	})
	a.dockerEmpty.SetBlurFunc(func() {
		updateBoxStyle(a.dockerEmpty, false)
	})

	// K8s List
	a.k8s.SetFocusFunc(func() {
		a.currentEnv = a.k8sFlex
		updateListStyle(a.k8s, true)
		a.UpdateFooter("[K8s Environments]", KeyDescriptions["k8s"])
	})
	a.k8s.SetBlurFunc(func() {
		updateListStyle(a.k8s, false)
	})

	// K8s Empty
	a.k8sEmpty.SetFocusFunc(func() {
		a.currentEnv = a.k8sFlex
		updateBoxStyle(a.k8sEmpty, true)
		a.UpdateFooter("[K8s Environments]", KeyDescriptions["k8s"])
	})
	a.k8sEmpty.SetBlurFunc(func() {
		updateBoxStyle(a.k8sEmpty, false)
	})

	// Details
	a.details.SetFocusFunc(func() {
		updateBoxStyle(a.details, true)
		key := DetailsK8sKey
		if a.currentEnv == a.dockerFlex {
			key = DetailsDockerKey
		}
		a.UpdateFooter("[Environment Details]", KeyDescriptions[key])
	})
	a.details.SetBlurFunc(func() {
		if a.detailsShown {
			updateBoxStyle(a.details, true)
		} else {
			updateBoxStyle(a.details, false)
		}
	})

	// Name and Directory Table
	a.nameDirTable.SetFocusFunc(func() {
		updateBoxStyle(a.nameDirTable, true)
	})
	a.nameDirTable.SetBlurFunc(func() {
		updateBoxStyle(a.nameDirTable, false)
	})
}

// boxLike checks for SetBorderColor satisfaction to support List, TextView, etc.
type boxLike interface {
	SetBorderColor(tcell.Color) *tview.Box
}

// updateBoxStyle sets the border color based on focus state.
func updateBoxStyle(b boxLike, active bool) {
	if active {
		b.SetBorderColor(DefaultTheme.Primary)
	} else {
		b.SetBorderColor(DefaultTheme.Surface)
	}
}

// updateListStyle sets border and selection colors based on focus state.
func updateListStyle(l *tview.List, active bool) {
	updateBoxStyle(l, active)
	if active {
		l.SetSelectedBackgroundColor(DefaultTheme.Primary)
		l.SetSelectedTextColor(DefaultTheme.OnPrimary)
	} else {
		l.SetSelectedBackgroundColor(DefaultTheme.Surface)
		l.SetSelectedTextColor(DefaultTheme.OnSurface)
	}
}
