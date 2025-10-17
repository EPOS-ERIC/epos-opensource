package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/google/go-github/v72/github"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "epos-opensource",
	Short: "Manage EPOS environments and utilities.",
	Long: `epos-opensource provides commands for managing local EPOS environments
using Docker Compose or Kubernetes. Use the "docker" and "kubernetes" command
groups to deploy, populate, update, or delete an environment.`,

	// If no subcommand is provided, show help.
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
		}
	},
	Version: common.GetVersion(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// TODO: implement a version & update check using this
func GetLatestGitHubTag(owner, repoName string) (string, error) {
	client := github.NewClient(nil)

	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repoName)
	if err != nil {
		return "", fmt.Errorf("failed to get latest release: %w", err)
	}

	if release.TagName == nil {
		return "", fmt.Errorf("latest release has no tag name")
	}

	return *release.TagName, nil
}
