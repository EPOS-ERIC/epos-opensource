//go:build windows

package command

import (
	"os/exec"
	"syscall"
)

// configurePlatformCmd applies Windows specific configuration to a command.
// On Windows, we hide the console window to prevent popups.
func configurePlatformCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
