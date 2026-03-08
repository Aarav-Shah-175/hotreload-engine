package config

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"time"
)

type Config struct {
	Root     string
	BuildCmd string
	ExecCmd  string
	Debounce time.Duration
}

func Parse() (Config, error) {
	cfg := Config{}
	flag.StringVar(&cfg.Root, "root", ".", "project folder to watch recursively")
	flag.StringVar(&cfg.BuildCmd, "build", "", "build command")
	flag.StringVar(&cfg.ExecCmd, "exec", "", "run command")
	flag.DurationVar(&cfg.Debounce, "debounce", 300*time.Millisecond, "file event debounce window")
	flag.Parse()

	if cfg.BuildCmd == "" || cfg.ExecCmd == "" {
		return Config{}, errors.New("both --build and --exec are required")
	}

	root, err := filepath.Abs(cfg.Root)
	if err != nil {
		return Config{}, fmt.Errorf("resolve root path: %w", err)
	}
	cfg.Root = root

	if cfg.Debounce <= 0 {
		return Config{}, errors.New("--debounce must be > 0")
	}

	return cfg, nil
}
