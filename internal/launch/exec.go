package launch

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func BuildExecCommand(cmdStr, dir string) (*exec.Cmd, error) {
	parts := strings.Fields(cmdStr)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir
	cmd.Env = BuildEnv()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd, nil
}

type ExecFinishedMsg struct {
	Err error
}

func BuildEnv() []string {
	return os.Environ()
}
