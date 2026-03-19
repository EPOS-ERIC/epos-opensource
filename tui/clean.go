package tui

import (
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/pkg/docker"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"
)

// showCleanConfirm displays a confirmation dialog for cleaning an environment.
func (a *App) showCleanConfirm() {
	envName, isDocker, k8sContext := a.envList.GetSelected()
	if envName == "" {
		return
	}

	message := "This will permanently delete all data in environment '" + envName + "'.\n\n" + DefaultTheme.DestructiveTag("b") + "This action cannot be undone." + "[-]"
	if !isDocker && k8sContext != "" {
		message = "This will permanently delete all data in environment '" + envName + "' (context: '" + k8sContext + "').\n\n" + DefaultTheme.DestructiveTag("b") + "This action cannot be undone." + "[-]"
	}

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
			a.showCleanProgress(envName, isDocker, k8sContext)
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
func (a *App) showCleanProgress(envName string, isDocker bool, context string) {
	a.RunBackgroundTask(TaskOptions{
		Operation: "Clean",
		EnvName:   envName,
		IsDocker:  isDocker,
		Task: func() (string, error) {
			if isDocker {
				env, err := docker.Clean(docker.CleanOpts{
					Name: envName,
				})
				if err != nil {
					return "", err
				}

				urls, err := env.BuildEnvURLs()
				if err != nil {
					return "", err
				}

				return fmt.Sprintf("Environment cleaned successfully! GUI: %s", urls.GUIURL), nil
			} else {
				env, err := k8s.Clean(k8s.CleanOpts{
					Name:    envName,
					Context: context,
				})
				if err != nil {
					return "", err
				}

				urls, err := env.BuildEnvURLs()
				if err != nil {
					return "", err
				}

				return fmt.Sprintf("Environment cleaned successfully! GUI: %s", urls.GUIURL), nil
			}
		},
	})
}
