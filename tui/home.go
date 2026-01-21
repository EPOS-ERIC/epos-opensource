package tui

import (
	"github.com/rivo/tview"
)

// createHome builds the home screen layout with environment lists and details panel.
func (a *App) createHome() *tview.Flex {
	envsFlex := a.envList.GetFlex()
	detailsFlex := a.detailsPanel.GetFlex()

	home := tview.NewFlex().
		AddItem(envsFlex, 0, 1, false). // this has to be false otherwise the focus when clicking on the background of the details will go to the env lists instead of the details
		AddItem(detailsFlex, 0, 4, true)

	a.homeFlex = home
	a.envList.SetupInput()
	a.detailsPanel.SetupInput()
	a.envList.setupFocusHandlers()
	a.detailsPanel.setupFocusHandlers()

	return home
}
