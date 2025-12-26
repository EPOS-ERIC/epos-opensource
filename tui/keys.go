// Package tui provides screen key constants and utilities.
package tui

// ScreenKey represents a key for identifying UI screens.
type ScreenKey string

// FooterText represents a footer text string.
type FooterText string

const (
	DetailsDockerKey ScreenKey = "details-docker"
	DetailsK8sKey    ScreenKey = "details-k8s"
	DockerKey        ScreenKey = "docker"
	K8sKey           ScreenKey = "k8s"
	FilePickerKey    ScreenKey = "file-picker"
	HomeKey          ScreenKey = "home"
	PopulateFormKey  ScreenKey = "populate-form"
	DeleteConfirmKey ScreenKey = "delete-confirm"
	CleanConfirmKey  ScreenKey = "clean-confirm"
	HelpKey          ScreenKey = "help"
	DeployFormKey    ScreenKey = "deploy-form"
)

const (
	DockerFooter     FooterText = "[Docker Environments]"
	K8sFooter        FooterText = "[K8s Environments]"
	DetailsFooter    FooterText = "[Environment Details]"
	FilePickerFooter FooterText = "[File Picker]"
	HomeFooter       FooterText = "[Home]"
	PopulateFooter   FooterText = "[Populate Environment]"
	DeleteFooter     FooterText = "[Delete Environment]"
	CleanFooter      FooterText = "[Clean Environment]"
	HelpFooter       FooterText = "[Help]"
)

// getDetailsKey returns the appropriate details screen key based on environment type.
func getDetailsKey(envType string) ScreenKey {
	switch envType {
	case "docker":
		return DetailsDockerKey
	case "k8s":
		return DetailsK8sKey
	default:
		return ""
	}
}
