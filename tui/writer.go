package tui

import (
	"io"
	"strings"
	"sync"

	"github.com/rivo/tview"
)

// OutputWriter captures command output and routes it to a TUI TextView.
// Converts ANSI escape codes to tview color format.
type OutputWriter struct {
	app        *tview.Application
	view       *tview.TextView
	ansiWriter io.Writer
	buffer     strings.Builder
	mu         sync.Mutex
}

// Write implements io.Writer. Buffers output and writes to the active view.
func (w *OutputWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buffer.Write(p)

	if w.ansiWriter != nil && w.app != nil && w.view != nil {
		data := make([]byte, len(p))
		copy(data, p)
		view := w.view

		w.app.QueueUpdateDraw(func() {
			escaped := escapeBrackets(data)
			_, _ = w.ansiWriter.Write(escaped)
			view.ScrollToEnd()
		})
	}

	return len(p), nil
}

// SetView connects the writer to a TextView for output display.
func (w *OutputWriter) SetView(app *tview.Application, view *tview.TextView) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.app = app
	w.view = view
	w.ansiWriter = tview.ANSIWriter(view)

	// Flush buffered content
	if w.view != nil && w.buffer.Len() > 0 {
		data := w.buffer.String()
		w.app.QueueUpdateDraw(func() {
			_, _ = w.ansiWriter.Write([]byte(data))
		})
	}
}

// ClearView disconnects the writer from the current view.
func (w *OutputWriter) ClearView() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.app = nil
	w.view = nil
	w.ansiWriter = nil
}

// ClearBuffer clears the output buffer.
func (w *OutputWriter) ClearBuffer() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buffer.Reset()
}

// escapeBrackets replaces square brackets with angle brackets for display.
// tview interprets [xxx] as color tags, so we replace non-ANSI brackets.
func escapeBrackets(data []byte) []byte {
	result := make([]byte, 0, len(data))
	for i := range data {
		switch data[i] {
		case '[':
			if i > 0 && data[i-1] == '\033' {
				result = append(result, '[') // ANSI escape
			} else {
				result = append(result, '[') // ANSI escape
			}
		case ']':
			if i > 0 && isANSICodeChar(data[i-1]) {
				result = append(result, ']')
			} else {
				result = append(result, ']')
			}
		default:
			result = append(result, data[i])
		}
	}
	return result
}

// isANSICodeChar returns true if c is part of an ANSI escape sequence.
func isANSICodeChar(c byte) bool {
	return (c >= '0' && c <= '9') || c == ';' || c == 'm' || c == '[' ||
		c == 'A' || c == 'B' || c == 'C' || c == 'D' || c == 'H' || c == 'J' || c == 'K'
}
