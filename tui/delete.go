package tui

import (
	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/cmd/k8s/k8score"
)

// showDeleteConfirm displays a confirmation dialog for deleting a Docker or K8s environment.
func (a *App) showDeleteConfirm() {
	envName, isDocker := a.envList.GetSelected()

	if envName == "" {
		return
	}

	message := "This will permanently remove all containers, volumes, and associated resources.\n\n"
	if !isDocker {
		message = "This will permanently remove the namespace and all associated K8s resources.\n\n"
	}
	message += DefaultTheme.DestructiveTag("b") + "This action cannot be undone." + "[-]"

	a.UpdateFooter(DeleteConfirmKey)

	a.ShowConfirmation(ConfirmationOptions{
		PageName:           "delete-confirm",
		Title:              " [::b]Delete Environment ",
		Message:            message,
		ConfirmLabel:       "Delete",
		CancelLabel:        "Cancel",
		Destructive:        true,
		ConfirmDestructive: true,
		OnConfirm: func() {
			a.showDeleteProgress(envName, isDocker)
		},
		OnCancel: func() {
			a.ResetToHome(ResetOptions{
				PageNames:    []string{"delete-confirm"},
				RestoreFocus: true,
			})
		},
	})
}

// showDeleteProgress displays the deletion progress with live output.
func (a *App) showDeleteProgress(envName string, isDocker bool) {
	a.RunBackgroundTask(TaskOptions{
		Operation: "Delete",
		EnvName:   envName,
		IsDocker:  isDocker,
		Task: func() (string, error) {
			if isDocker {
				return "", dockercore.Delete(dockercore.DeleteOpts{
					Name: []string{envName},
				})
			}
			return "", k8score.Delete(k8score.DeleteOpts{
				Name: []string{envName},
			})
		},
	})
}
