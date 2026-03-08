package main

import (
	"log/slog"
	"os"

	"hotreload/internal/config"
	"hotreload/internal/engine"
	"hotreload/internal/logging"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(2)
	}

	logger := logging.New()
	if err := engine.Run(cfg, logger); err != nil {
		logger.Error("hotreload failed", "error", err)
		os.Exit(1)
	}
}
