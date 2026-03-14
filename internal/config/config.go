package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed harnesses.json
var harnessesJSON []byte

type Harness struct {
	Name          string `yaml:"name" json:"name"`
	Cmd           string `yaml:"cmd" json:"cmd"`
	DangerousArgs string `yaml:"dangerous_args,omitempty" json:"dangerous_args,omitempty"`
}

func (h Harness) CmdWithArgs(dangerous bool) string {
	if dangerous && h.DangerousArgs != "" {
		return h.Cmd + " " + h.DangerousArgs
	}
	return h.Cmd
}

type Config struct {
	DefaultHarness string    `yaml:"default_harness"`
	WorktreeDir    string    `yaml:"worktree_dir"`
	Harnesses      []Harness `yaml:"harnesses"`
}

func knownHarnesses() []Harness {
	var harnesses []Harness
	if err := json.Unmarshal(harnessesJSON, &harnesses); err != nil {
		panic(fmt.Sprintf("parse embedded harnesses: %v", err))
	}
	return harnesses
}

func DetectHarnesses() []Harness {
	var found []Harness
	for _, h := range knownHarnesses() {
		bin := strings.Fields(h.Cmd)[0]
		if _, err := exec.LookPath(bin); err == nil {
			found = append(found, h)
		}
	}
	return found
}

func defaultConfig() Config {
	return Config{
		DefaultHarness: "claude",
		WorktreeDir:    ".worktrees",
		Harnesses:      DetectHarnesses(),
	}
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "grove", "config.yaml")
}

func Load() (Config, bool, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := defaultConfig()
			if err := Write(cfg); err != nil {
				return cfg, true, fmt.Errorf("creating default config: %w", err)
			}
			return cfg, true, nil
		}
		return Config{}, false, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, false, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, false, nil
}

func Write(cfg Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (c Config) HarnessByName(name string) (Harness, bool) {
	for _, h := range c.Harnesses {
		if h.Name == name {
			return h, true
		}
	}
	return Harness{}, false
}

func (c Config) DefaultHarnessConfig() (Harness, bool) {
	return c.HarnessByName(c.DefaultHarness)
}
