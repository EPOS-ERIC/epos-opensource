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
	Background:       tcell.NewRGBColor(20, 20, 20),
	OnBackground:     tcell.NewRGBColor(20, 20, 20),
	Surface:          tcell.NewRGBColor(60, 72, 65),
	OnSurface:        tcell.NewRGBColor(255, 255, 255),
	Muted:            tcell.NewRGBColor(60, 72, 65),
	OnMuted:          tcell.NewRGBColor(20, 20, 20),
	HeaderBackground: tcell.NewRGBColor(40, 48, 43),
}

var FooterTexts = map[ScreenKey]FooterText{
	DockerKey:        DockerFooter,
	K8sKey:           K8sFooter,
	DetailsDockerKey: DetailsFooter,
	DetailsK8sKey:    DetailsFooter,
	FilePickerKey:    FilePickerFooter,
	HomeKey:          HomeFooter,
	PopulateFormKey:  PopulateFooter,
	DeleteConfirmKey: DeleteFooter,
	CleanConfirmKey:  CleanFooter,
	HelpKey:          HelpFooter,
	UpdateFormKey:    UpdateFooter,
	DeployFormKey:    NewFooter,
}

func GetFooterText(key ScreenKey) FooterText {
	if text, ok := FooterTexts[key]; ok {
		return text
	}
	return FooterText("[Unknown]")
}

// KeyHint represents a keyboard shortcut hint.
type KeyHint struct {
	Text         string
	LongText     string
	ShowInFooter bool
	Group        string
}

// KeyHints maps screen names to their available keyboard shortcuts.
// Used by updateFooter() and help screen.
var KeyHints = map[ScreenKey][]KeyHint{
	"docker": {
		{"tab: switch", "tab: switch between docker and k8s environments", true, "Navigation"},
		{"↑↓: nav", "↑↓: navigate through docker environments", true, "Navigation"},
		{"n: new", "n: create a new docker environment", true, "Environment"},
		{"d: del", "d: delete the selected docker environment", true, "Environment"},
		{"c: clean", "c: clean the selected docker environment", true, "Environment"},
		{"u: update", "u: update the selected docker environment", true, "Environment"},
		{"p: populate", "p: populate the selected docker environment", true, "Environment"},
		{"enter: details", "enter: view details of the selected docker environment", true, "Navigation"},
		{"?: help", "?: show help for current context", true, "Generic"},
		{"q: quit", "q: quit the application", true, "Generic"},
	},
	"k8s": {
		{"tab: switch", "tab: switch between docker and k8s environments", true, "Navigation"},
		{"↑↓: nav", "↑↓: navigate through k8s environments", true, "Navigation"},
		{"n: new", "n: create a new k8s environment", true, "Environment"},
		{"d: del", "d: delete the selected k8s environment", true, "Environment"},
		{"c: clean", "c: clean the selected k8s environment", true, "Environment"},
		{"u: update", "u: update the selected k8s environment", true, "Environment"},
		{"p: populate", "p: populate the selected k8s environment", true, "Environment"},
		{"enter: details", "enter: view details of the selected k8s environment", true, "Navigation"},
		{"?: help", "?: show help for current context", true, "Generic"},
		{"q: quit", "q: quit the application", true, "Generic"},
	},
	"details-docker": {
		{"esc: back", "esc: go back to docker environments list", true, "Navigation"},
		{"tab: cycle", "tab: cycle through available actions", true, "Navigation"},
		{"d: del", "d: delete this docker environment", true, "Environment"},
		{"c: clean", "c: clean this docker environment", true, "Environment"},
		{"u: update", "u: update this docker environment", true, "Environment"},
		{"p: populate", "p: populate this docker environment", true, "Environment"},
		{"g: gui", "g: open gui in browser", true, "Browser"},
		{"G: copy gui", "G: copy gui url to clipboard", false, "Browser"},
		{"b: backoffice", "b: open backoffice in browser", true, "Browser"},
		{"B: copy backoffice", "B: copy backoffice url to clipboard", false, "Browser"},
		{"a: api", "a: open api docs in browser", true, "Browser"},
		{"A: copy api", "A: copy api url to clipboard", false, "Browser"},
		{"e: directory", "e: open directory in browser", true, "Browser"},
		{"E: copy directory", "E: copy directory url to clipboard", false, "Browser"},
		{"enter: open file", "enter: open the selected ingested file/directory/url", false, "Browser"},
		{"y: copy file", "y: copy the selected ingested file path to clipboard", false, "Browser"},
		{"?: help", "?: show help for current context", true, "Generic"},
		{"q: quit", "q: quit the application", false, "Generic"},
	},
	"details-k8s": {
		{"esc: back", "esc: go back to k8s environments list", true, "Navigation"},
		{"tab: cycle", "tab: cycle through available actions", true, "Navigation"},
		{"d: del", "d: delete this k8s environment", true, "Environment"},
		{"c: clean", "c: clean this k8s environment", true, "Environment"},
		{"u: update", "u: update this k8s environment", true, "Environment"},
		{"p: populate", "p: populate this k8s environment", true, "Environment"},
		{"g: gui", "g: open gui in browser", true, "Browser"},
		{"G: copy gui", "G: copy gui url to clipboard", false, "Browser"},
		{"b: backoffice", "b: open backoffice in browser", true, "Browser"},
		{"B: copy backoffice", "B: copy backoffice url to clipboard", false, "Browser"},
		{"a: api", "a: open api docs in browser", true, "Browser"},
		{"A: copy api", "A: copy api url to clipboard", false, "Browser"},
		{"e: directory", "e: open directory in browser", true, "Browser"},
		{"E: copy directory", "E: copy directory url to clipboard", false, "Browser"},
		{"enter: open file", "enter: open the selected ingested file/directory/url", false, "Browser"},
		{"y: copy file", "y: copy the selected file path to clipboard", false, "Browser"},
		{"?: help", "?: show help for current context", true, "Generic"},
		{"q: quit", "q: quit the application", false, "Generic"},
	},
	"delete-confirm": {
		{"←→: switch", "", true, "Generic"},
		{"enter: confirm", "", true, "Generic"},
		{"esc: cancel", "", true, "Generic"},
	},
	"deleting": {
		{"please wait...", "", true, "Generic"},
	},
	"delete-complete": {
		{"esc/enter: back", "", true, "Generic"},
	},
	"clean-confirm": {
		{"←→: switch", "", true, "Generic"},
		{"enter: confirm", "", true, "Generic"},
		{"esc: cancel", "", true, "Generic"},
	},
	"cleaning": {
		{"please wait...", "", true, "Generic"},
	},
	"clean-complete": {
		{"esc/enter: back", "", true, "Generic"},
	},
	"update-form": {
		{"tab: next", "", true, "Generic"},
		{"S-tab: prev", "", true, "Generic"},
		{"enter: submit", "", true, "Generic"},
		{"esc: cancel", "", true, "Generic"},
	},
	"updating": {
		{"please wait...", "", true, "Generic"},
	},
	"update-complete": {
		{"esc/enter: back", "", true, "Generic"},
	},
	"populate-confirm": {
		{"←→: switch", "", true, "Generic"},
		{"enter: confirm", "", true, "Generic"},
		{"esc: cancel", "", true, "Generic"},
	},
	"populating": {
		{"please wait...", "", true, "Generic"},
	},
	"populate-complete": {
		{"esc/enter: back", "", true, "Generic"},
	},
	"populate-form": {
		{"tab: next", "", true, "Generic"},
		{"S-tab: prev", "", true, "Generic"},
		{"enter: submit", "", true, "Generic"},
		{"esc: cancel", "", true, "Generic"},
	},
	"file-picker": {
		{"↑↓←→: nav", "↑↓←→: navigate through files and directories", true, "Generic"},
		{"/: search", "/: enter search mode for files", true, "Generic"},
		{"space: mark", "space: mark/unmark files for selection", true, "Generic"},
		{"n/N: next/prev match", "n/N: jump to next/previous search match", true, "Generic"},
		{"enter: submit", "enter: submit selected files", true, "Generic"},
		{"esc: cancel", "esc: cancel file selection", true, "Generic"},
	},
	"deploy-form": {
		{"tab: next", "", true, "Generic"},
		{"S-tab: prev", "", true, "Generic"},
		{"enter: submit", "", true, "Generic"},
		{"esc: cancel", "", true, "Generic"},
	},
	"deploying": {
		{"esc: back (won't stop deployment)", "", true, "Generic"},
	},
	"deploy-complete": {
		{"esc/enter: back", "", true, "Generic"},
	},
	"help": {
		{"↑↓: nav", "↑↓: navigate through help content", true, "Generic"},
		{"esc: close", "esc: close the help screen", true, "Generic"},
	},
}

// getFooterHints returns key hints for the footer, filtered by ShowInFooter.
func getFooterHints(key ScreenKey) []string {
	hints, ok := KeyHints[key]
	if !ok {
		return nil
	}
	var result []string
	for _, h := range hints {
		if h.ShowInFooter {
			result = append(result, h.Text)
		}
	}
	return result
}

// getHelpHints returns grouped key hints for the help screen.
func getHelpHints(key ScreenKey) map[string][]string {
	hints, ok := KeyHints[key]
	if !ok {
		return nil
	}
	result := make(map[string][]string)
	for _, h := range hints {
		group := h.Group
		if group == "" {
			group = "Generic"
		}
		text := h.Text
		if h.LongText != "" {
			text = h.LongText
		}
		result[group] = append(result[group], text)
	}
	return result
}

// InitStyles sets up global tview styles and borders.
func InitStyles() {
	tview.Styles.PrimitiveBackgroundColor = DefaultTheme.Background
	tview.Borders.HorizontalFocus = tview.Borders.Horizontal
	tview.Borders.VerticalFocus = tview.Borders.Vertical
	tview.Borders.TopLeftFocus = tview.Borders.TopLeft
	tview.Borders.TopRightFocus = tview.Borders.TopRight
	tview.Borders.BottomLeftFocus = tview.Borders.BottomLeft
	tview.Borders.BottomRightFocus = tview.Borders.BottomRight
}

// Hex returns the hex string for a color.
func (t *Theme) Hex(color tcell.Color) string {
	r, g, b := color.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// Tag returns a tview color tag.
func (t *Theme) Tag(color tcell.Color, attrs string) string {
	if attrs == "" {
		return fmt.Sprintf("[%s]", t.Hex(color))
	}
	return fmt.Sprintf("[%s::%s]", t.Hex(color), attrs)
}

func (t *Theme) PrimaryTag(attrs string) string   { return t.Tag(t.Primary, attrs) }
func (t *Theme) SecondaryTag(attrs string) string { return t.Tag(t.Secondary, attrs) }

func (t *Theme) DestructiveTag(attrs string) string { return t.Tag(t.Destructive, attrs) }
func (t *Theme) MutedTag(attrs string) string       { return t.Tag(t.Muted, attrs) }

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
