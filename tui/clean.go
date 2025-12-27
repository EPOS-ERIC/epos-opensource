package tui

import (
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/cmd/docker/dockercore"
	"github.com/EPOS-ERIC/epos-opensource/cmd/k8s/k8score"
)

// showCleanConfirm displays a confirmation dialog for cleaning an environment.
func (a *App) showCleanConfirm() {
	envName, _ := a.envList.GetSelected()
	if envName == "" {
		return
	}

	isDocker := a.envList.IsDockerActive()

	message := "This will permanently delete all data in environment '" + envName + "'.\n\n" + DefaultTheme.DestructiveTag("b") + "This action cannot be undone." + "[-]"

	a.UpdateFooter(CleanConfirmKey)

	a.ShowConfirmation(ConfirmationOptions{
		PageName:           "clean-confirm",
		Title:              " [::b]Clean Environment ",
		Message:            message,
		ConfirmLabel:       "Clean",
		CancelLabel:        "Cancel",
		ConfirmDestructive: true,
		Secondary:          true,
		OnConfirm: func() {
			a.showCleanProgress(envName, isDocker)
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
func (a *App) showCleanProgress(envName string, isDocker bool) {
	a.RunBackgroundTask(TaskOptions{
		Operation: "Clean",
		EnvName:   envName,
		IsDocker:  isDocker,
		Task: func() (string, error) {
			if isDocker {
				docker, err := dockercore.Clean(dockercore.CleanOpts{
					Name: envName,
				})
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("Environment cleaned successfully! GUI: %s", docker.GuiUrl), nil
			} else {
				kube, err := k8score.Clean(k8score.CleanOpts{
					Name: envName,
				})
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("Environment cleaned successfully! GUI: %s", kube.GuiUrl), nil
			}
		},
	})
}
