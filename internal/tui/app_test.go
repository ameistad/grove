package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ameistad/grove/internal/config"
)

func TestModelCtrlCQuitsFromAnyScreen(t *testing.T) {
	t.Parallel()

	testCases := []screen{screenHome, screenCreate}
	for _, current := range testCases {
		m := newModel(config.Config{WorktreeDir: ".worktrees"}, t.TempDir(), "main")
		m.screen = current

		next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		if cmd == nil {
			t.Fatalf("expected quit command for screen %v", current)
		}

		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatalf("expected tea.QuitMsg for screen %v", current)
		}

		if _, ok := next.(model); !ok {
			t.Fatalf("expected updated model for screen %v", current)
		}
	}
}
