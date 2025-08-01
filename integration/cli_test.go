//go:build integration
// +build integration

package integration

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

var (
	update     = flag.Bool("update", false, "update golden files")
	binaryName = "epos-opensource"
	binaryPath = ""
)

type test struct {
	name           string
	args           []string
	fixture        string
	parallelizable bool
	before         func() ([]byte, error)
	cleanup        func() ([]byte, error)
}

func TestCliArgs(t *testing.T) {
	tests := []test{
		{
			name:           "version",
			args:           []string{"--version"},
			fixture:        "version.golden",
			parallelizable: true,
		},

		// -------------------- DOCKER --------------------
		{
			name:           "docker empty list",
			args:           []string{"docker", "list"},
			fixture:        "docker-empty-list.golden",
			parallelizable: true,
		},
		{
			name:           "docker deploy",
			args:           []string{"docker", "deploy", "docker-deploy-test"},
			fixture:        "docker-deploy.golden",
			parallelizable: true,
			cleanup: func() ([]byte, error) {
				return runBinary([]string{"docker", "delete", "docker-deploy-test"})
			},
		},
		{
			name:           "docker delete",
			args:           []string{"docker", "delete", "docker-delete-test"},
			fixture:        "docker-delete.golden",
			parallelizable: true,
			before: func() ([]byte, error) {
				return runBinary([]string{"docker", "deploy", "docker-delete-test"})
			},
		},
		{
			name:           "docker update",
			args:           []string{"docker", "update", "docker-update-test"},
			fixture:        "docker-update.golden",
			parallelizable: true,
			before: func() ([]byte, error) {
				return runBinary([]string{"docker", "deploy", "docker-update-test"})
			},
			cleanup: func() ([]byte, error) {
				return runBinary([]string{"docker", "delete", "docker-update-test"})
			},
		},

		// -------------------- KUBERNETES --------------------
		{
			name:           "kubernetes empty list",
			args:           []string{"kubernetes", "list"},
			fixture:        "kubernetes-empty-list.golden",
			parallelizable: true,
		},
		{
			name:           "kubernetes deploy",
			args:           []string{"kubernetes", "deploy", "kubernetes-deploy-test"},
			fixture:        "kubernetes-deploy.golden",
			parallelizable: true,
			cleanup: func() ([]byte, error) {
				return runBinary([]string{"kubernetes", "delete", "kubernetes-deploy-test"})
			},
		},
		{
			name:           "kubernetes delete",
			args:           []string{"kubernetes", "delete", "kubernetes-delete-test"},
			fixture:        "kubernetes-delete.golden",
			parallelizable: true,
			before: func() ([]byte, error) {
				return runBinary([]string{"kubernetes", "deploy", "kubernetes-delete-test"})
			},
		},
		{
			name:           "kubernetes update",
			args:           []string{"kubernetes", "update", "kubernetes-update-test"},
			fixture:        "kubernetes-delete.golden",
			parallelizable: true,
			before: func() ([]byte, error) {
				return runBinary([]string{"kubernetes", "deploy", "kubernetes-update-test"})
			},
			cleanup: func() ([]byte, error) {
				return runBinary([]string{"kubernetes", "delete", "kubernetes-update-test"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.parallelizable {
				t.Parallel()
			}
			if tt.before != nil {
				beforeOut, err := tt.before()
				if err != nil {
					t.Fatalf("cmd %q before cmd failed: %v\n--- before output ---\n%s", tt.args, err, beforeOut)
				}
				t.Logf("before: %s", beforeOut)
			}

			output, err := runBinary(tt.args)
			if err != nil {
				t.Fatalf("cmd %q failed: %v\n--- output ---\n%s", tt.args, err, output)
			}

			normOut := normalize(output)

			if *update {
				writeFixture(t, tt.fixture, normOut)
			}

			if tt.cleanup != nil {
				defer func() {
					if cleanupOut, err := tt.cleanup(); err != nil {
						t.Fatalf("cmd %q cleanup failed: %v\n--- output ---\n%s\n--- cleanup output ---\n%s", tt.args, err, output, cleanupOut)
					}
				}()
			}

			expected := loadFixture(t, tt.fixture)
			if diff := cmp.Diff(expected, string(normOut)); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMain(m *testing.M) {
	err := os.Chdir("..")
	if err != nil {
		fmt.Printf("failed to change dir: %v", err)
		os.Exit(1)
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("failed to get current dir: %v", err)
		os.Exit(1)
	}

	binaryPath = filepath.Join(dir, binaryName)

	os.Exit(m.Run())
}

func runBinary(args []string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	return cmd.CombinedOutput()
}

func fixturePath(t *testing.T, fixture string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("problems recovering caller information")
	}

	return filepath.Join(filepath.Dir(filename), fixture)
}

func writeFixture(t *testing.T, fixture string, content []byte) {
	err := os.WriteFile(fixturePath(t, fixture), content, 0o600)
	if err != nil {
		t.Fatal(err)
	}
}

func loadFixture(t *testing.T, fixture string) string {
	content, err := os.ReadFile(fixturePath(t, fixture))
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}
