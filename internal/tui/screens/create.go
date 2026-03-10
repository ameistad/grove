package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ameistad/grove/internal/config"
	"github.com/ameistad/grove/internal/git"
	"github.com/ameistad/grove/internal/launch"
	"github.com/ameistad/grove/internal/tui/theme"
)

type createStep int

const (
	stepPickHarness createStep = iota
	stepEnterSlug
)

type CreateScreen struct {
	cfg       config.Config
	repoRoot  string
	step      createStep
	cursor    int
	textInput textinput.Model
	err       error
	harness   config.Harness
	launching bool
}

func NewCreateScreen(cfg config.Config, root string) CreateScreen {
	ti := textinput.New()
	ti.Placeholder = "branch-slug"
	ti.CharLimit = 64
	ti.Width = 40

	return CreateScreen{
		cfg:       cfg,
		repoRoot:  root,
		step:      stepPickHarness,
		textInput: ti,
	}
}

func (c CreateScreen) Init() tea.Cmd {
	return nil
}

type SwitchToHomeMsg struct{}

type CreateDoneMsg struct {
	Slug    string
	Path    string
	Harness config.Harness
}

func (c CreateScreen) Update(msg tea.Msg) (CreateScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case launch.ExecFinishedMsg:
		c.launching = false
		return c, func() tea.Msg { return SwitchToHomeMsg{} }

	case tea.KeyMsg:
		if c.launching {
			return c, nil
		}

		if key.Matches(msg, theme.Keys.Back) {
			if c.step == stepEnterSlug {
				c.step = stepPickHarness
				return c, nil
			}
			return c, func() tea.Msg { return SwitchToHomeMsg{} }
		}

		if key.Matches(msg, theme.Keys.Quit) && c.step == stepPickHarness {
			return c, func() tea.Msg { return SwitchToHomeMsg{} }
		}

		switch c.step {
		case stepPickHarness:
			return c.updatePickHarness(msg)
		case stepEnterSlug:
			return c.updateEnterSlug(msg)
		}
	}
	return c, nil
}

func (c CreateScreen) updatePickHarness(msg tea.KeyMsg) (CreateScreen, tea.Cmd) {
	switch {
	case key.Matches(msg, theme.Keys.Up):
		if c.cursor > 0 {
			c.cursor--
		}
	case key.Matches(msg, theme.Keys.Down):
		if c.cursor < len(c.cfg.Harnesses)-1 {
			c.cursor++
		}
	case key.Matches(msg, theme.Keys.Enter):
		c.harness = c.cfg.Harnesses[c.cursor]
		c.step = stepEnterSlug
		c.textInput.Focus()
		return c, c.textInput.Cursor.BlinkCmd()
	}
	return c, nil
}

func (c CreateScreen) updateEnterSlug(msg tea.KeyMsg) (CreateScreen, tea.Cmd) {
	if msg.Type == tea.KeyEnter {
		slug := strings.TrimSpace(c.textInput.Value())
		if slug == "" {
			c.err = fmt.Errorf("slug cannot be empty")
			return c, nil
		}

		path, err := git.Create(c.repoRoot, c.cfg.WorktreeDir, slug, c.harness.Name)
		if err != nil {
			c.err = err
			return c, nil
		}

		c.launching = true
		return c, tea.Exec(buildCmd(c.harness.Cmd, path), func(err error) tea.Msg {
			return launch.ExecFinishedMsg{Err: err}
		})
	}

	var cmd tea.Cmd
	c.textInput, cmd = c.textInput.Update(msg)
	return c, cmd
}

func (c CreateScreen) View() string {
	if c.launching {
		return ""
	}

	var b strings.Builder

	b.WriteString(theme.StyleTitle.Render(" New Worktree") + "\n")
	b.WriteString(theme.StyleDivider.Render(strings.Repeat("─", 40)) + "\n\n")

	if c.err != nil {
		b.WriteString(theme.StyleError.Render("  error: "+c.err.Error()) + "\n\n")
	}

	switch c.step {
	case stepPickHarness:
		b.WriteString(theme.StyleSectionLabel.Render("  Select harness") + "\n\n")

		for i, h := range c.cfg.Harnesses {
			selected := i == c.cursor

			cursor := "  "
			if selected {
				cursor = theme.StyleCursor.Render("> ")
			}

			nameStyle := theme.StyleNormal
			if selected {
				nameStyle = theme.StyleSelected
			}

			b.WriteString(cursor + nameStyle.Render(h.Name) + "  " + theme.StyleHarness.Render(h.Cmd) + "\n")
		}

		b.WriteString("\n")
		b.WriteString(helpBar("enter", "select", "esc", "back"))

	case stepEnterSlug:
		b.WriteString(theme.StyleSectionLabel.Render("  Harness") + "  " + theme.StyleSelected.Render(c.harness.Name) + "\n\n")
		b.WriteString(theme.StyleSectionLabel.Render("  Branch slug") + "\n\n")
		b.WriteString("  " + c.textInput.View() + "\n\n")
		b.WriteString(helpBar("enter", "create", "esc", "back"))
	}

	return b.String()
}
