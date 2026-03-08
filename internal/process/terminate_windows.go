//go:build windows

package process

import (
	"fmt"
	"os/exec"
)

func terminateTree(pid int, force bool) error {
	args := []string{"/PID", fmt.Sprintf("%d", pid), "/T"}
	if force {
		args = append(args, "/F")
	}
	return exec.Command("taskkill", args...).Run()
}
