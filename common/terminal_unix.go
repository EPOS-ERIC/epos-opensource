//go:build darwin || linux

package common

import (
	"fmt"
	"os"
	"os/exec"
)

func runWithTerminal(c *exec.Cmd) error {
	if c == nil {
		return fmt.Errorf("cannot execute nil command")
	}

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}
