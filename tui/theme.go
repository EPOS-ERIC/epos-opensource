// Package tui provides a terminal user interface for managing EPOS environments.
package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Theme colors used throughout the TUI.
var (
	ColorGreen  = tcell.NewRGBColor(90, 180, 105)
	ColorYellow = tcell.NewRGBColor(229, 161, 14)
	ColorBlack  = tcell.NewRGBColor(0, 0, 0)
	ColorRed    = tcell.NewRGBColor(200, 60, 60)
	ColorWhite  = tcell.NewRGBColor(255, 255, 255)
)

// KeyDescriptions maps screen names to their available keyboard shortcuts.
// Used by updateFooter() to show context-sensitive help.
var KeyDescriptions = map[string][]string{
	"docker":  {"tab: switch", "↑↓: nav", "n: new", "d: del", "c: clean", "enter: select", "?: help"},
	"k8s":     {"tab: switch", "↑↓: nav", "n: new", "d: del", "c: clean", "enter: select", "?: help"},
	"details": {"?: help", "q: quit"},
}

// InitStyles sets up global tview styles and border characters.
// Call this once during app initialization.
func InitStyles() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Borders.HorizontalFocus = tview.Borders.Horizontal
	tview.Borders.VerticalFocus = tview.Borders.Vertical
	tview.Borders.TopLeft = '╭'
	tview.Borders.TopRight = '╮'
	tview.Borders.BottomLeft = '╰'
	tview.Borders.BottomRight = '╯'
	tview.Borders.TopLeftFocus = '╭'
	tview.Borders.TopRightFocus = '╮'
	tview.Borders.BottomLeftFocus = '╰'
	tview.Borders.BottomRightFocus = '╯'
}

// CreateGradient creates a gradient text string from start to end color.
// Used for decorative text like version display.
func CreateGradient(text string, startColor, endColor tcell.Color) string {
	result := ""
	runes := []rune(text)
	n := len(runes)
	if n == 0 {
		return ""
	}
	for i, char := range runes {
		ratio := float64(i) / float64(n-1)
		if n == 1 {
			ratio = 0
		}
		col := interpolateColor(startColor, endColor, ratio)
		r, g, b := col.RGB()
		result += fmt.Sprintf("[#%02x%02x%02x::b]%c", r, g, b, char)
	}
	return result
}

// interpolateColor blends between two colors based on a ratio (0.0 to 1.0).
func interpolateColor(start, end tcell.Color, ratio float64) tcell.Color {
	sr, sg, sb := start.RGB()
	er, eg, eb := end.RGB()
	r := uint8(float64(sr) + ratio*(float64(er)-float64(sr)))
	g := uint8(float64(sg) + ratio*(float64(eg)-float64(sg)))
	b := uint8(float64(sb) + ratio*(float64(eb)-float64(sb)))
	return tcell.NewRGBColor(int32(r), int32(g), int32(b))
}
