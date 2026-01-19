//go:build !windows

package command

import "os/exec"

// configurePlatformCmd applies Unix-specific configuration to a command.
// On Unix systems, no special configuration is needed.
func configurePlatformCmd(cmd *exec.Cmd) {
	// No special configuration needed for Unix
}
