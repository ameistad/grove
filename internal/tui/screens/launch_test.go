package screens

import (
	"errors"
	"testing"

	"github.com/ameistad/grove/internal/config"
	"github.com/ameistad/grove/internal/launch"
)

func TestCreateScreenPassesLaunchErrorBackHome(t *testing.T) {
	t.Parallel()

	c := NewCreateScreen(config.Config{}, t.TempDir())
	c.launching = true

	updated, cmd := c.Update(launch.ExecFinishedMsg{Err: errors.New("launch failed")})
	if updated.launching {
		t.Fatal("expected launching to be cleared")
	}
	if cmd == nil {
		t.Fatal("expected a command that switches back home")
	}

	_, ok := cmd().(SwitchToHomeMsg)
	if !ok {
		t.Fatalf("expected SwitchToHomeMsg, got %T", cmd())
	}
}

func TestHomeScreenReloadsAfterLaunchError(t *testing.T) {
	t.Parallel()

	h := NewHomeScreen(config.Config{WorktreeDir: ".worktrees"}, t.TempDir(), "main")
	h.launching = true

	updated, cmd := h.Update(launch.ExecFinishedMsg{Err: errors.New("launch failed")})
	if updated.launching {
		t.Fatal("expected launching to be cleared")
	}
	if cmd == nil {
		t.Fatal("expected LoadWorktrees command after harness exit")
	}
}
