package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/EPOS-ERIC/epos-opensource/pkg/docker"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"
)

type renderFormData struct {
	envName    string
	outputPath string
}

func (a *App) showRenderForm() {
	a.PushFocus()
	envName, isDocker, _ := a.envList.GetSelected()
	if envName == "" {
		return
	}

	title := "Render Docker Environment"
	if !isDocker {
		title = "Render K8s Environment"
	}

	a.UpdateFooter(RenderFormKey)

	data := &renderFormData{
		envName:    envName,
		outputPath: envName,
	}

	fields := []FormField{
		{
			Type:  "input",
			Label: "Output Path",
			Value: data.outputPath,
			InputChangedFunc: func(text string) {
				data.outputPath = text
			},
		},
	}

	buttons := []FormButton{
		{Label: "Render", SelectedFunc: func() { a.handleRender(data, isDocker) }},
		{Label: "Cancel", SelectedFunc: func() {
			a.ResetToHome(ResetOptions{
				PageNames:    []string{"render-form"},
				RestoreFocus: true,
			})
		}},
	}

	a.ShowModalForm(ModalFormOptions{
		PageName: "render-form",
		Title:    title,
		Fields:   fields,
		Buttons:  buttons,
		OnCancel: func() {
			a.ResetToHome(ResetOptions{
				PageNames:    []string{"render-form"},
				RestoreFocus: true,
			})
		},
	})
}

func (a *App) handleRender(data *renderFormData, isDocker bool) {
	outputPath := strings.TrimSpace(data.outputPath)
	openDir, err := resolveRenderOutputDir(outputPath)
	if err != nil {
		a.ShowError(fmt.Sprintf("Failed to resolve output path: %v", err))
		return
	}

	if isDocker {
		_, err = docker.Render(docker.RenderOpts{
			Name:       data.envName,
			OutputPath: outputPath,
		})
	} else {
		_, err = k8s.Render(k8s.RenderOpts{
			Name:       data.envName,
			OutputPath: outputPath,
		})
	}
	if err != nil {
		a.ShowError(err.Error())
		return
	}

	if err := a.openValue(openDir); err != nil {
		a.ShowError(fmt.Sprintf("Rendered successfully but %v", err))
		return
	}

	a.ResetToHome(ResetOptions{
		PageNames:    []string{"render-form"},
		RestoreFocus: true,
	})

	a.FlashMessage("Rendered files and opened output directory.", 2*time.Second)
}

func resolveRenderOutputDir(outputPath string) (string, error) {
	if outputPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}

		return cwd, nil
	}

	return filepath.Abs(outputPath)
}
