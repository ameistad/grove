package screens

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ameistad/grove/internal/config"
	"github.com/ameistad/grove/internal/git"
	"github.com/ameistad/grove/internal/launch"
	"github.com/ameistad/grove/internal/tui/theme"
)

type HomeScreen struct {
	worktrees  []git.WorktreeInfo
	cursor     int
	repoRoot   string
	repoName   string
	branch     string
	cfg        config.Config
	err        error
	confirming bool
	merging    bool
	mergeMsg   string
	launching  bool
	width      int
}

func NewHomeScreen(cfg config.Config, root, branch string) HomeScreen {
	repoName := root
	if idx := strings.LastIndex(root, "/"); idx >= 0 {
		repoName = root[idx+1:]
	}

	return HomeScreen{
		cfg:      cfg,
		repoRoot: root,
		repoName: repoName,
		branch:   branch,
	}
}

type WorktreesLoadedMsg struct {
	Worktrees []git.WorktreeInfo
	Err       error
}

func LoadWorktrees(root string, wtDir string) tea.Cmd {
	return func() tea.Msg {
		wts, err := git.List(root, wtDir)
		return WorktreesLoadedMsg{Worktrees: wts, Err: err}
	}
}

func (h HomeScreen) Init() tea.Cmd {
	return LoadWorktrees(h.repoRoot, h.cfg.WorktreeDir)
}

func (h HomeScreen) Update(msg tea.Msg) (HomeScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width

	case WorktreesLoadedMsg:
		h.worktrees = msg.Worktrees
		h.err = msg.Err
		if h.cursor >= len(h.worktrees) && len(h.worktrees) > 0 {
			h.cursor = len(h.worktrees) - 1
		}

	case launch.ExecFinishedMsg:
		h.launching = false
		return h, LoadWorktrees(h.repoRoot, h.cfg.WorktreeDir)

	case tea.KeyMsg:
		if h.launching {
			return h, nil
		}

		if h.confirming {
			return h.handleConfirm(msg)
		}

		if h.merging {
			h.merging = false
			h.mergeMsg = ""
			return h, nil
		}

		switch {
		case key.Matches(msg, theme.Keys.Up):
			if h.cursor > 0 {
				h.cursor--
			}
		case key.Matches(msg, theme.Keys.Down):
			if h.cursor < len(h.worktrees)-1 {
				h.cursor++
			}
		case key.Matches(msg, theme.Keys.New):
			return h, func() tea.Msg { return SwitchToCreateMsg{} }
		case key.Matches(msg, theme.Keys.Launch):
			return h.launchSelected()
		case key.Matches(msg, theme.Keys.Delete):
			return h.startDelete()
		case key.Matches(msg, theme.Keys.Merge):
			return h.startMerge()
		case key.Matches(msg, theme.Keys.Quit):
			return h, tea.Quit
		}
	}
	return h, nil
}

func (h HomeScreen) handleConfirm(msg tea.KeyMsg) (HomeScreen, tea.Cmd) {
	switch {
	case key.Matches(msg, theme.Keys.ForceYes):
		wt := h.worktrees[h.cursor]
		h.confirming = false
		err := git.Remove(h.repoRoot, h.cfg.WorktreeDir, wt.Slug, true)
		if err != nil {
			h.err = err
			return h, nil
		}
		return h, LoadWorktrees(h.repoRoot, h.cfg.WorktreeDir)
	case key.Matches(msg, theme.Keys.Back), key.Matches(msg, theme.Keys.ForceNo):
		h.confirming = false
	}
	return h, nil
}

func (h HomeScreen) launchSelected() (HomeScreen, tea.Cmd) {
	if len(h.worktrees) == 0 {
		return h, nil
	}
	wt := h.worktrees[h.cursor]
	harness, ok := h.cfg.HarnessByName(wt.Harness)
	if !ok {
		harness, _ = h.cfg.DefaultHarnessConfig()
	}
	h.launching = true
	return h, tea.Exec(buildCmd(harness.Cmd, wt.Path), func(err error) tea.Msg {
		return launch.ExecFinishedMsg{Err: err}
	})
}

func buildCmd(cmdStr, dir string) *execCmd {
	return &execCmd{cmdStr: cmdStr, dir: dir}
}

type execCmd struct {
	cmdStr string
	dir    string
}

func (e *execCmd) Run() error {
	cmd, err := launch.BuildExecCommand(e.cmdStr, e.dir)
	if err != nil {
		return err
	}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func (e *execCmd) SetStdin(r io.Reader)  {}
func (e *execCmd) SetStdout(w io.Writer) {}
func (e *execCmd) SetStderr(w io.Writer) {}

func (h HomeScreen) startDelete() (HomeScreen, tea.Cmd) {
	if len(h.worktrees) == 0 {
		return h, nil
	}
	wt := h.worktrees[h.cursor]
	if wt.Dirty {
		h.confirming = true
		return h, nil
	}
	err := git.Remove(h.repoRoot, h.cfg.WorktreeDir, wt.Slug, false)
	if err != nil {
		h.err = err
		return h, nil
	}
	return h, LoadWorktrees(h.repoRoot, h.cfg.WorktreeDir)
}

func (h HomeScreen) startMerge() (HomeScreen, tea.Cmd) {
	if len(h.worktrees) == 0 {
		return h, nil
	}
	wt := h.worktrees[h.cursor]
	result, err := git.Merge(h.repoRoot, h.cfg.WorktreeDir, wt.Slug)
	if err != nil {
		h.err = err
		return h, nil
	}
	if result.Success {
		h.merging = true
		h.mergeMsg = fmt.Sprintf("Merged %s successfully. Worktree removed.", wt.Slug)
		return h, LoadWorktrees(h.repoRoot, h.cfg.WorktreeDir)
	}
	h.merging = true
	h.mergeMsg = result.AgentPrompt
	return h, nil
}

type SwitchToCreateMsg struct{}

type LaunchAfterCreateMsg struct {
	Slug string
}

func helpBar(bindings ...string) string {
	var parts []string
	for i := 0; i+1 < len(bindings); i += 2 {
		k := theme.StyleHelpKey.Render(bindings[i])
		v := theme.StyleHelp.Render(bindings[i+1])
		parts = append(parts, k+" "+v)
	}
	return strings.Join(parts, theme.StyleHelpSep.Render("  "))
}

func (h HomeScreen) View() string {
	if h.launching {
		return ""
	}

	w := h.width
	if w <= 0 {
		w = 60
	}

	var b strings.Builder

	logo := theme.StyleLogo.Render(" grove")
	repoInfo := theme.StyleSubtitle.Render(
		fmt.Sprintf("%s %s %s",
			h.repoName,
			theme.StyleDivider.Render("/"),
			theme.StyleBranch.Render(h.branch),
		),
	)
	pad := max(0, w-lipgloss.Width(logo)-lipgloss.Width(repoInfo)-1)
	b.WriteString(logo + strings.Repeat(" ", pad) + repoInfo + "\n")
	b.WriteString(theme.StyleDivider.Render(strings.Repeat("─", min(w, 72))) + "\n")
	b.WriteString("\n")

	if h.merging {
		b.WriteString(h.mergeMsg + "\n\n")
		b.WriteString(helpBar("any key", "continue"))
		return b.String()
	}

	if h.confirming {
		wt := h.worktrees[h.cursor]
		b.WriteString(theme.StyleWarning.Render("  Worktree '"+wt.Slug+"' has uncommitted changes.") + "\n")
		b.WriteString(theme.StyleWarning.Render("  Remove anyway?") + "\n\n")
		b.WriteString(helpBar("y", "confirm", "esc", "cancel"))
		return b.String()
	}

	if h.err != nil {
		b.WriteString(theme.StyleError.Render("  error: "+h.err.Error()) + "\n\n")
		h.err = nil
	}

	if len(h.worktrees) == 0 {
		b.WriteString(theme.StyleEmptyState.Render("  No worktrees yet.") + "\n")
		b.WriteString(theme.StyleEmptyState.Render("  Press n to create one.") + "\n")
	} else {
		for i, wt := range h.worktrees {
			selected := i == h.cursor

			cursor := "  "
			if selected {
				cursor = theme.StyleCursor.Render("> ")
			}

			nameStyle := theme.StyleNormal
			if selected {
				nameStyle = theme.StyleSelected
			}

			dirty := ""
			if wt.Dirty {
				dirty = theme.StyleDirty.Render(" ~modified")
			}

			name := nameStyle.Render(wt.Slug)
			harness := theme.StyleHarness.Render(wt.Harness)
			branch := theme.StyleBranch.Render(wt.Branch)

			b.WriteString(cursor + name + "  " + harness + "  " + branch + dirty + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpBar("n", "new", "enter", "launch", "d", "remove", "m", "merge", "q", "quit"))
	return b.String()
}
