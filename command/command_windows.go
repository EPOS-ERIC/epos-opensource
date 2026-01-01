//go:build windows

package command

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/EPOS-ERIC/epos-opensource/display"
)

var (
	Stdout io.Writer = os.Stdout
	Stderr io.Writer = os.Stderr
)

func RunCommand(cmd *exec.Cmd, interceptOut bool) (string, error) {
	var stdout bytes.Buffer

	if interceptOut {
		cmd.Stdout = &stdout
	} else {
		cmd.Stdout = Stdout
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
			display.Error("%s", scanner.Text())
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

// StartCommand starts cmd after applying the same configuration as RunCommand,
// but without waiting for it to exit. It does not redirect stdout or stderr,
// allowing callers to set up pipes as needed.
func StartCommand(cmd *exec.Cmd) error {
	if cmd.Stdin == nil {
		cmd.Stdin = os.Stdin
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, "COMPOSE_STATUS_STDOUT=1")

	return cmd.Start()
}
