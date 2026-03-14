package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ameistad/grove/internal/config"
	"github.com/ameistad/grove/internal/tui/screens"
)

type screen int

const (
	screenHome screen = iota
	screenCreate
)

type model struct {
	cfg      config.Config
	repoRoot string
	branch   string
	screen   screen
	home     screens.HomeScreen
	create   screens.CreateScreen
}

func Run(cfg config.Config, root, branch string) error {
	m := newModel(cfg, root, branch)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func newModel(cfg config.Config, root, branch string) model {
	return model{
		cfg:      cfg,
		repoRoot: root,
		branch:   branch,
		screen:   screenHome,
		home:     screens.NewHomeScreen(cfg, root, branch),
		create:   screens.NewCreateScreen(cfg, root),
	}
}

func (m model) Init() tea.Cmd {
	return m.home.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case screens.SwitchToCreateMsg:
		m.screen = screenCreate
		m.create = screens.NewCreateScreen(m.cfg, m.repoRoot)
		return m, m.create.Init()

	case screens.SwitchToHomeMsg:
		m.screen = screenHome
		if msg.Err != nil {
			return m, tea.Sequence(
				screens.LoadWorktrees(m.repoRoot, m.cfg.WorktreeDir),
				func() tea.Msg { return screens.HomeErrorMsg{Err: msg.Err} },
			)
		}
		return m, screens.LoadWorktrees(m.repoRoot, m.cfg.WorktreeDir)

	case screens.LaunchAfterCreateMsg:
		m.screen = screenHome
		return m, screens.LoadWorktrees(m.repoRoot, m.cfg.WorktreeDir)

	case tea.WindowSizeMsg:
		var cmd1, cmd2 tea.Cmd
		m.home, cmd1 = m.home.Update(msg)
		m.create, cmd2 = m.create.Update(msg)
		return m, tea.Batch(cmd1, cmd2)
	}

	switch m.screen {
	case screenHome:
		var cmd tea.Cmd
		m.home, cmd = m.home.Update(msg)
		return m, cmd
	case screenCreate:
		var cmd tea.Cmd
		m.create, cmd = m.create.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	switch m.screen {
	case screenCreate:
		return m.create.View()
	default:
		return m.home.View()
	}
}
