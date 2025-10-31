package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/display"
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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Name() == "update" {
			return
		}

		tag, err := getLatestTagWithCache("EPOS-ERIC", "epos-opensource")
		if err != nil {
			log.Println("Failed to get latest tag:", err)
			return
		}

		current := common.GetVersion()
		if current == "dev" {
			return
		}

		currentVer, err := semver.NewVersion(current)
		if err != nil {
			log.Println("Failed to parse current version:", err)
			return
		}

		latestVer, err := semver.NewVersion(tag)
		if err != nil {
			log.Println("Failed to parse latest tag:", err)
			return
		}

		if currentVer.LessThan(latestVer) {
			display.UpdateAvailable(current, tag)
		}
	},

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
		log.Println("Failed to execute root command:", err)
		os.Exit(1)
	}
}

// fetchLatestGitHubRelease fetches the latest release from GitHub
func fetchLatestGitHubRelease(owner, repoName string) (*github.RepositoryRelease, error) {
	client := github.NewClient(nil)

	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repoName)
	if err != nil {
		return nil, err
	}

	return release, nil
}

// getLatestTagWithCache gets the latest tag with caching
func getLatestTagWithCache(owner, repoName string) (string, error) {
	cached, err := db.GetLatestReleaseCache()
	if err == nil && cached.FetchedAt != nil && time.Since(*cached.FetchedAt) < 12*time.Hour {
		return cached.TagName, nil
	}

	release, err := fetchLatestGitHubRelease(owner, repoName)
	if err != nil || release.TagName == nil {
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}

	err = db.UpsertLatestReleaseCache(*release.TagName, time.Now())
	if err != nil {
		return "", fmt.Errorf("failed to upsert latest release cache: %w", err)
	}
	return *release.TagName, nil
}
