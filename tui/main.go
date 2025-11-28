package tui

// These imports will be used later on the tutorial. If you save the file
// now, Go might complain they are unused, but that's fine.
// You may also need to run `go mod tidy` to download bubbletea and its
// dependencies.
import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/epos-eu/epos-opensource/config"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/db/sqlc"
)

type Model struct {
	docker []sqlc.Docker
	k8s    []sqlc.Kubernetes

	selected int

	width  int
	height int

	envView EnvView

	keys keyMap
	help help.Model

	state string
}

func InitialModel() Model {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("TODO")
	}

	docker, err := db.GetAllDocker()
	if err != nil {
		panic("TODO")
	}

	k8s, err := db.GetAllKubernetes()
	if err != nil {
		panic("TODO")
	}

	envView := NewEnvView(docker, k8s)
	state := "Docker Environments"
	if envView.ActiveList() == 1 {
		state = "K8s Environments"
	}

	return Model{
		docker: docker,
		k8s:    k8s,

		envView: envView,

		keys:  buildKeyMap(cfg.Keymaps),
		help:  help.New(),
		state: state,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Select):
			m.selected = m.envView.SelectedIndex()
		case key.Matches(msg, m.keys.Tab):
			m.envView = m.envView.SwitchActive()
			if m.envView.ActiveList() == 0 {
				m.state = "Docker Environments"
			} else {
				m.state = "K8s Environments"
			}
		default:
			sectionWidth := min((m.width-2)/3, 40)
			sectionHeight := (m.height - 2 - lipgloss.Height(m.statusBarView())) / 2
			m.envView = m.envView.Update(msg, m.keys, sectionWidth, sectionHeight)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		sectionWidth := min((m.width-2)/3, 40)
		sectionHeight := (m.height - 2 - lipgloss.Height(m.statusBarView())) / 2
		m.envView = m.envView.Update(msg, m.keys, sectionWidth, sectionHeight)
	}

	return m, nil
}

func (m Model) View() string {
	statusBar := m.statusBarView()
	statusHeight := lipgloss.Height(statusBar)

	mainWidth := m.width - 2
	mainHeight := m.height - 2 - statusHeight

	mainStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Width(mainWidth).
		Height(mainHeight).
		BorderForeground(lipgloss.Color("#2ecc71"))

	sectionWidth := min((mainWidth-4)/3, 40)
	sectionHeight := (mainHeight - 4) / 2

	envSection := m.envView.View(sectionWidth, sectionHeight)

	detailsStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Width(mainWidth - lipgloss.Width(envSection) - 2).
		Height(mainHeight - 2).
		BorderForeground(lipgloss.Color("#2ecc71"))
	detailsSection := detailsStyle.Render("The selected environment details will show here TODO")

	return lipgloss.JoinVertical(lipgloss.Top, mainStyle.Render(lipgloss.JoinHorizontal(lipgloss.Right, envSection, detailsSection)), statusBar)
}
