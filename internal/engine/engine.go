package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"hotreload/internal/build"
	"hotreload/internal/config"
	"hotreload/internal/process"
	"hotreload/internal/watcher"
)

func Run(cfg config.Config, logger *slog.Logger) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fileWatcher, err := watcher.New(cfg.Root)
	if err != nil {
		return fmt.Errorf("watcher setup: %w", err)
	}

	events, watchErrs := fileWatcher.Start(ctx)
	builder := build.NewRunner(logger)
	proc := process.NewManager(logger)
	debouncer := NewDebouncer(ctx, cfg.Debounce)

	queueTrigger := func(reason string) {
		logger.Info("change queued", "reason", reason)
		debouncer.Notify()
	}

	queueTrigger("startup")

	var activeMu sync.Mutex
	var activeCancel context.CancelFunc
	startCycle := func() {
		activeMu.Lock()
		if activeCancel != nil {
			activeCancel()
		}

		cycleCtx, cycleCancel := context.WithCancel(ctx)
		activeCancel = cycleCancel
		activeMu.Unlock()

		go func() {
			logger.Info("reload started")

			stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
			stopErr := proc.Stop(stopCtx)
			if stopErr != nil && !errors.Is(stopErr, context.Canceled) {
				logger.Warn("stop previous server", "error", stopErr)
			}
			stopCancel()

			if err := builder.Run(cycleCtx, cfg.Root, cfg.BuildCmd); err != nil {
				if cycleCtx.Err() != nil {
					logger.Info("build canceled due to newer changes")
					return
				}
				logger.Error("build failed", "error", err)
				return
			}

			if cycleCtx.Err() != nil {
				logger.Info("skipping start because reload was superseded")
				return
			}

			if stopErr == nil {
				time.Sleep(500 * time.Millisecond)
			}

			if err := proc.Start(cfg.Root, cfg.ExecCmd); err != nil {
				logger.Error("exec failed", "error", err)
				return
			}
			logger.Info("reload complete")
		}()
	}

	for {
		select {
		case <-ctx.Done():
			activeMu.Lock()
			if activeCancel != nil {
				activeCancel()
			}
			activeMu.Unlock()

			stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = proc.Stop(stopCtx)
			stopCancel()
			return nil

		case ev, ok := <-events:
			if !ok {
				return nil
			}
			logger.Info("file event", "path", ev.Path, "op", ev.Op.String())
			queueTrigger("file event")

		case err, ok := <-watchErrs:
			if !ok {
				watchErrs = nil
				continue
			}
			logger.Warn("watcher error", "error", err)

		case <-debouncer.C():
			startCycle()
		}
	}
}
