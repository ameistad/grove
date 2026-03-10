package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func CheckGit() error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not installed or not in PATH")
	}
	out, err := run("", "version")
	if err != nil {
		return fmt.Errorf("could not determine git version: %w", err)
	}
	parts := strings.Fields(out)
	if len(parts) < 3 {
		return fmt.Errorf("unexpected git version output: %s", out)
	}
	ver := parts[2]
	segments := strings.SplitN(ver, ".", 4)
	if len(segments) < 2 {
		return nil
	}
	major, _ := strconv.Atoi(segments[0])
	minor, _ := strconv.Atoi(segments[1])
	if major < 2 || (major == 2 && minor < 20) {
		return fmt.Errorf("git %s is too old, grove requires git 2.20 or later", ver)
	}
	return nil
}

func RepoRoot() (string, error) {
	out, err := run("", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("not in a git repo")
	}
	return out, nil
}

func BranchName(dir string) (string, error) {
	out, err := run(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("could not determine branch: %w", err)
	}
	if out == "HEAD" {
		return "", fmt.Errorf("HEAD is detached (not on a branch)")
	}
	return out, nil
}

func WorktreeBase(root string, wtDir string) string {
	return filepath.Join(root, wtDir)
}

func IsDirty(dir string) bool {
	out, err := run(dir, "status", "--porcelain")
	if err != nil {
		return false
	}
	return out != ""
}

func BranchExists(slug string) bool {
	_, err := run("", "show-ref", "--verify", "--quiet", "refs/heads/"+slug)
	return err == nil
}

func DefaultBranch(root string) string {
	out, err := run(root, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		parts := strings.Split(strings.TrimSpace(out), "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	for _, name := range []string{"main", "master"} {
		if BranchExists(name) {
			return name
		}
	}
	return "main"
}
