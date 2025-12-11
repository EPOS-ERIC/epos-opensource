package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showHelp displays a modal with keyboard shortcuts for all screens.
func (a *App) showHelp() {
	a.previousFocus = a.tview.GetFocus()
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(false, false)
	table.SetBorderPadding(1, 0, 2, 2)

	row := 0

	// Docker section
	row = addHelpSection(table, row, "DOCKER ENVIRONMENTS", KeyDescriptions["docker"])
	row++

	// K8s section
	row = addHelpSection(table, row, "KUBERNETES ENVIRONMENTS", KeyDescriptions["k8s"])
	row++

	// General section
	table.SetCell(row, 0, tview.NewTableCell(DefaultTheme.SecondaryTag("b")+"GENERAL").SetAlign(tview.AlignLeft).SetExpansion(1))
	row++
	table.SetCell(row, 0, tview.NewTableCell("  "+DefaultTheme.PrimaryTag("b")+"?").SetAlign(tview.AlignLeft))
	table.SetCell(row, 1, tview.NewTableCell("show this help").SetAlign(tview.AlignRight).SetExpansion(1))
	row++
	table.SetCell(row, 0, tview.NewTableCell("  "+DefaultTheme.PrimaryTag("b")+"q").SetAlign(tview.AlignLeft))
	table.SetCell(row, 1, tview.NewTableCell("quit application").SetAlign(tview.AlignRight).SetExpansion(1))

	// Layout
	content := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true)

	content.SetBorder(true).
		SetBorderColor(DefaultTheme.Secondary).
		SetTitle(" [::b]Help ").
		SetTitleColor(DefaultTheme.Secondary)
		// SetBorderPadding(1, 0, 2, 2)

	// Close handler
	content.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Rune() == 'q' {
			a.pages.RemovePage("help")
			if a.previousFocus != nil {
				a.tview.SetFocus(a.previousFocus)
			}
			return nil
		}
		return event
	})

	a.UpdateFooter("[Help]", KeyDescriptions["help"])
	a.pages.AddPage("help", CenterPrimitive(content, 1, 2), true, true)
	a.tview.SetFocus(content)
}

// addHelpSection adds a section of key descriptions to the help table.
func addHelpSection(table *tview.Table, row int, title string, keys []string) int {
	table.SetCell(row, 0, tview.NewTableCell(DefaultTheme.SecondaryTag("b")+title).SetAlign(tview.AlignLeft).SetExpansion(1))
	row++

	for _, key := range keys {
		parts := strings.Split(key, ": ")
		if len(parts) == 2 {
			table.SetCell(row, 0, tview.NewTableCell("  "+DefaultTheme.PrimaryTag("b")+parts[0]).SetAlign(tview.AlignLeft))
			table.SetCell(row, 1, tview.NewTableCell(parts[1]).SetAlign(tview.AlignRight).SetExpansion(1))
		} else {
			table.SetCell(row, 0, tview.NewTableCell("  "+key).SetAlign(tview.AlignLeft).SetExpansion(1))
		}
		row++
	}
	return row
}
