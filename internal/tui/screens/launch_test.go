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

	want := errors.New("launch failed")
	updated, cmd := c.Update(launch.ExecFinishedMsg{Err: want})
	if updated.launching {
		t.Fatal("expected launching to be cleared")
	}
	if cmd == nil {
		t.Fatal("expected a command that switches back home")
	}

	msg, ok := cmd().(SwitchToHomeMsg)
	if !ok {
		t.Fatalf("expected SwitchToHomeMsg, got %T", cmd())
	}
	if !errors.Is(msg.Err, want) {
		t.Fatalf("expected error %v, got %v", want, msg.Err)
	}
}

func TestHomeScreenStoresLaunchError(t *testing.T) {
	t.Parallel()

	h := NewHomeScreen(config.Config{WorktreeDir: ".worktrees"}, t.TempDir(), "main")
	h.launching = true

	want := errors.New("launch failed")
	updated, cmd := h.Update(launch.ExecFinishedMsg{Err: want})
	if updated.launching {
		t.Fatal("expected launching to be cleared")
	}
	if !errors.Is(updated.err, want) {
		t.Fatalf("expected error %v, got %v", want, updated.err)
	}
	if cmd != nil {
		t.Fatal("expected no follow-up command when launch fails")
	}
}
