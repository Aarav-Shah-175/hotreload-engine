package process

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"runtime"
	"sync"
	"syscall"
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

	ctx, cancel := context.WithCancel(context.Background())
	cmd := shellCommand(ctx, command)
	cmd.Dir = root
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

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
	go func() { done <- cmd.Wait() }()
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

	if cancel != nil {
		cancel()
	}

	pid := cmd.Process.Pid
	_ = terminateTree(pid, false)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return nil
		}
		return nil
	case <-time.After(2 * time.Second):
	}

	_ = terminateTree(pid, true)
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

func shellCommand(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/C", command)
	}
	return exec.CommandContext(ctx, "sh", "-c", command)
}

func terminateTree(pid int, force bool) error {
	if runtime.GOOS == "windows" {
		args := []string{"/PID", fmt.Sprintf("%d", pid), "/T"}
		if force {
			args = append(args, "/F")
		}
		return exec.Command("taskkill", args...).Run()
	}

	sig := syscall.SIGTERM
	if force {
		sig = syscall.SIGKILL
	}
	return syscall.Kill(-pid, sig)
}

func stream(reader io.Reader, logger *slog.Logger, level slog.Level) {
	s := bufio.NewScanner(reader)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for s.Scan() {
		logger.Log(context.Background(), level, s.Text())
	}
}
