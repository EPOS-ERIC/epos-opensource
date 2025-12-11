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

// createHome builds the home screen layout with environment lists and details panel.
func (a *App) createHome() *tview.Flex {
	envsFlex := a.createEnvLists()
	a.detailsTable = tview.NewTable()
	a.detailsTable.SetBorders(true)
	a.detailsTable.SetBordersColor(DefaultTheme.Secondary)
	a.detailsTable.SetSelectable(true, true)
	a.detailsTable.SetSelectedStyle(tcell.StyleDefault.Foreground(DefaultTheme.Secondary).Background(DefaultTheme.Background))
	a.detailsTable.Select(0, 2)
	a.detailsTable.SetBorderPadding(1, 1, 0, 0)

	// Action handler helper
	triggerAction := func(row int) {
		if row < 0 || row >= a.detailsTable.GetRowCount() {
			return
		}
		value := a.detailsTable.GetCell(row, 1).Text
		action := a.detailsTable.GetCell(row, 2).Text
		switch action {
		case " Copy ":
			a.copyToClipboard(value)
		case " Open ":
			a.openInBrowser(value)
		}
	}

	// Capture mouse events to handle clicks
	mouseClicked := false
	a.detailsTable.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftClick {
			mouseClicked = true
		}
		return action, event
	})

	// Restrict selection and handle single-click actions
	a.detailsTable.SetSelectionChangedFunc(func(row, column int) {
		if column != 2 {
			a.detailsTable.Select(row, 2)
			return
		}

		if mouseClicked {
			mouseClicked = false
			triggerAction(row)
		}
	})

	// Handle activation (Enter key)
	a.detailsTable.SetSelectedFunc(func(row, column int) {
		triggerAction(row)
	})

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
	a.detailsList.SetTitle(" [::b]Services ")
	a.detailsList.SetTitleColor(DefaultTheme.Secondary)
	updateListStyle(a.detailsList, false)
	a.detailsList.AddItem("Mock Service 1", "", 0, nil)
	a.detailsList.AddItem("Mock Service 2", "", 0, nil)
	a.detailsList.AddItem("Mock Service 3", "", 0, nil)

	a.detailsPlaceholder = tview.NewTextView()
	a.detailsPlaceholder.SetText(DefaultTheme.MutedTag("i") + "\nSelect an environment to view details")
	a.detailsPlaceholder.SetTextAlign(tview.AlignCenter)
	a.detailsPlaceholder.SetDynamicColors(true)
	a.detailsPlaceholder.SetTextColor(DefaultTheme.OnSurface)

	a.details = tview.NewFlex().SetDirection(tview.FlexRow)
	a.details.SetBorder(true)
	a.details.SetBorderColor(DefaultTheme.Surface)
	a.details.SetTitle(" [::b]Environment Details ")
	a.details.SetTitleColor(DefaultTheme.Secondary)
	a.details.SetBorderPadding(1, 0, 1, 1)
	a.details.AddItem(a.detailsPlaceholder, 0, 1, true)

	home := tview.NewFlex().
		AddItem(envsFlex, 0, 1, true).
		AddItem(a.details, 0, 4, false)

	a.homeFlex = home
	a.setupHomeInput(envsFlex)
	a.setupFocusHandlers()

	return home
}

// showPlaceholder shows the placeholder text in the details panel.
func (a *App) showPlaceholder() {
	if a.detailsShown {
		a.details.Clear()
		a.details.AddItem(a.detailsPlaceholder, 0, 1, true)
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
	handler := func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEsc:
			if a.details.HasFocus() {
				a.showPlaceholder()
				a.tview.SetFocus(a.currentEnv)
			}
			return nil
		case event.Key() == tcell.KeyTab:
			if a.details.HasFocus() {
				a.cycleDetailsFocus()
				return nil
			}
			a.switchEnvFocus()
			return nil
		case event.Key() == tcell.KeyEnter:
			if a.details.HasFocus() {
				return event // Let the table handle the Enter key via SetSelectedFunc
			}
			if a.currentEnv == a.dockerFlex && a.docker.GetItemCount() > 0 {
				name := a.SelectedDockerEnv()
				if name != "" {
					a.showDetails(name, "docker")
				}
				return nil
			}
			if a.currentEnv == a.k8sFlex && a.k8s.GetItemCount() > 0 {
				name := a.SelectedK8sEnv()
				if name != "" {
					a.showDetails(name, "k8s")
				}
				return nil
			}
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
		case event.Rune() == '?':
			a.showHelp()
			return nil
		case event.Rune() == 'q':
			a.Quit()
			return nil
		}
		return event
	}

	envsFlex.SetInputCapture(handler)
	a.docker.SetInputCapture(handler)
	a.dockerEmpty.SetInputCapture(handler)
	a.k8s.SetInputCapture(handler)
	a.k8sEmpty.SetInputCapture(handler)
	a.details.SetInputCapture(handler)
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

// cycleDetailsFocus cycles focus between buttons, table, and list in the details view.
func (a *App) cycleDetailsFocus() {
	focus := a.tview.GetFocus()
	switch focus {
	case a.deleteButton:
		a.detailsTable.Select(0, 2)
		a.tview.SetFocus(a.detailsTable)
	case a.cleanButton:
		a.tview.SetFocus(a.deleteButton)
	case a.updateButton:
		a.tview.SetFocus(a.cleanButton)
	case a.populateButton:
		a.tview.SetFocus(a.updateButton)
	case a.detailsTable:
		if rows := a.detailsTable.GetRowCount(); rows > 0 {
			r, _ := a.detailsTable.GetSelection()
			if r < rows-1 {
				a.detailsTable.Select(r+1, 2)
				return
			}
			if r < rows-1 {
				a.detailsTable.Select(r+1, 2)
				return
			}
		}
		a.tview.SetFocus(a.detailsList)
	case a.detailsList:
		a.tview.SetFocus(a.populateButton)
	default:
		// If we are in details but nothing recognized is focused, start at the top
		a.tview.SetFocus(a.populateButton)
	}
}

// showDetails fetches and displays environment details in a table.
func (a *App) showDetails(name, envType string) {
	if !a.detailsShown {
		a.details.Clear()
		a.details.AddItem(a.buttonsFlex, 1, 0, true)
		a.details.AddItem(a.detailsTable, 0, 3, false)
		a.details.AddItem(a.detailsList, 0, 1, false)
		a.detailsShown = true
		updateBoxStyle(a.details, true)
	}

	padString := func(str string) string {
		return " " + str + " "
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
			a.detailsTable.SetCell(0, 0, tview.NewTableCell(padString("Name")).SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(0, 1, tview.NewTableCell(padString(d.Name)).SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.detailsTable.SetCell(0, 2, tview.NewTableCell(padString("Copy")).SetTextColor(DefaultTheme.Secondary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(1, 0, tview.NewTableCell(padString("Directory")).SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(1, 1, tview.NewTableCell(padString(d.Directory)).SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.detailsTable.SetCell(1, 2, tview.NewTableCell(padString("Copy")).SetTextColor(DefaultTheme.Secondary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(2, 0, tview.NewTableCell(padString("API URL")).SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(2, 1, tview.NewTableCell(padString(apiURL)).SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.detailsTable.SetCell(2, 2, tview.NewTableCell(padString("Open")).SetTextColor(DefaultTheme.Secondary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(3, 0, tview.NewTableCell(padString("GUI URL")).SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(3, 1, tview.NewTableCell(padString(d.GuiUrl)).SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.detailsTable.SetCell(3, 2, tview.NewTableCell(padString("Open")).SetTextColor(DefaultTheme.Secondary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(4, 0, tview.NewTableCell(padString("Backoffice URL")).SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(4, 1, tview.NewTableCell(padString(backofficeURL)).SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.detailsTable.SetCell(4, 2, tview.NewTableCell(padString("Open")).SetTextColor(DefaultTheme.Secondary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(5, 0, tview.NewTableCell(padString("API Port")).SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(5, 1, tview.NewTableCell(padString(fmt.Sprintf("%d", d.ApiPort))).SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.detailsTable.SetCell(5, 2, tview.NewTableCell(padString("Copy")).SetTextColor(DefaultTheme.Secondary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(6, 0, tview.NewTableCell(padString("GUI Port")).SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(6, 1, tview.NewTableCell(padString(fmt.Sprintf("%d", d.GuiPort))).SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.detailsTable.SetCell(6, 2, tview.NewTableCell(padString("Copy")).SetTextColor(DefaultTheme.Secondary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(7, 0, tview.NewTableCell(padString("Backoffice Port")).SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(7, 1, tview.NewTableCell(padString(fmt.Sprintf("%d", d.BackofficePort))).SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.detailsTable.SetCell(7, 2, tview.NewTableCell(padString("Copy")).SetTextColor(DefaultTheme.Secondary).SetAttributes(tcell.AttrBold))
		} else {
			a.detailsTable.SetCell(0, 0, tview.NewTableCell(padString(fmt.Sprintf("Error fetching details for %s: %v", name, err))).SetTextColor(DefaultTheme.Destructive))
		}
	case "k8s":
		if k, err := db.GetKubernetesByName(name); err == nil {
			a.detailsTable.SetCell(0, 0, tview.NewTableCell(padString("Name")).SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(0, 1, tview.NewTableCell(padString(k.Name)).SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.detailsTable.SetCell(0, 2, tview.NewTableCell(padString("Copy")).SetTextColor(DefaultTheme.Secondary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(1, 0, tview.NewTableCell(padString("Directory")).SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(1, 1, tview.NewTableCell(padString(k.Directory)).SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.detailsTable.SetCell(1, 2, tview.NewTableCell(padString("Copy")).SetTextColor(DefaultTheme.Secondary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(2, 0, tview.NewTableCell(padString("Context")).SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(2, 1, tview.NewTableCell(padString(k.Context)).SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.detailsTable.SetCell(2, 2, tview.NewTableCell(padString("Copy")).SetTextColor(DefaultTheme.Secondary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(3, 0, tview.NewTableCell(padString("API URL")).SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(3, 1, tview.NewTableCell(padString(k.ApiUrl)).SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.detailsTable.SetCell(3, 2, tview.NewTableCell(padString("Open")).SetTextColor(DefaultTheme.Secondary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(4, 0, tview.NewTableCell(padString("GUI URL")).SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(4, 1, tview.NewTableCell(padString(k.GuiUrl)).SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.detailsTable.SetCell(4, 2, tview.NewTableCell(padString("Open")).SetTextColor(DefaultTheme.Secondary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(5, 0, tview.NewTableCell(padString("Backoffice URL")).SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(5, 1, tview.NewTableCell(padString(k.BackofficeUrl)).SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.detailsTable.SetCell(5, 2, tview.NewTableCell(padString("Open")).SetTextColor(DefaultTheme.Secondary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(6, 0, tview.NewTableCell(padString("Protocol")).SetTextColor(DefaultTheme.Primary).SetAttributes(tcell.AttrBold))
			a.detailsTable.SetCell(6, 1, tview.NewTableCell(padString(k.Protocol)).SetTextColor(DefaultTheme.OnSurface).SetAlign(tview.AlignLeft).SetExpansion(1))
			a.detailsTable.SetCell(6, 2, tview.NewTableCell(padString("Copy")).SetTextColor(DefaultTheme.Secondary).SetAttributes(tcell.AttrBold))
		} else {
			a.detailsTable.SetCell(0, 0, tview.NewTableCell(padString(fmt.Sprintf("Error fetching details for %s: %v", name, err))).SetTextColor(DefaultTheme.Destructive))
		}
	}

	a.tview.SetFocus(a.details)
	a.UpdateFooter("[Environment Details]", KeyDescriptions["details"])
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
		a.UpdateFooter("[Environment Details]", KeyDescriptions["details"])
	})
	a.details.SetBlurFunc(func() {
		if a.detailsShown {
			updateBoxStyle(a.details, true)
		} else {
			updateBoxStyle(a.details, false)
		}
	})

	// Details Table
	a.detailsTable.SetFocusFunc(func() {
		a.detailsTable.SetSelectedStyle(tcell.StyleDefault.Foreground(DefaultTheme.Primary))
	})
	a.detailsTable.SetBlurFunc(func() {
		a.detailsTable.SetSelectedStyle(tcell.StyleDefault.Foreground(DefaultTheme.Secondary).Background(DefaultTheme.Background))
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
