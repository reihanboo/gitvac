package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	Root     string   `json:"root"`
	Targets  []string `json:"targets"`
	Excludes []string `json:"excludes"`
}

func Load() (Config, error) {
	root, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Root: filepath.Join(root, "git"),
		Targets: []string{
			"node_modules", ".venv", "venv", "__pycache__", ".pytest_cache",
			".mypy_cache", "target", "dist", "build", ".next", ".svelte-kit", ".turbo",
		},
	}

	path, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return Config{}, err
	}
	var file Config
	if err := json.Unmarshal(data, &file); err != nil {
		return Config{}, err
	}
	if file.Root != "" {
		cfg.Root = file.Root
	}
	if file.Targets != nil {
		cfg.Targets = file.Targets
	}
	if file.Excludes != nil {
		cfg.Excludes = file.Excludes
	}
	return cfg, nil
}

func ConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gitvac", "config.json"), nil
}
