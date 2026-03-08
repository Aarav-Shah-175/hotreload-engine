package build

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"runtime"
	"sync"
)

type Runner struct {
	logger *slog.Logger
}

func NewRunner(logger *slog.Logger) *Runner {
	return &Runner{logger: logger.With("component", "build")}
}

func (r *Runner) Run(ctx context.Context, root, command string) error {
	cmd := shellCommand(ctx, command)
	cmd.Dir = root

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("build stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("build stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start build: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go streamLines(&wg, stdout, r.logger, slog.LevelInfo)
	go streamLines(&wg, stderr, r.logger, slog.LevelError)

	waitErr := cmd.Wait()
	wg.Wait()
	if waitErr != nil {
		return fmt.Errorf("build failed: %w", waitErr)
	}
	return nil
}

func shellCommand(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/C", command)
	}
	return exec.CommandContext(ctx, "sh", "-c", command)
}

func streamLines(wg *sync.WaitGroup, reader io.Reader, logger *slog.Logger, level slog.Level) {
	defer wg.Done()
	s := bufio.NewScanner(reader)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for s.Scan() {
		logger.Log(context.Background(), level, s.Text())
	}
}
