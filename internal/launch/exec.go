package launch

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func BuildExecCommand(cmdStr, dir string) (*exec.Cmd, error) {
	parts := strings.Fields(strings.TrimSpace(cmdStr))
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty harness command")
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir
	cmd.Env = BuildEnv()
	return cmd, nil
}

type ExecFinishedMsg struct {
	Err error
}

func BuildEnv() []string {
	return os.Environ()
}
