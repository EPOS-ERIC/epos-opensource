package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/Masterminds/semver/v3"
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
	Long:  "Update the CLI to the latest version. Downloads the latest GitHub release for your platform and replaces the current binary. Prompts for confirmation before major version upgrades.",
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

		// If current version has pre-release and same base as latest it's a dev build of the current release
		if currentVer.Prerelease() != "" &&
			currentVer.Major() == latestVer.Major() &&
			currentVer.Minor() == latestVer.Minor() &&
			currentVer.Patch() == latestVer.Patch() {
			display.Info("You are already on the latest version (%s)", current)
			return
		}

		if !currentVer.LessThan(latestVer) {
			display.Info("You are already on the latest version (%s)", current)
			return
		}

		if currentVer.Major() < latestVer.Major() {
			display.Warn("Major version upgrade detected (%s -> %s). This may include breaking changes.", current, tag)

			if release.Body != nil {
				breakingChanges, found := extractBreakingChangesSection(*release.Body)
				if found {
					display.Warn("The following section is marked as BREAKING CHANGES by the release author:")
					_, _ = fmt.Fprintf(display.Stdout, "\n%s\n\n", breakingChanges)
				} else {
					display.Warn("No '# Breaking Change(s)' section was found in the release notes.")
				}
			} else {
				display.Warn("Release notes are not available for this release.")
			}

			if release.HTMLURL != nil && *release.HTMLURL != "" {
				display.Info("Review full release notes: %s", *release.HTMLURL)
			}

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

func extractBreakingChangesSection(body string) (string, bool) {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	lines := strings.Split(body, "\n")

	start := -1
	for i, line := range lines {
		if isBreakingChangesHeading(line) {
			start = i + 1
			break
		}
	}

	if start == -1 {
		return "", false
	}

	end := len(lines)
	for i := start; i < len(lines); i++ {
		if isMarkdownHeading(lines[i]) {
			end = i
			break
		}
	}

	section := strings.TrimSpace(strings.Join(lines[start:end], "\n"))
	if section == "" {
		return "", false
	}

	return section, true
}

func isBreakingChangesHeading(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "#") {
		return false
	}

	idx := 0
	for idx < len(trimmed) && trimmed[idx] == '#' {
		idx++
	}

	heading := strings.TrimSpace(trimmed[idx:])
	return strings.EqualFold(heading, "breaking change") || strings.EqualFold(heading, "breaking changes")
}

func isMarkdownHeading(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "#") {
		return false
	}

	idx := 0
	for idx < len(trimmed) && trimmed[idx] == '#' {
		idx++
	}

	if idx == 0 {
		return false
	}

	return strings.TrimSpace(trimmed[idx:]) != ""
}
