package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/epos-eu/epos-opensource/config"
)

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	New    key.Binding
	Quit   key.Binding
	Tab    key.Binding
}

func buildKeyMap(cfg config.KeymapsConfig) keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys(cfg.Up...),
			key.WithHelp(strings.Join(cfg.Up, "/"), "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys(cfg.Down...),
			key.WithHelp(strings.Join(cfg.Down, "/"), "move down"),
		),
		Select: key.NewBinding(
			key.WithKeys(cfg.Select...),
			key.WithHelp(strings.Join(cfg.Select, "/"), "select"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		Quit: key.NewBinding(
			key.WithKeys(cfg.Quit...),
			key.WithHelp(strings.Join(cfg.Quit, "/"), "quit"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch section"),
		),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.New, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},    // first column
		{k.Select, k.New}, // second column
		{k.Tab, k.Quit},   // third column
	}
}
