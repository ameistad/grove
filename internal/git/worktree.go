package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type WorktreeMeta struct {
	Harness string    `yaml:"harness"`
	Created time.Time `yaml:"created"`
}

type WorktreeInfo struct {
	Slug    string
	Path    string
	Branch  string
	Harness string
	Dirty   bool
}

func EnsureGitignore(root, wtDir string) error {
	gitignore := filepath.Join(root, ".gitignore")
	data, err := os.ReadFile(gitignore)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	content := string(data)
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == wtDir || trimmed == wtDir+"/" {
			return nil
		}
	}

	f, err := os.OpenFile(gitignore, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		fmt.Fprintln(f)
	}
	fmt.Fprintln(f, wtDir)
	return nil
}

func Create(root, wtDir, slug, harness string) (string, error) {
	base := WorktreeBase(root, wtDir)

	if err := EnsureGitignore(root, wtDir); err != nil {
		return "", fmt.Errorf("updating .gitignore: %w", err)
	}

	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", fmt.Errorf("creating worktree dir: %w", err)
	}

	run(root, "worktree", "prune")

	path := filepath.Join(base, slug)
	if BranchExists(slug) {
		if _, err := run(root, "worktree", "add", path, slug); err != nil {
			return "", fmt.Errorf("adding worktree (existing branch): %w", err)
		}
	} else {
		if _, err := run(root, "worktree", "add", path); err != nil {
			return "", fmt.Errorf("adding worktree: %w", err)
		}
	}

	if err := writeMeta(base, slug, harness); err != nil {
		return "", fmt.Errorf("writing metadata: %w", err)
	}

	return path, nil
}

func List(root, wtDir string) ([]WorktreeInfo, error) {
	base := WorktreeBase(root, wtDir)

	run(root, "worktree", "prune")

	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var infos []WorktreeInfo
	for _, e := range entries {
		if !e.IsDir() || e.Name() == ".meta" {
			continue
		}
		slug := e.Name()
		path := filepath.Join(base, slug)

		branch, err := BranchName(path)
		if err != nil {
			continue
		}

		meta := readMeta(base, slug)
		infos = append(infos, WorktreeInfo{
			Slug:    slug,
			Path:    path,
			Branch:  branch,
			Harness: meta.Harness,
			Dirty:   IsDirty(path),
		})
	}
	return infos, nil
}

func Remove(root, wtDir, slug string, force bool) error {
	base := WorktreeBase(root, wtDir)
	path := filepath.Join(base, slug)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("worktree not found: %s", slug)
	}

	if !force && IsDirty(path) {
		return fmt.Errorf("worktree %s has uncommitted changes, use force to remove", slug)
	}

	if _, err := run(root, "worktree", "remove", "--force", path); err != nil {
		return fmt.Errorf("removing worktree: %w", err)
	}

	run(root, "worktree", "prune")
	deleteMeta(base, slug)
	return nil
}

func metaDir(base string) string {
	return filepath.Join(base, ".meta")
}

func metaPath(base, slug string) string {
	return filepath.Join(metaDir(base), slug+".yaml")
}

func writeMeta(base, slug, harness string) error {
	dir := metaDir(base)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	meta := WorktreeMeta{
		Harness: harness,
		Created: time.Now(),
	}
	data, err := yaml.Marshal(meta)
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath(base, slug), data, 0o644)
}

func readMeta(base, slug string) WorktreeMeta {
	var meta WorktreeMeta
	data, err := os.ReadFile(metaPath(base, slug))
	if err != nil {
		return meta
	}
	yaml.Unmarshal(data, &meta)
	return meta
}

func deleteMeta(base, slug string) {
	os.Remove(metaPath(base, slug))
}
