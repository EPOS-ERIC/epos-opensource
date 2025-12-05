package tui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/epos-eu/epos-opensource/command"
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	colorGreen   = tcell.NewRGBColor(90, 180, 105)
	colorYellow  = tcell.NewRGBColor(229, 161, 14)
	colorBlack   = tcell.NewRGBColor(0, 0, 0)
	colorDefault = tcell.ColorDefault
)

var keyDescriptions = map[string][]string{
	"docker":    {"tab: switch", "↑↓: nav", "n: new", "d: del", "c: clean", "enter: select", "?: help"},
	"k8s":       {"tab: switch", "↑↓: nav", "n: new", "d: del", "c: clean", "enter: select", "?: help"},
	"details":   {"?: help", "q: quit"},
	"newDocker": {""},
}

type App struct {
	app           *tview.Application
	pages         *tview.Pages
	docker        *tview.List
	dockerEmpty   *tview.TextView
	dockerFlex    *tview.Flex
	k8s           *tview.List
	k8sEmpty      *tview.TextView
	k8sFlex       *tview.Flex
	details       *tview.Box
	frame         *tview.Frame
	currentEnv    tview.Primitive
	refreshTicker *time.Ticker
	refreshMutex  sync.Mutex
	outputWriter  *TUIViewWriter
}

func Run() error {
	tuiApp := &App{}

	tuiApp.setupApp()

	// Set up output redirection for all CLI output
	tuiApp.outputWriter = &TUIViewWriter{}
	display.Stdout = tuiApp.outputWriter
	display.Stderr = tuiApp.outputWriter
	command.Stdout = tuiApp.outputWriter
	command.Stderr = tuiApp.outputWriter

	envsFlex := tuiApp.createEnvs()

	details := tuiApp.createDetails()

	home := tview.NewFlex().
		AddItem(envsFlex, 0, 1, true).
		AddItem(details, 0, 4, false)

	tuiApp.setupInputCapture(envsFlex)

	tuiApp.setupFocusHandlers()

	// Create pages and add main UI
	tuiApp.pages = tview.NewPages().
		AddPage("home", home, true, true)

	tuiApp.frame = tview.NewFrame(tuiApp.pages).
		SetBorders(0, 0, 0, 0, 0, 0)
	tuiApp.frame.SetBackgroundColor(colorGreen)

	tuiApp.updateFooter("[Docker Environments]", keyDescriptions["docker"])

	tuiApp.refreshTicker = time.NewTicker(1 * time.Second)
	go func() {
		for range tuiApp.refreshTicker.C {
			tuiApp.app.QueueUpdateDraw(func() {
				tuiApp.refreshLists()
			})
		}
	}()

	return tuiApp.runApp(home)
}

// setupApp initializes the application and sets global styles.
func (t *App) setupApp() {
	t.app = tview.NewApplication()
	tview.Styles.PrimitiveBackgroundColor = colorDefault
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

// createEnvs creates the Docker and K8s environment lists and their flex layout.
func (t *App) createEnvs() *tview.Flex {
	// Create Docker components
	t.docker = tview.NewList()
	t.docker.SetBorder(true)
	t.docker.SetBorderPadding(1, 1, 1, 1)
	t.docker.SetTitle("Docker Environments")
	t.docker.SetTitleColor(colorYellow)
	t.docker.SetSelectedBackgroundColor(colorGreen)
	t.docker.SetSelectedTextColor(colorBlack)

	t.dockerEmpty = tview.NewTextView()
	t.dockerEmpty.SetBorder(true)
	t.dockerEmpty.SetBorderPadding(1, 1, 1, 1)
	t.dockerEmpty.SetTitle("Docker Environments")
	t.dockerEmpty.SetTitleColor(colorYellow)
	t.dockerEmpty.SetTextAlign(tview.AlignCenter)
	t.dockerEmpty.SetDynamicColors(true)
	t.dockerEmpty.SetText("[#808080::i]No Docker environments found")

	t.dockerFlex = tview.NewFlex()

	// Create K8s components
	t.k8s = tview.NewList()
	t.k8s.SetBorder(true)
	t.k8s.SetBorderPadding(1, 1, 1, 1)
	t.k8s.SetTitle("K8s Environments")
	t.k8s.SetTitleColor(colorYellow)
	t.k8s.SetSelectedBackgroundColor(colorGreen)
	t.k8s.SetSelectedTextColor(colorBlack)

	t.k8sEmpty = tview.NewTextView()
	t.k8sEmpty.SetBorder(true)
	t.k8sEmpty.SetBorderPadding(1, 1, 1, 1)
	t.k8sEmpty.SetTitle("K8s Environments")
	t.k8sEmpty.SetTitleColor(colorYellow)
	t.k8sEmpty.SetTextAlign(tview.AlignCenter)
	t.k8sEmpty.SetDynamicColors(true)
	t.k8sEmpty.SetText("[#808080::i]No Kubernetes environments found")

	t.k8sFlex = tview.NewFlex()

	t.refreshLists()

	envsFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(t.dockerFlex, 0, 1, true).
		AddItem(t.k8sFlex, 0, 1, false)

	t.currentEnv = t.dockerFlex
	return envsFlex
}

// refreshLists updates the Docker and K8s lists with current environment data.
func (t *App) refreshLists() {
	t.refreshMutex.Lock()
	defer t.refreshMutex.Unlock()

	// Preserve current selection
	dockerIndex := t.docker.GetCurrentItem()
	k8sIndex := t.k8s.GetCurrentItem()

	// Handle Docker environments
	t.dockerFlex.Clear()
	t.docker.Clear()
	if dockers, err := db.GetAllDocker(); err == nil {
		if len(dockers) == 0 {
			t.dockerFlex.AddItem(t.dockerEmpty, 0, 1, true)
		} else {
			t.dockerFlex.AddItem(t.docker, 0, 1, true)
			for _, d := range dockers {
				t.docker.AddItem("[::b] • "+d.Name+" ", "", 0, nil)
			}
			// Restore selection if valid
			if dockerIndex < t.docker.GetItemCount() {
				t.docker.SetCurrentItem(dockerIndex)
			}
		}
	}

	// Handle K8s environments
	t.k8sFlex.Clear()
	t.k8s.Clear()
	if k8sEnvs, err := db.GetAllKubernetes(); err == nil {
		if len(k8sEnvs) == 0 {
			t.k8sFlex.AddItem(t.k8sEmpty, 0, 1, false)
		} else {
			t.k8sFlex.AddItem(t.k8s, 0, 1, false)
			for _, k := range k8sEnvs {
				t.k8s.AddItem("[::b] • "+k.Name+" ", "", 0, nil)
			}
			// Restore selection if valid
			if k8sIndex < t.k8s.GetItemCount() {
				t.k8s.SetCurrentItem(k8sIndex)
			}
		}
	}
}

// createDetails creates the environment details text area.
func (t *App) createDetails() *tview.Box {
	t.details = tview.NewTextArea().SetBorder(true).SetTitle("Environment Details").SetTitleColor(colorYellow)
	return t.details
}

// updateFooter updates the frame footer with the given text.
func (t *App) updateFooter(sectionText string, keysSlice []string) {
	t.frame.Clear()
	t.frame.AddText("[::b]"+sectionText, false, tview.AlignLeft, colorBlack)
	keysText := strings.Join(keysSlice, ", ")
	t.frame.AddText("[::b]"+keysText, false, tview.AlignCenter, colorBlack)

	versionText := fmt.Sprintf("EPOS Open source [%s]", common.GetVersion())
	t.frame.AddText(createGradient(versionText, tcell.NewRGBColor(0, 255, 0), tcell.NewRGBColor(0, 0, 0)), false, tview.AlignRight, colorDefault)
}

// showHelpModal displays a help modal with key descriptions for all sections.
func (t *App) showHelpModal() {
	// Create table for help content
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(false, false)

	row := 0

	// Docker Environments Section
	table.SetCell(row, 0, tview.NewTableCell("[yellow::b]DOCKER ENVIRONMENTS").
		SetAlign(tview.AlignLeft).
		SetExpansion(1))
	row++

	for _, key := range keyDescriptions["docker"] {
		parts := strings.Split(key, ": ")
		if len(parts) == 2 {
			table.SetCell(row, 0, tview.NewTableCell("  [green::b]"+parts[0]).
				SetAlign(tview.AlignLeft))
			table.SetCell(row, 1, tview.NewTableCell(parts[1]).
				SetAlign(tview.AlignLeft).
				SetExpansion(1))
		} else {
			table.SetCell(row, 0, tview.NewTableCell("  "+key).
				SetAlign(tview.AlignLeft).
				SetExpansion(1))
		}
		row++
	}
	row++ // Empty row

	// K8s Environments Section
	table.SetCell(row, 0, tview.NewTableCell("[yellow::b]KUBERNETES ENVIRONMENTS").
		SetAlign(tview.AlignLeft).
		SetExpansion(1))
	row++

	for _, key := range keyDescriptions["k8s"] {
		parts := strings.Split(key, ": ")
		if len(parts) == 2 {
			table.SetCell(row, 0, tview.NewTableCell("  [green::b]"+parts[0]).
				SetAlign(tview.AlignLeft))
			table.SetCell(row, 1, tview.NewTableCell(parts[1]).
				SetAlign(tview.AlignLeft).
				SetExpansion(1))
		} else {
			table.SetCell(row, 0, tview.NewTableCell("  "+key).
				SetAlign(tview.AlignLeft).
				SetExpansion(1))
		}
		row++
	}
	row++ // Empty row

	// General Section
	table.SetCell(row, 0, tview.NewTableCell("[yellow::b]GENERAL").
		SetAlign(tview.AlignLeft).
		SetExpansion(1))
	row++

	table.SetCell(row, 0, tview.NewTableCell("  [green::b]?").
		SetAlign(tview.AlignLeft))
	table.SetCell(row, 1, tview.NewTableCell("show this help").
		SetAlign(tview.AlignLeft).
		SetExpansion(1))
	row++

	table.SetCell(row, 0, tview.NewTableCell("  [green::b]q").
		SetAlign(tview.AlignLeft))
	table.SetCell(row, 1, tview.NewTableCell("quit application").
		SetAlign(tview.AlignLeft).
		SetExpansion(1))

	// Create footer text
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[::d]Press ESC or q to close")

	// Wrap table and footer in a bordered flex
	content := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true).
		AddItem(footer, 1, 0, false)

	content.SetBorder(true).
		SetBorderColor(colorYellow).
		SetTitle(" Help ").
		SetTitleColor(colorYellow).
		SetBorderPadding(1, 0, 2, 2)

	// Add close handler
	content.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Rune() == 'q' {
			t.pages.RemovePage("help")
			return nil
		}
		return event
	})

	// Add modal to pages and show it with larger dimensions
	t.pages.AddPage("help", modal(content, 1, 1), true, true)
	t.app.SetFocus(content)
}

// setupInputCapture sets up tab switching between environments.
func (t *App) setupInputCapture(envsFlex *tview.Flex) {
	handleInput := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			if t.currentEnv == t.dockerFlex {
				// Switch to K8s - focus the appropriate component
				if t.k8s.GetItemCount() > 0 {
					t.app.SetFocus(t.k8s)
				} else {
					t.app.SetFocus(t.k8sEmpty)
				}
			} else {
				// Switch to Docker - focus the appropriate component
				if t.docker.GetItemCount() > 0 {
					t.app.SetFocus(t.docker)
				} else {
					t.app.SetFocus(t.dockerEmpty)
				}
			}
			return nil
		}
		if event.Rune() == 'n' {
			if t.currentEnv == t.dockerFlex {
				t.newDockerEnv()
				return nil
			}
		}
		if event.Rune() == '?' {
			t.showHelpModal()
			return nil
		}
		if event.Rune() == 'q' {
			t.app.Stop()
			return nil
		}
		return event
	}

	envsFlex.SetInputCapture(handleInput)
	t.docker.SetInputCapture(handleInput)
	t.dockerEmpty.SetInputCapture(handleInput)
	t.k8s.SetInputCapture(handleInput)
	t.k8sEmpty.SetInputCapture(handleInput)
}

// setupFocusHandlers sets up focus and blur handlers for all components.
func (t *App) setupFocusHandlers() {
	// Docker List focus handlers
	t.docker.SetFocusFunc(func() {
		t.currentEnv = t.dockerFlex
		t.docker.SetBorderColor(colorGreen)
		t.docker.SetSelectedBackgroundColor(colorGreen)
		t.docker.SetSelectedTextColor(colorBlack)
		t.updateFooter("[Docker Environments]", keyDescriptions["docker"])
	})
	t.docker.SetBlurFunc(func() {
		t.docker.SetBorderColor(colorDefault)
		t.docker.SetSelectedBackgroundColor(colorDefault)
		t.docker.SetSelectedTextColor(colorDefault)
	})

	// Docker Empty focus handlers
	t.dockerEmpty.SetFocusFunc(func() {
		t.currentEnv = t.dockerFlex
		t.dockerEmpty.SetBorderColor(colorGreen)
		t.updateFooter("[Docker Environments]", keyDescriptions["docker"])
	})
	t.dockerEmpty.SetBlurFunc(func() {
		t.dockerEmpty.SetBorderColor(colorDefault)
	})

	// K8s List focus handlers
	t.k8s.SetFocusFunc(func() {
		t.currentEnv = t.k8sFlex
		t.k8s.SetBorderColor(colorGreen)
		t.k8s.SetSelectedBackgroundColor(colorGreen)
		t.k8s.SetSelectedTextColor(colorBlack)
		t.updateFooter("[K8s Environments]", keyDescriptions["k8s"])
	})
	t.k8s.SetBlurFunc(func() {
		t.k8s.SetBorderColor(colorDefault)
		t.k8s.SetSelectedBackgroundColor(colorDefault)
		t.k8s.SetSelectedTextColor(colorDefault)
	})

	// K8s Empty focus handlers
	t.k8sEmpty.SetFocusFunc(func() {
		t.currentEnv = t.k8sFlex
		t.k8sEmpty.SetBorderColor(colorGreen)
		t.updateFooter("[K8s Environments]", keyDescriptions["k8s"])
	})
	t.k8sEmpty.SetBlurFunc(func() {
		t.k8sEmpty.SetBorderColor(colorDefault)
	})

	// Details focus handlers
	t.details.SetFocusFunc(func() {
		t.details.SetBorderColor(colorGreen)
		t.updateFooter("[Environment Details]", keyDescriptions["details"])
	})
	t.details.SetBlurFunc(func() {
		t.details.SetBorderColor(colorDefault)
	})
}

// runApp runs the application.
func (t *App) runApp(flex *tview.Flex) error {
	if err := t.app.SetRoot(t.frame, true).SetFocus(flex).Run(); err != nil {
		return err
	}
	return nil
}

// interpolateColor interpolates between two colors based on a ratio.
func interpolateColor(start, end tcell.Color, ratio float64) tcell.Color {
	sr, sg, sb := start.RGB()
	er, eg, eb := end.RGB()
	r := uint8(float64(sr) + ratio*(float64(er)-float64(sr)))
	g := uint8(float64(sg) + ratio*(float64(eg)-float64(sg)))
	b := uint8(float64(sb) + ratio*(float64(eb)-float64(sb)))
	return tcell.NewRGBColor(int32(r), int32(g), int32(b))
}

// createGradient creates a gradient text string from start to end color.
func createGradient(text string, startColor, endColor tcell.Color) string {
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
