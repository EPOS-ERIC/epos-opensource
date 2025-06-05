//go:build windows
// +build windows

package common

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func RunCommand(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Create a pipe to capture stderr
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe for %s: %w", cmd.Path, err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", cmd.Path, err)
	}

	// Read stderr line by line and print in red
	scanner := bufio.NewScanner(stderrPipe)
	for scanner.Scan() {
		PrintError("%s", scanner.Text())
	}

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s execution failed: %w", cmd.Path, err)
	}
	return nil
}
