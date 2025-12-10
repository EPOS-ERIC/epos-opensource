package tui

import (
	"strings"
	"sync"

	"github.com/rivo/tview"
)

// OutputWriter captures command output and routes it to a TUI TextView.
// Converts ANSI escape codes to tview color format.
type OutputWriter struct {
	app    *tview.Application
	view   *tview.TextView
	buffer strings.Builder
	mu     sync.Mutex
}

// Write implements io.Writer. Buffers output and writes to the active view.
func (w *OutputWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buffer.Write(p)

	if w.app != nil && w.view != nil {
		w.writeToView(string(p), true)
	}

	return len(p), nil
}

// SetView connects the writer to a TextView for output display.
// Flushes any buffered content that arrived before the view was connected.
func (w *OutputWriter) SetView(app *tview.Application, view *tview.TextView) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.app = app
	w.view = view

	// Flush buffered content that arrived before view was ready
	if w.view != nil && w.buffer.Len() > 0 {
		w.writeToView(w.buffer.String(), false)
	}
}

// writeToView processes text and writes it to the view.
// Must be called with mutex held.
func (w *OutputWriter) writeToView(text string, scroll bool) {
	view := w.view
	w.app.QueueUpdateDraw(func() {
		escaped := tview.Escape(text)
		translated := tview.TranslateANSI(escaped)
		_, _ = view.Write([]byte(translated))
		if scroll {
			view.ScrollToEnd()
		}
	})
}

// ClearView disconnects the writer from the current view.
func (w *OutputWriter) ClearView() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.app = nil
	w.view = nil
}

// ClearBuffer clears the output buffer.
func (w *OutputWriter) ClearBuffer() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buffer.Reset()
}
