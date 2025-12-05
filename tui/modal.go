package tui

import "github.com/rivo/tview"

// Returns a new primitive which puts the provided primitive in the center.
// width/height are proportions (e.g., 50 means 50% of screen, 0 means auto-size to content)
func modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, 0, max(1, height), true).
			AddItem(nil, 0, 1, false), 0, max(1, width), true).
		AddItem(nil, 0, 1, false)
}
