//go:build windows
// +build windows

package common

import (
	"bufio"
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
