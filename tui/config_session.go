package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EPOS-ERIC/epos-opensource/common"
)

// configEditSession tracks a temporary editable config file used by deploy/update forms.
type configEditSession struct {
	dir      string
	filePath string
}

func newConfigEditSession(fileName string) (*configEditSession, error) {
	if strings.TrimSpace(fileName) == "" {
		return nil, fmt.Errorf("config filename is required")
	}

	dir, err := os.MkdirTemp("", "epos-tui-config-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp config directory: %w", err)
	}

	return &configEditSession{
		dir:      dir,
		filePath: filepath.Join(dir, fileName),
	}, nil
}

func (s *configEditSession) FilePath() string {
	if s == nil {
		return ""
	}

	return s.filePath
}

func (s *configEditSession) Cleanup() error {
	if s == nil || s.dir == "" {
		return nil
	}

	err := os.RemoveAll(s.dir)
	s.dir = ""
	s.filePath = ""
	if err != nil {
		return fmt.Errorf("failed to remove temp config directory: %w", err)
	}

	return nil
}

func (a *App) openConfigEditor(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("config file path is required")
	}

	editorCmd := strings.TrimSpace(os.Getenv("EDITOR"))
	if editorCmd == "" {
		editorCmd = strings.TrimSpace(a.config.TUI.OpenFileCommand)
	}

	if editorCmd == "" {
		return fmt.Errorf("no editor command configured")
	}

	if !strings.Contains(editorCmd, "%s") {
		editorCmd += " %s"
	}

	var openErr error
	a.tview.Suspend(func() {
		openErr = common.OpenWithCommand(editorCmd, path)
	})
	if openErr != nil {
		return fmt.Errorf("failed to open config editor: %w", openErr)
	}

	return nil
}
