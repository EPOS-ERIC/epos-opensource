package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/minio/selfupdate"
	"github.com/spf13/cobra"
)

const MinisignPublicKey = `untrusted comment: minisign public key 6194A08FF2B43F37
RWQ3P7Tyj6CUYaNQMTYSW4G40Xq4cGRwCDNX/GV1qRm8E40+RDbhP4lt`

const (
	windowsOS = "windows"
	macOS     = "darwin"
	linuxOS   = "linux"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the CLI to the latest version.",
	Long:  `Update the CLI to the latest version available on GitHub.`,
	Run: func(cmd *cobra.Command, args []string) {
		release, err := fetchLatestGitHubRelease("EPOS-ERIC", "epos-opensource")
		if err != nil {
			display.Error("Failed to fetch latest release: %v", err)
			os.Exit(1)
		}

		if release.TagName == nil {
			display.Error("Latest release has no tag name")
			os.Exit(1)
		}

		tag := *release.TagName

		current := common.GetVersion()
		if current == "dev" {
			display.Warn("Development version detected. Skipping update.")
			return
		}

		currentVer, err := semver.NewVersion(current)
		if err != nil {
			display.Error("Invalid current version: %v", err)
			return
		}

		latestVer, err := semver.NewVersion(tag)
		if err != nil {
			display.Error("Invalid latest version: %v", err)
			return
		}

		if !currentVer.LessThan(latestVer) {
			display.Info("You are already on the latest version (%s)", current)
			return
		}

		if currentVer.Major() < latestVer.Major() {
			// TODO add an url to the release page to manually check for the breaking changes
			display.Warn("Major version upgrade detected (%s -> %s). This may include breaking changes.", current, tag)
			confirmed, err := common.Confirm("Are you sure you want to continue? (y/n):")
			if err != nil {
				display.Error("Failed to read confirmation: %v", err)
				os.Exit(1)
			}
			if !confirmed {
				display.Info("Update operation cancelled.")
				return
			}
		}

		display.UpdateStarting(current, tag)

		var platformOS, platformArch string
		switch runtime.GOOS {
		case "windows":
			platformOS = windowsOS
		case "linux":
			platformOS = linuxOS
		case "darwin":
			platformOS = macOS
		default:
			display.Error("Unsupported platform: %s", runtime.GOOS)
			os.Exit(1)
		}

		switch runtime.GOARCH {
		case "amd64":
			platformArch = "amd64"
		case "arm64":
			platformArch = "arm64"
		default:
			display.Error("Unsupported architecture: %s", runtime.GOARCH)
			os.Exit(1)
		}

		binaryName := fmt.Sprintf("epos-opensource-%s-%s", platformOS, platformArch)
		if platformOS == windowsOS {
			binaryName += ".exe"
		}
		signatureName := binaryName + ".minisig"

		var binaryURL, signatureURL string
		for _, asset := range release.Assets {
			switch *asset.Name {
			case binaryName:
				binaryURL = *asset.BrowserDownloadURL
			case signatureName:
				signatureURL = *asset.BrowserDownloadURL
			}
		}

		if binaryURL == "" {
			display.Error("No binary found for platform: %s-%s", platformOS, platformArch)
			os.Exit(1)
		}

		if signatureURL == "" {
			display.Error("No signature found for platform: %s-%s", platformOS, platformArch)
			os.Exit(1)
		}

		display.Step("Downloading update...")
		resp, err := http.Get(binaryURL)
		if err != nil {
			display.Error("Failed to download binary: %v", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		binary, err := io.ReadAll(resp.Body)
		if err != nil {
			display.Error("Failed to read binary: %v", err)
			os.Exit(1)
		}

		display.Step("Verifying signature...")
		verifier := selfupdate.NewVerifier()
		err = verifier.LoadFromURL(signatureURL, MinisignPublicKey, http.DefaultTransport)
		if err != nil {
			display.Error("Failed to load signature: %v", err)
			os.Exit(1)
		}

		err = verifier.Verify(binary)
		if err != nil {
			display.Error("Signature verification failed: %v", err)
			os.Exit(1)
		}

		display.Step("Applying update...")
		err = selfupdate.Apply(bytes.NewReader(binary), selfupdate.Options{})
		if err != nil {
			if rerr := selfupdate.RollbackError(err); rerr != nil {
				display.Error("Update failed and rollback failed: %v", rerr)
				os.Exit(1)
			}
			if strings.Contains(strings.ToLower(err.Error()), "permission denied") {
				display.Error("Update failed due to insufficient permissions. Please run with sudo: 'sudo epos-opensource update'")
			} else {
				display.Error("Update failed: %v", err)
			}
			os.Exit(1)
		}

		display.Done("CLI updated successfully to version %s", tag)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
