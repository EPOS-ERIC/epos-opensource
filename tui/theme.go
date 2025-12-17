// Package tui provides a terminal user interface for managing EPOS environments.
package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Theme defines the color scheme for the TUI.
type Theme struct {
	Primary          tcell.Color
	OnPrimary        tcell.Color
	Secondary        tcell.Color
	OnSecondary      tcell.Color
	Error            tcell.Color
	OnError          tcell.Color
	Destructive      tcell.Color
	OnDestructive    tcell.Color
	Success          tcell.Color
	OnSuccess        tcell.Color
	Background       tcell.Color
	OnBackground     tcell.Color
	Surface          tcell.Color
	OnSurface        tcell.Color
	Muted            tcell.Color
	OnMuted          tcell.Color
	HeaderBackground tcell.Color
}

// DefaultTheme is the default color scheme.
var DefaultTheme = &Theme{
	Primary:          tcell.NewRGBColor(90, 180, 105),
	OnPrimary:        tcell.NewRGBColor(0, 0, 0),
	Secondary:        tcell.NewRGBColor(229, 161, 14),
	OnSecondary:      tcell.NewRGBColor(0, 0, 0),
	Error:            tcell.NewRGBColor(200, 60, 60),
	OnError:          tcell.NewRGBColor(255, 255, 255),
	Destructive:      tcell.NewRGBColor(200, 60, 60),
	OnDestructive:    tcell.NewRGBColor(255, 255, 255),
	Success:          tcell.NewRGBColor(90, 180, 105),
	OnSuccess:        tcell.NewRGBColor(0, 0, 0),
	Background:       tcell.ColorDefault,
	OnBackground:     tcell.ColorDefault,
	Surface:          tcell.NewRGBColor(60, 72, 65),
	OnSurface:        tcell.NewRGBColor(255, 255, 255),
	Muted:            tcell.NewRGBColor(60, 72, 65),
	OnMuted:          tcell.ColorDefault,
	HeaderBackground: tcell.NewRGBColor(40, 48, 43),
}

const (
	DetailsDockerKey = "details-docker"
	DetailsK8sKey    = "details-k8s"
)

// KeyDescriptions maps screen names to their available keyboard shortcuts.
// Used by updateFooter() to show context-sensitive help.
var KeyDescriptions = map[string][]string{
	"docker":            {"tab: switch", "↑↓: nav", "n: new", "d: del", "c: clean", "u: update", "p: populate", "enter: details", "?: help", "q: quit"},
	"k8s":               {"tab: switch", "↑↓: nav", "enter: details", "?: help", "q: quit"},
	"details-docker":    {"esc: back", "tab: cycle", "d: del", "c: clean", "u: update", "p: populate", "?: help"},
	"details-k8s":       {"esc: back", "tab: cycle", "?: help"},
	"delete-confirm":    {"←→: switch", "enter: confirm", "esc: cancel"},
	"deleting":          {"please wait..."},
	"delete-complete":   {"esc/enter: back"},
	"clean-confirm":     {"←→: switch", "enter: confirm", "esc: cancel"},
	"cleaning":          {"please wait..."},
	"clean-complete":    {"esc/enter: back"},
	"update-confirm":    {"←→: switch", "enter: confirm", "esc: cancel"},
	"updating":          {"please wait..."},
	"update-complete":   {"esc/enter: back"},
	"populate-confirm":  {"←→: switch", "enter: confirm", "esc: cancel"},
	"populating":        {"please wait..."},
	"populate-complete": {"esc/enter: back"},
	"populate-form":     {"tab: next", "S-tab: prev", "enter: submit", "esc: cancel"},
	"file-picker":       {"↑↓←→: nav", "/: search", "space: mark", "enter: submit", "esc: cancel"},
	"deploy-form":       {"tab: next", "S-tab: prev", "enter: submit", "esc: cancel"},
	"deploying":         {"esc: back (won't stop deployment)"},
	"deploy-complete":   {"esc/enter: back"},
	"help":              {"↑↓: nav", "esc/q: close"},
}

// InitStyles sets up global tview styles and border characters.
// Call this once during app initialization.
func InitStyles() {
	tview.Styles.PrimitiveBackgroundColor = DefaultTheme.Background
	tview.Borders.HorizontalFocus = tview.Borders.Horizontal
	tview.Borders.VerticalFocus = tview.Borders.Vertical
	tview.Borders.TopLeftFocus = tview.Borders.TopLeft
	tview.Borders.TopRightFocus = tview.Borders.TopRight
	tview.Borders.BottomLeftFocus = tview.Borders.BottomLeft
	tview.Borders.BottomRightFocus = tview.Borders.BottomRight
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

// Hex returns the hex string for a color.
func (t *Theme) Hex(color tcell.Color) string {
	r, g, b := color.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// Tag returns a tview color tag for a color with attributes.
func (t *Theme) Tag(color tcell.Color, attrs string) string {
	if attrs == "" {
		return fmt.Sprintf("[%s]", t.Hex(color))
	}
	return fmt.Sprintf("[%s::%s]", t.Hex(color), attrs)
}

// Convenience methods for common tags.

func (t *Theme) PrimaryTag(attrs string) string     { return t.Tag(t.Primary, attrs) }
func (t *Theme) SecondaryTag(attrs string) string   { return t.Tag(t.Secondary, attrs) }
func (t *Theme) ErrorTag(attrs string) string       { return t.Tag(t.Error, attrs) }
func (t *Theme) SuccessTag(attrs string) string     { return t.Tag(t.Success, attrs) }
func (t *Theme) DestructiveTag(attrs string) string { return t.Tag(t.Destructive, attrs) }
func (t *Theme) MutedTag(attrs string) string       { return t.Tag(t.Muted, attrs) }
