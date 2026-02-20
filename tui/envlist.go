package tui

import (
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/db"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// EnvList manages the left-side navigation for environment selection.
type EnvList struct {
	app                *App
	docker             *tview.List
	dockerEmpty        *tview.TextView
	dockerFlex         *tview.Flex
	dockerFlexInner    *tview.Flex
	dockerEnvs         []string
	k8s                *tview.List
	k8sEmpty           *tview.TextView
	k8sFlex            *tview.Flex
	k8sFlexInner       *tview.Flex
	k8sEnvs            []string
	createNewButton    *tview.Button
	buttonFlex         *tview.Flex
	createNewButtonK8s *tview.Button
	buttonFlexK8s      *tview.Flex
	currentEnv         tview.Primitive
	flex               *tview.Flex
}

// NewEnvList creates a new EnvList component.
func NewEnvList(app *App) *EnvList {
	el := &EnvList{app: app}
	el.flex = el.buildUI()
	return el
}

// buildUI constructs the component layout.
func (el *EnvList) buildUI() *tview.Flex {
	el.docker = NewStyledList()

	el.dockerEmpty = NewStyledTextView()
	el.dockerEmpty.SetTextAlign(tview.AlignCenter)
	el.dockerEmpty.SetText(DefaultTheme.MutedTag("i") + "No Docker environments found")

	el.dockerFlexInner = tview.NewFlex().SetDirection(tview.FlexRow)

	el.dockerFlex = tview.NewFlex().SetDirection(tview.FlexRow).AddItem(el.dockerFlexInner, 0, 1, true)
	el.dockerFlex.SetBorder(true)
	el.dockerFlex.SetTitle(" [::b]Docker Environments ")
	el.dockerFlex.SetTitleColor(DefaultTheme.Secondary)

	el.k8s = NewStyledList()

	el.k8sEmpty = NewStyledTextView()
	el.k8sEmpty.SetTextAlign(tview.AlignCenter)
	el.k8sEmpty.SetText(DefaultTheme.MutedTag("i") + "No K8s environments found")

	el.k8sFlexInner = tview.NewFlex().SetDirection(tview.FlexRow)

	el.k8sFlex = tview.NewFlex().SetDirection(tview.FlexRow).AddItem(el.k8sFlexInner, 0, 1, true)
	el.k8sFlex.SetBorder(true)
	el.k8sFlex.SetTitle(" [::b]K8s Environments ")
	el.k8sFlex.SetTitleColor(DefaultTheme.Secondary)
	el.k8sFlex.SetBorderColor(DefaultTheme.Surface)

	el.createNewButton = NewStyledButton("Create New Environment", func() {
		el.app.showDeployForm()
	})

	el.buttonFlex = tview.NewFlex().SetDirection(tview.FlexColumn)
	el.buttonFlex.AddItem(tview.NewBox(), 0, 1, false)
	el.buttonFlex.AddItem(el.createNewButton, 26, 0, true)
	el.buttonFlex.AddItem(tview.NewBox(), 0, 1, false)

	el.createNewButtonK8s = NewStyledButton("Create New Environment", func() {
		el.app.showDeployForm()
	})

	el.buttonFlexK8s = tview.NewFlex().SetDirection(tview.FlexColumn)
	el.buttonFlexK8s.AddItem(tview.NewBox(), 0, 1, false)
	el.buttonFlexK8s.AddItem(el.createNewButtonK8s, 26, 0, true)
	el.buttonFlexK8s.AddItem(tview.NewBox(), 0, 1, false)

	el.Refresh()

	envsFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(el.dockerFlex, 0, 1, true).
		AddItem(el.k8sFlex, 0, 1, false)

	el.currentEnv = el.dockerFlex
	return envsFlex
}

// GetFlex returns the main flex for this component.
func (el *EnvList) GetFlex() *tview.Flex {
	return el.flex
}

// Refresh updates the lists from the database, preserving current focus/selection.
func (el *EnvList) Refresh() {
	el.app.refreshMutex.Lock()
	defer el.app.refreshMutex.Unlock()

	dockerIndex := el.docker.GetCurrentItem()
	k8sIndex := el.k8s.GetCurrentItem()

	el.dockerFlexInner.Clear()
	el.docker.Clear()
	el.dockerEnvs = nil
	if dockers, err := db.GetAllDocker(); err == nil {
		if len(dockers) == 0 {
			el.dockerFlexInner.AddItem(el.dockerEmpty, 0, 1, false)
		} else {
			el.dockerFlexInner.AddItem(el.docker, 0, 1, true)
			for _, d := range dockers {
				el.docker.AddItem("[::b] • "+d.Name+"  ", "", 0, nil)
				el.dockerEnvs = append(el.dockerEnvs, d.Name)
			}
			if dockerIndex < el.docker.GetItemCount() {
				el.docker.SetCurrentItem(dockerIndex)
			}
		}
		el.dockerFlexInner.AddItem(el.buttonFlex, 1, 0, true)
	} else {
		el.app.ShowError("Failed to load Docker environments")
	}

	el.k8sFlexInner.Clear()
	el.k8s.Clear()
	el.k8sEnvs = nil
	if k8sEnvs, err := db.GetAllK8s(); err == nil {
		if len(k8sEnvs) == 0 {
			el.k8sFlexInner.AddItem(el.k8sEmpty, 0, 1, false)
		} else {
			el.k8sFlexInner.AddItem(el.k8s, 0, 1, true)
			for _, k := range k8sEnvs {
				el.k8s.AddItem("[::b] • "+k.Name+" ", "", 0, nil)
				el.k8sEnvs = append(el.k8sEnvs, k.Name)
			}
			if k8sIndex < el.k8s.GetItemCount() {
				el.k8s.SetCurrentItem(k8sIndex)
			}
		}
		el.k8sFlexInner.AddItem(el.buttonFlexK8s, 1, 0, true)
	} else {
		el.app.ShowError("Failed to load K8s environments")
	}
}

// GetSelected returns the currently selected environment name and whether it's Docker.
func (el *EnvList) GetSelected() (string, bool) {
	if el.currentEnv == el.dockerFlex {
		idx := el.docker.GetCurrentItem()
		if idx >= 0 && idx < len(el.dockerEnvs) {
			return el.dockerEnvs[idx], true
		}
	} else {
		idx := el.k8s.GetCurrentItem()
		if idx >= 0 && idx < len(el.k8sEnvs) {
			return el.k8sEnvs[idx], false
		}
	}
	return "", false
}

// IsDockerActive returns true if the Docker list is currently active.
func (el *EnvList) IsDockerActive() bool {
	return el.currentEnv == el.dockerFlex
}

// SwitchFocus toggles focus between Docker and K8s lists.
func (el *EnvList) SwitchFocus() {
	if el.currentEnv == el.dockerFlex {
		if el.k8s.GetItemCount() > 0 {
			el.app.tview.SetFocus(el.k8s)
		} else {
			el.app.tview.SetFocus(el.createNewButtonK8s)
		}
	} else {
		if el.docker.GetItemCount() > 0 {
			el.app.tview.SetFocus(el.docker)
		} else {
			el.app.tview.SetFocus(el.createNewButton)
		}
	}
}

// SetInitialFocus sets the focus to the default list or button.
func (el *EnvList) SetInitialFocus() {
	if el.docker.GetItemCount() > 0 {
		el.app.tview.SetFocus(el.docker)
	} else {
		el.app.tview.SetFocus(el.createNewButton)
	}
}

// FocusActiveList sets the focus to the currently active environment list (Docker or K8s).
// If the list is empty, it focuses the "Create New" button instead.
func (el *EnvList) FocusActiveList() {
	if el.IsDockerActive() {
		if el.docker.GetItemCount() > 0 {
			el.app.tview.SetFocus(el.docker)
		} else {
			el.app.tview.SetFocus(el.createNewButton)
		}
	} else {
		if el.k8s.GetItemCount() > 0 {
			el.app.tview.SetFocus(el.k8s)
		} else {
			el.app.tview.SetFocus(el.createNewButtonK8s)
		}
	}
}

// SetupInput configures keyboard and focus handlers.
func (el *EnvList) SetupInput() {
	el.setupRootInput(el.flex)
	el.setupListInput(el.docker, true)
	el.setupListInput(el.k8s, false)
	el.setupEmptyInput(el.dockerEmpty)
	el.setupEmptyInput(el.k8sEmpty)
	el.setupFocusHandlers()
}

// setupRootInput configures global key handlers for the home screen root.
func (el *EnvList) setupRootInput(envsFlex *tview.Flex) {
	handler := func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyTab, event.Key() == tcell.KeyBacktab:
			el.SwitchFocus()
			return nil
		case event.Rune() == 'n':
			el.app.showDeployForm()
			return nil
		case event.Rune() == 'd':
			if el.IsDockerActive() && el.docker.GetItemCount() > 0 {
				el.app.showDeleteConfirm()
				return nil
			}
			if !el.IsDockerActive() && el.k8s.GetItemCount() > 0 {
				el.app.showDeleteConfirm()
				return nil
			}
		case event.Rune() == 'c':
			if el.IsDockerActive() && el.docker.GetItemCount() > 0 {
				el.app.showCleanConfirm()
				return nil
			}
		case event.Rune() == 'u':
			if el.IsDockerActive() && el.docker.GetItemCount() > 0 {
				el.app.showUpdateForm()
				return nil
			}
			if !el.IsDockerActive() && el.k8s.GetItemCount() > 0 {
				el.app.showUpdateForm()
				return nil
			}
		case event.Rune() == 'p':
			if el.IsDockerActive() && el.docker.GetItemCount() > 0 {
				el.app.showPopulateForm()
				return nil
			}
			if !el.IsDockerActive() && el.k8s.GetItemCount() > 0 {
				el.app.showPopulateForm()
				return nil
			}
		}
		return event
	}
	envsFlex.SetInputCapture(handler)
}

// setupListInput configures key handlers for environment lists.
func (el *EnvList) setupListInput(list *tview.List, isDocker bool) {
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if isDocker && list.GetItemCount() > 0 {
			name := el.dockerEnvs[index]
			el.app.detailsPanel.Update(name, "docker", true)
		} else if !isDocker && el.k8s.GetItemCount() > 0 {
			name := el.k8sEnvs[index]
			el.app.detailsPanel.Update(name, "k8s", true)
		}
	})

	// No InputCapture needed for Enter as SetSelectedFunc handles it
}

// setupEmptyInput configures key handlers for empty state views.
func (el *EnvList) setupEmptyInput(empty *tview.TextView) {
	handler := func(event *tcell.EventKey) *tcell.EventKey {
		return event
	}
	empty.SetInputCapture(handler)
}

// setupFocusHandlers manages visual feedback for focus/blur.
func (el *EnvList) setupFocusHandlers() {
	el.docker.SetFocusFunc(func() {
		el.currentEnv = el.dockerFlex
		updateListStyle(el.docker, true)
		updateBoxStyle(el.dockerFlex, true)
		el.app.UpdateFooter(DockerKey)
		el.app.detailsPanel.Clear()
	})
	el.docker.SetBlurFunc(func() {
		updateListStyle(el.docker, false)
		updateBoxStyle(el.dockerFlex, false)
	})

	// Docker Empty
	el.dockerEmpty.SetFocusFunc(func() {
		el.currentEnv = el.dockerFlex
		updateBoxStyle(el.dockerFlex, true)
		el.app.UpdateFooter(DockerKey)
		el.app.detailsPanel.Clear()
	})
	el.dockerEmpty.SetBlurFunc(func() {
		updateBoxStyle(el.dockerFlex, false)
	})

	// Create New Button
	el.createNewButton.SetFocusFunc(func() {
		el.currentEnv = el.dockerFlex
		updateBoxStyle(el.dockerFlex, true)
		el.app.UpdateFooter(DockerKey)
		el.app.detailsPanel.Clear()
	})
	el.createNewButton.SetBlurFunc(func() {
		updateBoxStyle(el.dockerFlex, false)
	})

	// K8s List
	el.k8s.SetFocusFunc(func() {
		el.currentEnv = el.k8sFlex
		updateListStyle(el.k8s, true)
		updateBoxStyle(el.k8sFlex, true)
		el.app.UpdateFooter(K8sKey)
		el.app.detailsPanel.Clear()
	})
	el.k8s.SetBlurFunc(func() {
		updateListStyle(el.k8s, false)
		updateBoxStyle(el.k8sFlex, false)
	})

	// K8s Empty
	el.k8sEmpty.SetFocusFunc(func() {
		el.currentEnv = el.k8sFlex
		updateBoxStyle(el.k8sEmpty, true)
		updateBoxStyle(el.k8sFlex, true)
		el.app.UpdateFooter(K8sKey)
		el.app.detailsPanel.Clear()
	})
	el.k8sEmpty.SetBlurFunc(func() {
		updateBoxStyle(el.k8sEmpty, false)
		updateBoxStyle(el.k8sFlex, false)
	})

	// Create New Button K8s
	el.createNewButtonK8s.SetFocusFunc(func() {
		el.currentEnv = el.k8sFlex
		updateBoxStyle(el.k8sFlex, true)
		el.app.UpdateFooter(K8sKey)
		el.app.detailsPanel.Clear()
	})
	el.createNewButtonK8s.SetBlurFunc(func() {
		updateBoxStyle(el.k8sFlex, false)
	})
}
