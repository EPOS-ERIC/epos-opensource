//go:build !windows
// +build !windows

package common

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

func RunCommand(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "COMPOSE_STATUS_STDOUT=1")

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
