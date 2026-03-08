//go:build !windows

package process

import (
	"os/exec"
	"syscall"
)

func configureProcessForPlatform(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
