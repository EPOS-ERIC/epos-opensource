//go:build integration
// +build integration

package integration

import (
	"bytes"
	"os"
	"regexp"
)

type scrubRule struct {
	re   *regexp.Regexp // what to match
	repl []byte         // what to replace it with
}

var scrubRules = []scrubRule{
	// colors
	{regexp.MustCompile(`\x1b\[[0-9;]*m`), nil},

	// absolute path printed by "docker deploy"
	{
		regexp.MustCompile(`Environment created in directory: .*`),
		[]byte("Environment created in directory: /testdir/"),
	},

	// absolute path printed by "docker deploy"
	{
		regexp.MustCompile(`Deleting environment directory: .*`),
		[]byte("Deleting environment directory: /testdir/"),
	},
	{
		regexp.MustCompile(`Deleted environment directory: .*`),
		[]byte("Deleted environment directory: /testdir/"),
	},
	{
		regexp.MustCompile(`^.*The requested image's platform (linux/amd64) does not match the detected host platform (linux/arm64/v8) and no specific platform was requested$`),
		[]byte(""),
	},
	{
		regexp.MustCompile(`Backup created at: .*`),
		[]byte("Backup created at: /backup/dir"),
	},
	{
		regexp.MustCompile(`Updated environment created in directory: .*`),
		[]byte("Updated environment created in directory: /testdir/"),
	},
	{
		regexp.MustCompile(`Cleaning up backup directory: .*`),
		[]byte("Cleaning up backup directory: /backup/dir"),
	},
	{
		regexp.MustCompile(`Backup directory cleaned up: .*`),
		[]byte("Backup directory cleaned up: /backup/dir"),
	},

	{
		regexp.MustCompile(`Ingress assigned hostname: .*`),
		[]byte("Ingress assigned hostname: 123.456.7.890"),
	},

	{
		regexp.MustCompile(`(https?://)(?:\d{1,3}\.){3}\d{1,3}`),
		[]byte("123.456.7.890"),
	},
}

func normalize(out []byte) []byte {
	out = bytes.ReplaceAll(out, []byte{os.PathSeparator}, []byte("/"))

	for _, r := range scrubRules {
		out = r.re.ReplaceAll(out, r.repl)
	}
	return out
}
