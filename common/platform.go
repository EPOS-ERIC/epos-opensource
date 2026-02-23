package common

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
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
	if err := runWithTerminal(cmd); err != nil {
		log.Printf("error executing custom open command: %v", err)
		return err
	}

	return nil
}
