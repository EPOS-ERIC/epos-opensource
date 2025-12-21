package tui

import (
	"github.com/rivo/tview"
)

// createHome builds the home screen layout with environment lists and details panel.
func (a *App) createHome() *tview.Flex {
	envsFlex := a.envList.GetFlex()
	detailsFlex := a.detailsPanel.GetFlex()

	home := tview.NewFlex().
		AddItem(envsFlex, 0, 1, true).
		AddItem(detailsFlex, 0, 4, false)

	a.homeFlex = home
	a.envList.SetupInput(envsFlex)
	a.detailsPanel.SetupInput(detailsFlex)
	a.envList.setupFocusHandlers()
	a.detailsPanel.setupFocusHandlers()

	return home
}
