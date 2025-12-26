package tui

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
)

// showCleanConfirm displays a confirmation dialog for cleaning a Docker environment.
func (a *App) showCleanConfirm() {
	envName, _ := a.envList.GetSelected()
	if envName == "" {
		return
	}

	message := "This will permanently delete all data in environment '" + envName + "'.\n\n" + DefaultTheme.DestructiveTag("b") + "This action cannot be undone." + "[-]"

	a.UpdateFooter("[Clean Environment]", "clean-confirm")

	a.ShowConfirmation(ConfirmationOptions{
		PageName:           "clean-confirm",
		Title:              " [::b]Clean Environment ",
		Message:            message,
		ConfirmLabel:       "Clean",
		CancelLabel:        "Cancel",
		ConfirmDestructive: true,
		Secondary:          true,
		OnConfirm: func() {
			a.showCleanProgress(envName)
		},
		OnCancel: func() {
			a.ResetToHome(ResetOptions{
				PageNames:    []string{"clean-confirm"},
				RestoreFocus: true,
			})
		},
	})
}

// showCleanProgress displays the cleaning progress with live output.
func (a *App) showCleanProgress(envName string) {
	a.RunBackgroundTask(TaskOptions{
		Operation: "Clean",
		EnvName:   envName,
		IsDocker:  true,
		Task: func() (string, error) {
			docker, err := dockercore.Clean(dockercore.CleanOpts{
				Name: envName,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Environment cleaned successfully! GUI: %s", docker.GuiUrl), nil
		},
	})
}
