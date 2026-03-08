package process

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"sync"
	"time"
)

type Manager struct {
	logger *slog.Logger

	mu         sync.Mutex
	cmd        *exec.Cmd
	done       chan error
	cancelProc context.CancelFunc
}

func NewManager(logger *slog.Logger) *Manager {
	return &Manager{logger: logger.With("component", "process")}
}

func (m *Manager) Start(root, command string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cmd != nil {
		return errors.New("server already running")
	}

	args, err := splitCommandLine(command)
	if err != nil {
		return fmt.Errorf("parse exec command: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = root
	configureProcessForPlatform(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("exec stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("exec stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("start server: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		err := cmd.Wait()
		m.mu.Lock()
		if m.cmd == cmd {
			m.cmd = nil
			m.done = nil
			m.cancelProc = nil
		}
		m.mu.Unlock()
		done <- err
	}()
	go stream(stdout, m.logger.With("stream", "stdout"), slog.LevelInfo)
	go stream(stderr, m.logger.With("stream", "stderr"), slog.LevelError)

	m.cmd = cmd
	m.done = done
	m.cancelProc = cancel
	return nil
}

func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	cmd := m.cmd
	done := m.done
	cancel := m.cancelProc
	m.cmd = nil
	m.done = nil
	m.cancelProc = nil
	m.mu.Unlock()

	if cmd == nil {
		return nil
	}
	if cmd.Process == nil {
		return nil
	}

	pid := cmd.Process.Pid
	_ = terminateTree(pid, false)
	if cancel != nil {
		cancel()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	case <-time.After(2 * time.Second):
	}

	_ = terminateTree(pid, true)
	if cancel != nil {
		cancel()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func (m *Manager) Wait() <-chan error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.done == nil {
		closed := make(chan error)
		close(closed)
		return closed
	}
	return m.done
}

func stream(reader io.Reader, logger *slog.Logger, level slog.Level) {
	s := bufio.NewScanner(reader)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for s.Scan() {
		logger.Log(context.Background(), level, s.Text())
	}
}
