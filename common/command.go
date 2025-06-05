//go:build !windows
// +build !windows

package common

import (
	"bufio"
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
		return err
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Read stderr line by line and print in red
	scanner := bufio.NewScanner(stderrPipe)
	for scanner.Scan() {
		PrintError(scanner.Text())
	}

	// Wait for command to finish
	return cmd.Wait()
}
