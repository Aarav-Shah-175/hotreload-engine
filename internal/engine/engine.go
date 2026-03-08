package engine

import (
	"log/slog"

	"hotreload/internal/config"
)

func Run(_ config.Config, logger *slog.Logger) error {
	logger.Info("engine bootstrap complete")
	return nil
}
