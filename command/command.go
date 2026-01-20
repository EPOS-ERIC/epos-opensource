// Package command provides utilities for running external commands with proper
// stdout/stderr handling across different platforms.
package command

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/EPOS-ERIC/epos-opensource/display"
)

var Stdout io.Writer = os.Stdout

// RunCommand executes a command and handles its output.
// If interceptOut is true, stdout is captured and returned as a string.
// If interceptOut is false, stdout is passed through to Stdout.
// Stderr is handled specially for docker commands (printed as normal output)
// and for other commands (shown as warnings on success, errors on failure).
func RunCommand(cmd *exec.Cmd, interceptOut bool) (string, error) {
	var stdout bytes.Buffer
	var stderrLines []string

	isDockerCmd := filepath.Base(cmd.Path) == "docker"

	if interceptOut {
		cmd.Stdout = &stdout
	} else {
		cmd.Stdout = Stdout
	}
	cmd.Stdin = os.Stdin

	if cmd.Env == nil {
		cmd.Env = os.Environ()
	} else {
		cmd.Env = append(cmd.Env, os.Environ()...)
	}

	// Apply platform specific configuration
	configurePlatformCmd(cmd)

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe for %s: %w", cmd.Path, err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start %s: %w", cmd.Path, err)
	}

	// Always drain stderr to prevent the command from blocking if the pipe buffer fills up.
	// For docker commands: print immediately to stdout (docker uses stderr for CLI status).
	// For other commands: collect lines to display after completion.
	scanner := bufio.NewScanner(stderrPipe)
	for scanner.Scan() {
		line := scanner.Text()
		if isDockerCmd && !interceptOut {
			_, err := fmt.Fprintf(Stdout, "%s\n", line)
			if err != nil {
				return "", fmt.Errorf("failed to write to stdout: %w", err)
			}
		} else {
			stderrLines = append(stderrLines, line)
		}
	}
	cmdErr := cmd.Wait()

	// Handle stderr output for non-docker commands
	if !isDockerCmd && len(stderrLines) > 0 {
		if cmdErr != nil {
			for _, line := range stderrLines {
				display.Error("  %s", line)
			}
		} else if !interceptOut {
			for _, line := range stderrLines {
				display.Warn("  %s", line)
			}
		}
	}

	if cmdErr != nil {
		if interceptOut {
			return stdout.String(), fmt.Errorf("%s execution failed: %w", cmd.Path, cmdErr)
		}
		return "", fmt.Errorf("%s execution failed: %w", cmd.Path, cmdErr)
	}

	if interceptOut {
		return stdout.String(), nil
	}
	return "", nil
}

// StartCommand starts cmd after applying platform-specific configuration,
// but without waiting for it to exit. It does not redirect stdout or stderr,
// allowing callers to set up pipes as needed.
func StartCommand(cmd *exec.Cmd) error {
	if cmd.Stdin == nil {
		cmd.Stdin = os.Stdin
	}

	if cmd.Env == nil {
		cmd.Env = os.Environ()
	} else {
		cmd.Env = append(cmd.Env, os.Environ()...)
	}

	// Apply platform-specific configuration
	configurePlatformCmd(cmd)

	return cmd.Start()
}
