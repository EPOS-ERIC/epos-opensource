package tui

import (
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"
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
				return "", docker.Delete(docker.DeleteOpts{
					Name: []string{envName},
				})
			}
			return "", k8s.Delete(k8s.DeleteOpts{
				Name: []string{envName},
			})
		},
	})
}
