package tui

import (
	"github.com/epos-eu/epos-opensource/db"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// createHome builds the home screen layout with environment lists and details panel.
func (a *App) createHome() *tview.Flex {
	envsFlex := a.createEnvLists()
	a.details = tview.NewTextArea().
		SetBorder(true).
		SetTitle(" [::b]Environment Details ").
		SetTitleColor(DefaultTheme.Secondary)

	home := tview.NewFlex().
		AddItem(envsFlex, 0, 1, true).
		AddItem(a.details, 0, 4, false)

	a.setupHomeInput(envsFlex)
	a.setupFocusHandlers()

	return home
}

// createEnvLists creates the Docker and K8s environment list components.
func (a *App) createEnvLists() *tview.Flex {
	// Docker list
	a.docker = tview.NewList()
	a.docker.SetBorder(true)
	a.docker.SetBorderPadding(1, 1, 1, 1)
	a.docker.SetTitle(" [::b]Docker Environments ")
	a.docker.SetTitleColor(DefaultTheme.Secondary)
	a.docker.SetSelectedBackgroundColor(DefaultTheme.Primary)
	a.docker.SetSelectedTextColor(DefaultTheme.OnPrimary)

	a.dockerEmpty = tview.NewTextView()
	a.dockerEmpty.SetBorder(true)
	a.dockerEmpty.SetBorderPadding(1, 1, 1, 1)
	a.dockerEmpty.SetTitle(" [::b]Docker Environments ")
	a.dockerEmpty.SetTitleColor(DefaultTheme.Secondary)
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
	a.k8s.SetSelectedBackgroundColor(DefaultTheme.Primary)
	a.k8s.SetSelectedTextColor(DefaultTheme.OnPrimary)

	a.k8sEmpty = tview.NewTextView()
	a.k8sEmpty.SetBorder(true)
	a.k8sEmpty.SetBorderPadding(1, 1, 1, 1)
	a.k8sEmpty.SetTitle(" [::b]K8s Environments ")
	a.k8sEmpty.SetTitleColor(DefaultTheme.Secondary)
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
		case event.Key() == tcell.KeyTab:
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

// setupFocusHandlers configures visual feedback when components gain/lose focus.
func (a *App) setupFocusHandlers() {
	// Docker List
	a.docker.SetFocusFunc(func() {
		a.currentEnv = a.dockerFlex
		a.docker.SetBorderColor(DefaultTheme.Primary)
		a.docker.SetSelectedBackgroundColor(DefaultTheme.Primary)
		a.docker.SetSelectedTextColor(DefaultTheme.OnPrimary)
		a.UpdateFooter("[Docker Environments]", KeyDescriptions["docker"])
	})
	a.docker.SetBlurFunc(func() {
		a.docker.SetBorderColor(DefaultTheme.Surface)
		a.docker.SetSelectedBackgroundColor(DefaultTheme.Surface)
		a.docker.SetSelectedTextColor(DefaultTheme.OnSurface)
	})

	// Docker Empty
	a.dockerEmpty.SetFocusFunc(func() {
		a.currentEnv = a.dockerFlex
		a.dockerEmpty.SetBorderColor(DefaultTheme.Primary)
		a.UpdateFooter("[Docker Environments]", KeyDescriptions["docker"])
	})
	a.dockerEmpty.SetBlurFunc(func() {
		a.dockerEmpty.SetBorderColor(DefaultTheme.Surface)
	})

	// K8s List
	a.k8s.SetFocusFunc(func() {
		a.currentEnv = a.k8sFlex
		a.k8s.SetBorderColor(DefaultTheme.Primary)
		a.k8s.SetSelectedBackgroundColor(DefaultTheme.Primary)
		a.k8s.SetSelectedTextColor(DefaultTheme.OnPrimary)
		a.UpdateFooter("[K8s Environments]", KeyDescriptions["k8s"])
	})
	a.k8s.SetBlurFunc(func() {
		a.k8s.SetBorderColor(DefaultTheme.Surface)
		a.k8s.SetSelectedBackgroundColor(DefaultTheme.Surface)
		a.k8s.SetSelectedTextColor(DefaultTheme.OnSurface)
	})

	// K8s Empty
	a.k8sEmpty.SetFocusFunc(func() {
		a.currentEnv = a.k8sFlex
		a.k8sEmpty.SetBorderColor(DefaultTheme.Primary)
		a.UpdateFooter("[K8s Environments]", KeyDescriptions["k8s"])
	})
	a.k8sEmpty.SetBlurFunc(func() {
		a.k8sEmpty.SetBorderColor(DefaultTheme.Surface)
	})

	// Details
	a.details.SetFocusFunc(func() {
		a.details.SetBorderColor(DefaultTheme.Primary)
		a.UpdateFooter("[Environment Details]", KeyDescriptions["details"])
	})
	a.details.SetBlurFunc(func() {
		a.details.SetBorderColor(DefaultTheme.Surface)
	})
}
