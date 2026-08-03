package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	minimumTerminalWidth  = 170
	minimumTerminalHeight = 25
)

type terminalSizeGuard struct {
	*tview.Box
	app     *App
	content tview.Primitive
	warning *tview.TextView
}

func newTerminalSizeGuard(app *App, content tview.Primitive) *terminalSizeGuard {
	return &terminalSizeGuard{
		Box:     tview.NewBox(),
		app:     app,
		content: content,
		warning: newTerminalSizeWarning(),
	}
}

func (g *terminalSizeGuard) Draw(screen tcell.Screen) {
	x, y, width, height := g.GetRect()
	g.app.terminalTooSmall = width < minimumTerminalWidth || height < minimumTerminalHeight

	if g.app.terminalTooSmall {
		g.warning.SetText(terminalSizeWarningText(width, height))
		g.warning.SetRect(x, y, width, height)
		g.warning.Draw(screen)
		return
	}

	g.content.SetRect(x, y, width, height)
	g.content.Draw(screen)
}

func (g *terminalSizeGuard) Focus(delegate func(p tview.Primitive)) {
	g.content.Focus(delegate)
}

func (g *terminalSizeGuard) HasFocus() bool {
	return g.content.HasFocus() || g.Box.HasFocus()
}

func (g *terminalSizeGuard) Blur() {
	g.Box.Blur()
	g.content.Blur()
}

func (g *terminalSizeGuard) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	if g.app.terminalTooSmall {
		return nil
	}

	return g.content.InputHandler()
}

func (g *terminalSizeGuard) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	if g.app.terminalTooSmall {
		return func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
			return true, nil
		}
	}

	return g.content.MouseHandler()
}

func (g *terminalSizeGuard) PasteHandler() func(text string, setFocus func(p tview.Primitive)) {
	if g.app.terminalTooSmall {
		return nil
	}

	return g.content.PasteHandler()
}

func newTerminalSizeWarning() *tview.TextView {
	warning := NewStyledTextView()
	warning.SetBorder(true).
		SetBorderColor(DefaultTheme.Destructive).
		SetTitle(" [::b]Terminal Too Small ").
		SetTitleColor(DefaultTheme.Destructive)
	warning.SetTextAlign(tview.AlignCenter)
	warning.SetBackgroundColor(DefaultTheme.Background)

	return warning
}

func terminalSizeWarningText(width, height int) string {
	return fmt.Sprintf("\n%s\n\nCurrent size: %s%dx%d[-]\nMinimum size: %s%dx%d[-]\n\nIncrease terminal window size or reduce terminal zoom.",
		DefaultTheme.DestructiveTag("b")+"Terminal window is too small",
		DefaultTheme.PrimaryTag("b"), width, height,
		DefaultTheme.PrimaryTag("b"), minimumTerminalWidth, minimumTerminalHeight)
}
