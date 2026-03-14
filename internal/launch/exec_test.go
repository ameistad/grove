package launch

import "testing"

func TestBuildExecCommandRejectsEmptyCommand(t *testing.T) {
	t.Parallel()

	if _, err := BuildExecCommand("   ", t.TempDir()); err == nil {
		t.Fatal("expected an error for an empty harness command")
	}
}

func TestBuildExecCommandKeepsForegroundTTY(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cmd, err := BuildExecCommand("echo hello", dir)
	if err != nil {
		t.Fatalf("BuildExecCommand returned error: %v", err)
	}

	if cmd.Dir != dir {
		t.Fatalf("expected dir %q, got %q", dir, cmd.Dir)
	}

	if cmd.SysProcAttr != nil {
		t.Fatalf("expected SysProcAttr to be nil so the child stays in the foreground process group, got %#v", cmd.SysProcAttr)
	}
}
