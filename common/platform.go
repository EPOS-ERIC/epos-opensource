package common

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

const (
	darwinOS  = "darwin"
	linuxOS   = "linux"
	windowsOS = "windows"
)

func OpenBrowser(url string) error {
	url = strings.Trim(url, " ")
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case darwinOS:
		cmd = exec.Command("open", url)
	case linuxOS:
		cmd = exec.Command("xdg-open", url)
	case windowsOS:
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url) // Fallback
	}
	if err := cmd.Run(); err != nil {
		log.Printf("error opening in browser: %v", err)
		return err
	}
	return nil
}

func OpenDirectory(dir string) error {
	dir = strings.Trim(dir, " ")
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case darwinOS:
		cmd = exec.Command("open", dir)
	case linuxOS:
		cmd = exec.Command("xdg-open", dir)
	case windowsOS:
		cmd = exec.Command("cmd", "/c", "start", dir)
	default:
		cmd = exec.Command("xdg-open", dir) // Fallback
	}
	if err := cmd.Run(); err != nil {
		log.Printf("error opening directory: %v", err)
		return err
	}
	return nil
}

func CopyToClipboard(text string) error {
	text = strings.Trim(text, " ")
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case darwinOS:
		cmd = exec.Command("pbcopy")
	case linuxOS:
		// Try different common clipboard commands
		clipboardCmds := []struct {
			name string
			args []string
		}{
			{"xclip", []string{"-selection", "clipboard"}},
			{"wl-copy", []string{}},
			{"xsel", []string{"-b"}},
		}
		for _, c := range clipboardCmds {
			if _, err := exec.LookPath(c.name); err != nil {
				continue
			}
			cmd = exec.Command(c.name, c.args...)
			break
		}
		if cmd == nil {
			return fmt.Errorf("no clipboard command available")
		}
	case windowsOS:
		cmd = exec.Command("clip")
	default:
		cmd = exec.Command("xclip", "-selection", "clipboard") // Fallback
	}
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		log.Printf("error copying to clipboard: %v", err)
		return err
	}
	return nil
}

func OpenWithCommand(command, target string) error {
	target = strings.Trim(target, " ")
	if command == "" {
		// Fallback to default based on type
		if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
			return OpenBrowser(target)
		}
		return OpenDirectory(target)
	}

	// Build args, handling %s replacement or simple append
	var args []string
	if strings.Contains(command, "%s") {
		// Replace %s with target, preserving other parts
		parts := strings.FieldsSeq(command)
		for part := range parts {
			if part == "%s" {
				args = append(args, target)
			} else {
				args = append(args, part)
			}
		}
	} else {
		// Simple case: command is the executable, target is the argument
		args = []string{command, target}
	}

	if len(args) == 0 {
		return fmt.Errorf("invalid command")
	}

	cmd := exec.Command(args[0], args[1:]...)

	// Use PTY for interactive commands on non-Windows systems
	if runtime.GOOS == darwinOS || runtime.GOOS == linuxOS {
		return runInPty(cmd)
	}

	// Fallback for Windows and other OSes
	if err := cmd.Run(); err != nil {
		log.Printf("error executing custom open command: %v", err)
		return err
	}
	return nil
}

// runInPty executes a command in a pseudo-terminal (PTY) to handle nested TUIs.
func runInPty(c *exec.Cmd) error {
	// Start the command in a PTY.
	ptmx, err := pty.Start(c)
	if err != nil {
		return fmt.Errorf("failed to start in pty: %w", err)
	}
	// Make sure to close the PTY at the end.
	defer func() { _ = ptmx.Close() }()

	// Handle window size changes.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.

	// Set stdin in raw mode.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	// Copy stdin to the PTY and the PTY to stdout.
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	_, _ = io.Copy(os.Stdout, ptmx)

	// Wait for the command to exit.
	return c.Wait()
}
