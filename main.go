package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/ameistad/grove/internal/config"
	"github.com/ameistad/grove/internal/git"
	"github.com/ameistad/grove/internal/launch"
	"github.com/ameistad/grove/internal/tui"
)

var version = "dev"

func main() {
	args := os.Args[1:]

	if len(args) > 0 && (args[0] == "--version" || args[0] == "-v") {
		fmt.Printf("grove %s\n", version)
		return
	}

	if len(args) > 0 && args[0] == "init" {
		cmdInit()
		return
	}

	if err := git.CheckGit(); err != nil {
		fatal(err)
	}

	cfg, created, err := config.Load()
	if err != nil {
		fatal(err)
	}
	if created {
		fmt.Fprintf(os.Stderr, "Created default config at ~/.config/grove/config.yaml\n")
	}

	if len(args) == 0 {
		runTUI(cfg)
		return
	}

	switch args[0] {
	case "ls":
		cmdLs(cfg)
	case "new":
		newArgs := args[1:]
		var harnessName, cmdOverride string
		var dangerous bool
		for len(newArgs) > 0 {
			switch {
			case newArgs[0] == "--harness" && len(newArgs) > 1:
				harnessName = newArgs[1]
				newArgs = newArgs[2:]
			case newArgs[0] == "--cmd" && len(newArgs) > 1:
				cmdOverride = newArgs[1]
				newArgs = newArgs[2:]
			case newArgs[0] == "--dangerous":
				dangerous = true
				newArgs = newArgs[1:]
			default:
				goto doneNewFlags
			}
		}
	doneNewFlags:
		if len(newArgs) < 1 {
			fatal(fmt.Errorf("usage: grove new [--harness <name>] [--cmd <command>] [--dangerous] <slug>"))
		}
		cmdNew(cfg, newArgs[0], harnessName, cmdOverride, dangerous)
	case "rm":
		if len(args) < 2 {
			fatal(fmt.Errorf("usage: grove rm <slug>"))
		}
		force := false
		slug := args[1]
		if slug == "--force" {
			force = true
			if len(args) < 3 {
				fatal(fmt.Errorf("usage: grove rm [--force] <slug>"))
			}
			slug = args[2]
		}
		cmdRm(cfg, slug, force)
	case "merge":
		if len(args) < 2 {
			fatal(fmt.Errorf("usage: grove merge <slug>"))
		}
		cmdMerge(cfg, args[1])
	case "--help", "-h", "help":
		printUsage()
	default:
		fatal(fmt.Errorf("unknown command: %s\nusage: grove [init|ls|new|rm|merge]", args[0]))
	}
}

func printUsage() {
	fmt.Println("Usage: grove [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init          Detect harnesses and write config")
	fmt.Println("  ls            List active worktrees")
	fmt.Println("  new <slug>    Create a new worktree and launch harness")
	fmt.Println("                  --harness <name>  Use a specific harness")
	fmt.Println("                  --cmd <command>    Override the harness command")
	fmt.Println("                  --dangerous        Enable dangerous mode")
	fmt.Println("  rm <slug>     Remove a worktree (--force to skip dirty check)")
	fmt.Println("  merge <slug>  Merge a worktree back to the current branch")
	fmt.Println()
	fmt.Println("Run without arguments to launch the interactive TUI.")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --version, -v    Print version")
	fmt.Println("  --help, -h       Show this help")
}

var validSlug = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]*$`)

func validateSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("slug cannot be empty")
	}
	if len(slug) > 128 {
		return fmt.Errorf("slug too long (max 128 characters)")
	}
	if !validSlug.MatchString(slug) {
		return fmt.Errorf("invalid slug %q: must start with a letter or digit and contain only letters, digits, '.', '_', '/' or '-'", slug)
	}
	if strings.Contains(slug, "..") {
		return fmt.Errorf("invalid slug %q: must not contain '..'", slug)
	}
	return nil
}

func runTUI(cfg config.Config) {
	root, err := git.RepoRoot()
	if err != nil {
		fatal(fmt.Errorf("not in a git repo"))
	}
	branch, _ := git.BranchName(root)

	if err := tui.Run(cfg, root, branch); err != nil {
		fatal(err)
	}
}

func cmdInit() {
	detected := config.DetectHarnesses()

	existing, _, err := config.Load()
	if err != nil {
		existing = config.Config{}
	}

	if existing.WorktreeDir == "" {
		existing.WorktreeDir = ".worktrees"
	}

	overrides := make(map[string]string)
	for _, h := range existing.Harnesses {
		overrides[h.Name] = h.Cmd
	}

	var harnesses []config.Harness
	for _, h := range detected {
		if cmd, ok := overrides[h.Name]; ok {
			h.Cmd = cmd
		}
		harnesses = append(harnesses, h)
	}

	for _, h := range existing.Harnesses {
		found := false
		for _, d := range detected {
			if d.Name == h.Name {
				found = true
				break
			}
		}
		if !found {
			harnesses = append(harnesses, h)
		}
	}

	existing.Harnesses = harnesses

	if existing.DefaultHarness == "" {
		existing.DefaultHarness = "claude"
	}

	if err := config.Write(existing); err != nil {
		fatal(fmt.Errorf("writing config: %w", err))
	}

	if len(detected) == 0 {
		fmt.Println("No known harnesses found in PATH.")
		fmt.Println("Install one of: claude, codex, opencode, antigravity, amp, aider, goose")
	} else {
		fmt.Println("Detected harnesses:")
		for _, h := range detected {
			fmt.Printf("  %s\n", h.Name)
		}
	}
	fmt.Println("Config written to ~/.config/grove/config.yaml")
}

func cmdLs(cfg config.Config) {
	root, err := git.RepoRoot()
	if err != nil {
		fatal(fmt.Errorf("not in a git repo"))
	}

	wts, err := git.List(root, cfg.WorktreeDir)
	if err != nil {
		fatal(err)
	}

	if len(wts) == 0 {
		fmt.Println("no worktrees")
		return
	}

	for _, wt := range wts {
		dirty := ""
		if wt.Dirty {
			dirty = " *"
		}
		if wt.Branch != wt.Slug {
			fmt.Printf("%-16s %-14s %s%s\n", wt.Slug, wt.Harness, wt.Branch, dirty)
		} else {
			fmt.Printf("%-16s %-14s%s\n", wt.Slug, wt.Harness, dirty)
		}
	}
}

func cmdNew(cfg config.Config, slug, harnessName, cmdOverride string, dangerous bool) {
	if err := validateSlug(slug); err != nil {
		fatal(err)
	}

	root, err := git.RepoRoot()
	if err != nil {
		fatal(fmt.Errorf("not in a git repo"))
	}

	var harness config.Harness
	if harnessName != "" {
		h, ok := cfg.HarnessByName(harnessName)
		if !ok {
			fatal(fmt.Errorf("harness %q not found in config", harnessName))
		}
		harness = h
	} else {
		h, ok := cfg.DefaultHarnessConfig()
		if !ok {
			fatal(fmt.Errorf("default harness %q not found in config", cfg.DefaultHarness))
		}
		harness = h
	}

	if cmdOverride != "" {
		harness.Cmd = cmdOverride
	}

	path, err := git.Create(root, cfg.WorktreeDir, slug, harness.Name)
	if err != nil {
		fatal(err)
	}

	fmt.Printf("Created worktree %s at %s\n", slug, path)

	cmdStr := harness.CmdWithArgs(dangerous)
	parts := strings.Fields(cmdStr)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = path
	cmd.Env = launch.BuildEnv()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

func cmdRm(cfg config.Config, slug string, force bool) {
	root, err := git.RepoRoot()
	if err != nil {
		fatal(fmt.Errorf("not in a git repo"))
	}

	if err := git.Remove(root, cfg.WorktreeDir, slug, force); err != nil {
		fatal(err)
	}
	fmt.Printf("Removed worktree %s\n", slug)
}

func cmdMerge(cfg config.Config, slug string) {
	root, err := git.RepoRoot()
	if err != nil {
		fatal(fmt.Errorf("not in a git repo"))
	}

	result, err := git.Merge(root, cfg.WorktreeDir, slug)
	if err != nil {
		fatal(err)
	}

	if result.Success {
		fmt.Printf("Merged and removed worktree %s\n", slug)
		return
	}

	fmt.Println()
	fmt.Println("Copy this prompt into your agent to resolve:")
	fmt.Println("---")
	fmt.Println(result.AgentPrompt)
	fmt.Println("---")
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "grove: %s\n", err)
	os.Exit(1)
}
