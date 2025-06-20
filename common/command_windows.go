//go:build windows
// +build windows

package common

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func RunCommand(cmd *exec.Cmd, interceptOut bool) (string, error) {
	var stdout bytes.Buffer

	if interceptOut {
		cmd.Stdout = &stdout
	} else {
		cmd.Stdout = os.Stdout
	}
	cmd.Stdin = os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, "COMPOSE_STATUS_STDOUT=1")

	// Create a pipe to capture stderr
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe for %s: %w", cmd.Path, err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start %s: %w", cmd.Path, err)
	}

	// Read stderr line by line and print in red
	if !interceptOut {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			PrintError("%s", scanner.Text())
		}
	}

	if err := cmd.Wait(); err != nil {
		if interceptOut {
			return stdout.String(), fmt.Errorf("%s execution failed: %w", cmd.Path, err)
		}
		return "", fmt.Errorf("%s execution failed: %w", cmd.Path, err)
	}

	if interceptOut {
		return stdout.String(), nil
	}
	return "", nil
}
