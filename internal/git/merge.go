package git

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type MergeResult struct {
	Success       bool
	ConflictFiles []string
	AgentPrompt   string
	Stashed       bool
}

func Merge(root, wtDir, slug string) (*MergeResult, error) {
	base := WorktreeBase(root, wtDir)
	path := filepath.Join(base, slug)

	if _, err := run(path, "rev-parse", "--abbrev-ref", "HEAD"); err != nil {
		return nil, fmt.Errorf("worktree not found: %s", slug)
	}

	branch, err := BranchName(path)
	if err != nil {
		return nil, err
	}

	if IsDirty(path) {
		if _, err := run(path, "add", "-A"); err != nil {
			return nil, fmt.Errorf("staging worktree changes: %w", err)
		}
		if _, err := run(path, "commit", "-m", fmt.Sprintf("grove: auto-commit changes from %s", slug)); err != nil {
			return nil, fmt.Errorf("committing worktree changes: %w", err)
		}
	}

	stashed := false
	if IsDirty(root) {
		if _, err := run(root, "stash", "push", "-m", fmt.Sprintf("grove: auto-stash before merging %s", slug)); err != nil {
			return nil, fmt.Errorf("stashing main worktree: %w", err)
		}
		stashed = true
	}

	mainBranch, _ := BranchName(root)
	_, mergeErr := run(root, "merge", branch)

	if mergeErr == nil {
		var cleanupErrs []error
		if stashed {
			if err := popStash(root); err != nil {
				cleanupErrs = append(cleanupErrs, err)
			}
		}
		if err := cleanupMergedWorktree(root, base, path, branch, slug); err != nil {
			cleanupErrs = append(cleanupErrs, err)
		}
		if err := errors.Join(cleanupErrs...); err != nil {
			return nil, err
		}
		return &MergeResult{Success: true}, nil
	}

	conflictOut, _ := run(root, "diff", "--name-only", "--diff-filter=U")
	if conflictOut == "" {
		err := fmt.Errorf("merge failed: %w", mergeErr)
		if stashed {
			if popErr := popStash(root); popErr != nil {
				return nil, errors.Join(err, popErr)
			}
		}
		return nil, err
	}

	files := strings.Split(conflictOut, "\n")

	var fileList strings.Builder
	for _, f := range files {
		f = strings.TrimSpace(f)
		if f != "" {
			fmt.Fprintf(&fileList, "- %s\n", f)
		}
	}

	prompt := fmt.Sprintf(`Resolve the merge conflicts in this repository. I was merging branch '%s' (worktree '%s') into %s.

Conflicted files:
%s
For each file, look at the conflict markers (<<<<<<< / ======= / >>>>>>>), understand the intent of both sides, and pick the correct resolution. Remove all conflict markers when done.

After resolving, stage the files with git add and run: git commit --no-edit

Then clean up the worktree:
  git worktree remove --force %s
  git worktree prune
  git branch -d %s`, branch, slug, mainBranch, fileList.String(), path, branch)

	if stashed {
		prompt += "\n  git stash pop  # restore stashed changes"
	}

	return &MergeResult{
		Success:       false,
		ConflictFiles: files,
		AgentPrompt:   prompt,
		Stashed:       stashed,
	}, nil
}

func cleanupMergedWorktree(root, base, path, branch, slug string) error {
	var errs []error

	if _, err := run(root, "worktree", "remove", "--force", path); err != nil {
		errs = append(errs, fmt.Errorf("removing merged worktree: %w", err))
	}
	if err := pruneWorktrees(root); err != nil {
		errs = append(errs, err)
	}
	if _, err := run(root, "branch", "-d", branch); err != nil {
		errs = append(errs, fmt.Errorf("deleting merged branch: %w", err))
	}
	if err := deleteMeta(base, slug); err != nil {
		errs = append(errs, fmt.Errorf("removing metadata: %w", err))
	}

	return errors.Join(errs...)
}

func popStash(root string) error {
	if _, err := run(root, "stash", "pop", "--quiet"); err != nil {
		return fmt.Errorf("restoring stashed changes: %w", err)
	}
	return nil
}
