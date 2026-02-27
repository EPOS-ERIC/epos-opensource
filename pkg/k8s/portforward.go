package k8s

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/EPOS-ERIC/epos-opensource/command"
	"github.com/EPOS-ERIC/epos-opensource/display"
)

// ForwardAndRun spins up kubectl port-forward, runs fn, then cleans up.
func ForwardAndRun(namespace, deployment string, localPort, remotePort int, context string, fn func(host string, port int) error) error {
	args := []string{
		"port-forward",
		"deployment/" + deployment,
		fmt.Sprintf("%d:%d", localPort, remotePort),
		"-n",
		namespace,
	}
	if context != "" {
		args = append(args, "--context", context)
	}

	cmd := exec.Command("kubectl", args...)
	display.Debug("starting kubectl port-forward: kubectl %s", strings.Join(args, " "))

	// collect stdout & stderr so we can detect readiness
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create kubectl stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create kubectl stderr pipe: %w", err)
	}

	if err := command.StartCommand(cmd); err != nil {
		return fmt.Errorf("starting kubectl: %w", err)
	}
	defer stopProcess(cmd)

	if cmd.Process != nil {
		display.Debug("kubectl port-forward started with pid %d", cmd.Process.Pid)
	}

	ready := make(chan struct{})
	done := make(chan struct{})
	readErr := make(chan error, 1)
	go streamPortForwardOutput(stdout, stderr, ready, done, readErr)

	if err := waitForForward(30*time.Second, ready, done, readErr); err != nil {
		return fmt.Errorf("error waiting for port-forward: %w", err)
	}

	display.Done("Port forward started successfully")

	if err := fn("127.0.0.1", localPort); err != nil {
		return err
	}

	return nil
}

var fwdRE = regexp.MustCompile(`Forwarding\s+from\s+\S+:(\d+)`)

// streamPortForwardOutput drains kubectl port-forward output until process exit.
// It emits readiness once the standard forwarding line appears, reports scanner
// errors (if any), and always closes done when streaming completes.
func streamPortForwardOutput(stdout, stderr io.Reader, ready chan struct{}, done chan struct{}, readErr chan error) {
	defer close(done)

	reader := io.MultiReader(stdout, stderr)

	scan := bufio.NewScanner(reader)
	readyEmitted := false

	for scan.Scan() {
		line := scan.Text()
		display.Debug("kubectl port-forward: %s", line)

		if !readyEmitted && fwdRE.MatchString(line) {
			readyEmitted = true
			close(ready)
		}
	}

	if err := scan.Err(); err != nil {
		select {
		case readErr <- err:
		default:
		}
	}
}

// waitForForward blocks until port-forward is ready, exits early on stream/process
// failure, or returns on timeout.
func waitForForward(timeout time.Duration, ready chan struct{}, done chan struct{}, readErr chan error) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ready:
		return nil
	case err := <-readErr:
		return fmt.Errorf("error reading kubectl port-forward output: %w", err)
	case <-done:
		return fmt.Errorf("kubectl exited before forwarding was ready")
	case <-timer.C:
		return fmt.Errorf("timeout waiting for kubectl port-forward")
	}
}

// stopProcess sends SIGINT on *nix, otherwise Kill; waits to avoid a zombie.
func stopProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	p := cmd.Process

	if cmd.ProcessState != nil {
		display.Debug("kubectl port-forward process already exited: %s", cmd.ProcessState.String())
		return
	}

	display.Debug("stopping kubectl port-forward process")

	if runtime.GOOS != "windows" {
		err := p.Signal(os.Interrupt)
		if err != nil {
			display.Warn("failed to send SIGINT to kubectl: %v", err)
		}
	} else {
		err := p.Kill()
		if err != nil {
			display.Warn("failed to kill kubectl: %v", err)
		}
	}

	err := cmd.Wait()
	if err != nil {
		display.Warn("failed to wait for kubectl to exit: %v", err)
		return
	}

	if cmd.ProcessState != nil {
		display.Debug("kubectl port-forward process stopped: %s", cmd.ProcessState.String())
	}
}
