package tui

import (
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TUIViewWriter captures output and routes it to the active TUI view
type TUIViewWriter struct {
	app    *tview.Application
	view   *tview.TextView
	buffer strings.Builder
	mu     sync.Mutex
}

func (w *TUIViewWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Always buffer the output
	w.buffer.Write(p)

	// If we have an active view, update it
	if w.view != nil && w.app != nil {
		w.app.QueueUpdateDraw(func() {
			_, _ = w.view.Write(p)
		})
	}

	return len(p), nil
}

func (w *TUIViewWriter) SetView(app *tview.Application, view *tview.TextView) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.app = app
	w.view = view
	// Write any buffered content to the new view
	if w.view != nil && w.buffer.Len() > 0 {
		w.app.QueueUpdateDraw(func() {
			_, _ = w.view.Write([]byte(w.buffer.String()))
		})
	}
}

func (w *TUIViewWriter) ClearView() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.app = nil
	w.view = nil
}

// DeployScreen manages the deployment interface
type DeployScreen struct {
	app   *tview.Application
	pages *tview.Pages
	// outputView   *tview.TextView
	// statusText   *tview.TextView
	// nameField    *tview.InputField
	// envFileField *tview.InputField
	// pathField    *tview.InputField
	// composeField *tview.InputField
	// pullCheckbox *tview.Checkbox
	writer *TUIViewWriter
	// isDeploying  bool
}

func (t *App) newDockerEnv() {
	t.updateFooter("[New Docker Environment]", keyDescriptions["newDocker"])
	screen := &DeployScreen{
		app:    t.app,
		pages:  t.pages,
		writer: &TUIViewWriter{},
	}

	form := tview.NewForm().
		AddInputField("Name", "", 0, nil, nil)

	// Add input capture for ESC
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Rune() == 'q' {
			screen.returnToMain()
			return nil
		}
		return event
	})

	// Add to pages
	t.pages.AddAndSwitchToPage("deploy", form, true)
}

// a full screen window with an header for the title
// func newForm() {
// }

func (s *DeployScreen) returnToMain() {
	s.writer.ClearView()
	s.pages.RemovePage("deploy")
	s.pages.SwitchToPage("home")
}
