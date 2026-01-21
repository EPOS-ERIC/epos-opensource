//go:build windows

package common

import (
	"fmt"
	"os/exec"
)

func runInPty(c *exec.Cmd) error {
	if c == nil {
		return fmt.Errorf("cannot execute nil command")
	}

	// PTY not supported on Windows, fallback to regular execution
	if err := c.Run(); err != nil {
		return fmt.Errorf("fallback execution failed (PTY not supported on Windows): %w", err)
	}
	return nil
}
