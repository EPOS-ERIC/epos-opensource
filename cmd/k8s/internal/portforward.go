package internal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"time"

	"github.com/epos-eu/epos-opensource/common"
)

// ForwardAndRun spins up kubectl port-forward, runs fn, then cleans up.
func ForwardAndRun(namespace, deployment string, localPort, remotePort int, fn func(host string, port int) error) error {
	args := []string{
		"port-forward",
		"deployment/" + deployment,
		fmt.Sprintf("%d:%d", localPort, remotePort),
		"-n",
		namespace,
	}

	cmd := exec.Command("kubectl", args...)

	// collect stdout & stderr so we can detect readiness
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := common.StartCommand(cmd); err != nil {
		return fmt.Errorf("starting kubectl: %w", err)
	}
	defer stopProcess(cmd.Process)

	if err := waitForForward(stdout, stderr, 10*time.Second); err != nil {
		return err
	}

	if err := fn("127.0.0.1", localPort); err != nil {
		return err
	}

	return nil
}

var fwdRE = regexp.MustCompile(`Forwarding\s+from\s+\S+:(\d+)`)

func waitForForward(stdout, stderr io.Reader, timeout time.Duration) error {
	combined := io.MultiReader(stdout, stderr)
	scan := bufio.NewScanner(combined)

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return errors.New("timeout waiting for kubectl port-forward")
		default:
			if scan.Scan() {
				if fwdRE.Match(scan.Bytes()) {
					return nil
				}
			} else {
				return errors.New("kubectl exited before forwarding was ready")
			}
		}
	}
}

// stopProcess sends SIGINT on *nix, otherwise Kill; waits to avoid a zombie.
func stopProcess(p *os.Process) {
	if p == nil {
		return
	}
	if runtime.GOOS != "windows" {
		_ = p.Signal(os.Interrupt)
	} else {
		_ = p.Kill()
	}
	_, _ = p.Wait()
}
