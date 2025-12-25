package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// NewStyledForm creates a form with theme colors.
func NewStyledForm() *tview.Form {
	form := tview.NewForm()
	form.SetFieldBackgroundColor(DefaultTheme.Surface)
	form.SetFieldTextColor(DefaultTheme.Secondary)
	form.SetLabelColor(DefaultTheme.Secondary)
	form.SetButtonBackgroundColor(DefaultTheme.Primary)
	form.SetButtonTextColor(DefaultTheme.OnPrimary)
	form.SetButtonActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))
	form.SetBorderPadding(1, 0, 2, 2)
	return form
}

// NewStyledButton creates a button with theme colors.
func NewStyledButton(label string, selected func()) *tview.Button {
	btn := tview.NewButton(label)
	if selected != nil {
		btn.SetSelectedFunc(selected)
	}
	ApplyButtonStyle(btn)
	return btn
}

// ApplyButtonStyle applies standard button styles.
func ApplyButtonStyle(btn *tview.Button) {
	btn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
	btn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))
}

// NewStyledInactiveButton creates a button with surface colors (used for "Browse Files" etc).
func NewStyledInactiveButton(label string, selected func()) *tview.Button {
	btn := tview.NewButton(label)
	if selected != nil {
		btn.SetSelectedFunc(selected)
	}
	btn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Surface).Foreground(DefaultTheme.Secondary))
	btn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))
	return btn
}

// NewStyledInputField creates an input field with theme colors.
func NewStyledInputField(label, value string) *tview.InputField {
	input := tview.NewInputField()
	input.SetLabel(label)
	input.SetText(value)
	input.SetFieldBackgroundColor(DefaultTheme.Surface)
	input.SetFieldTextColor(DefaultTheme.Secondary)
	input.SetLabelColor(DefaultTheme.Secondary)
	return input
}

// NewStyledList creates a list with theme colors.
func NewStyledList() *tview.List {
	l := tview.NewList()
	l.SetBorder(false)
	l.SetBorderPadding(1, 1, 1, 1)
	updateListStyle(l, false) // Initial state is blurred
	return l
}

// NewStyledTextView creates a text view with theme colors.
func NewStyledTextView() *tview.TextView {
	tv := tview.NewTextView()
	tv.SetBorder(false)
	tv.SetBorderPadding(1, 1, 1, 1)
	tv.SetDynamicColors(true)
	updateBoxStyle(tv, false)
	return tv
}
