package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/epos-eu/epos-opensource/db/sqlc"
)

const (
	activeBorderColor   = "#00ff00"
	inactiveBorderColor = "#2ecc71"
)

type item string

func (i item) FilterValue() string { return string(i) }

type envDelegate struct{ isActive bool }

func (d envDelegate) Height() int                             { return 1 }
func (d envDelegate) Spacing() int                            { return 0 }
func (d envDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d envDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := string(i)

	fn := func(s ...string) string {
		return itemStyle.Render(strings.Join(s, " "))
	}
	if index == m.Index() && d.isActive {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	_, _ = fmt.Fprint(w, fn(str))
}

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("#2ecc71")).Bold(true)
)

type EnvView struct {
	dockerList list.Model
	k8sList    list.Model
	activeList int
}

func NewEnvView(docker []sqlc.Docker, k8s []sqlc.Kubernetes) EnvView {
	activeList := 0
	if len(docker) == 0 && len(k8s) > 0 {
		activeList = 1
	}

	dockerItems := make([]list.Item, len(docker))
	for i, d := range docker {
		dockerItems[i] = item(d.Name)
	}
	dockerActive := activeList == 0
	dockerList := list.New(dockerItems, envDelegate{isActive: dockerActive}, 20, 10)
	dockerList.SetShowStatusBar(false)
	dockerList.SetFilteringEnabled(false)
	dockerList.SetShowHelp(false)
	dockerList.Title = "Docker environments"

	k8sItems := make([]list.Item, len(k8s))
	for i, k := range k8s {
		k8sItems[i] = item(k.Name)
	}
	k8sActive := activeList == 1
	k8sList := list.New(k8sItems, envDelegate{isActive: k8sActive}, 20, 10)
	k8sList.SetShowStatusBar(false)
	k8sList.SetFilteringEnabled(false)
	k8sList.SetShowHelp(false)
	k8sList.Title = "Kubernetes environments"

	return EnvView{
		dockerList: dockerList,
		k8sList:    k8sList,
		activeList: activeList,
	}
}

func (ev EnvView) Update(msg tea.Msg, keys keyMap, sectionWidth, sectionHeight int) EnvView {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Up):
			if ev.activeList == 0 {
				if ev.dockerList.Index() == 0 {
					if len(ev.k8sList.Items()) > 0 {
						ev.activeList = 1
						ev.k8sList.Select(len(ev.k8sList.Items()) - 1)
					}
				} else {
					ev.dockerList.CursorUp()
				}
			} else {
				if ev.k8sList.Index() == 0 {
					if len(ev.dockerList.Items()) > 0 {
						ev.activeList = 0
						ev.dockerList.Select(len(ev.dockerList.Items()) - 1)
					}
				} else {
					ev.k8sList.CursorUp()
				}
			}
		case key.Matches(msg, keys.Down):
			if ev.activeList == 0 {
				if ev.dockerList.Index() == len(ev.dockerList.Items())-1 {
					if len(ev.k8sList.Items()) > 0 {
						ev.activeList = 1
						ev.k8sList.Select(0)
					}
				} else {
					ev.dockerList.CursorDown()
				}
			} else {
				if ev.k8sList.Index() == len(ev.k8sList.Items())-1 {
					if len(ev.dockerList.Items()) > 0 {
						ev.activeList = 0
						ev.dockerList.Select(0)
					}
				} else {
					ev.k8sList.CursorDown()
				}
			}
		}
	case tea.WindowSizeMsg:
		ev.dockerList.SetSize(sectionWidth-2, sectionHeight-2)
		ev.k8sList.SetSize(sectionWidth-2, sectionHeight-2)
	}

	return ev
}

func (ev EnvView) SwitchActive() EnvView {
	ev.activeList = 1 - ev.activeList
	return ev
}

func (ev EnvView) ActiveList() int {
	return ev.activeList
}

func (ev EnvView) View(sectionWidth, sectionHeight int) string {
	dockerActive := ev.activeList == 0
	dockerList := list.New(ev.dockerList.Items(), envDelegate{isActive: dockerActive}, 20, 10)
	dockerList.SetShowStatusBar(false)
	dockerList.SetFilteringEnabled(false)
	dockerList.SetShowHelp(false)
	dockerList.Title = "Docker environments"
	dockerList.SetSize(sectionWidth-2, sectionHeight-2)
	dockerList.Select(ev.dockerList.Index())

	k8sActive := ev.activeList == 1
	k8sList := list.New(ev.k8sList.Items(), envDelegate{isActive: k8sActive}, 20, 10)
	k8sList.SetShowStatusBar(false)
	k8sList.SetFilteringEnabled(false)
	k8sList.SetShowHelp(false)
	k8sList.Title = "Kubernetes environments"
	k8sList.SetSize(sectionWidth-2, sectionHeight-2)
	k8sList.Select(ev.k8sList.Index())

	dockerBorderColor := inactiveBorderColor
	if ev.activeList == 0 {
		dockerBorderColor = activeBorderColor
	}
	dockerSectionStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Width(sectionWidth).
		Height(sectionHeight).
		BorderForeground(lipgloss.Color(dockerBorderColor))
	dockerSection := dockerSectionStyle.Render(dockerList.View())

	k8sBorderColor := inactiveBorderColor
	if ev.activeList == 1 {
		k8sBorderColor = activeBorderColor
	}
	k8sSectionStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Width(sectionWidth).
		Height(sectionHeight).
		BorderForeground(lipgloss.Color(k8sBorderColor))
	k8sSection := k8sSectionStyle.Render(k8sList.View())

	return lipgloss.JoinVertical(lipgloss.Top, dockerSection, k8sSection)
}

func (ev EnvView) SelectedIndex() int {
	if ev.activeList == 0 {
		return ev.dockerList.Index()
	}
	return len(ev.dockerList.Items()) + ev.k8sList.Index()
}
