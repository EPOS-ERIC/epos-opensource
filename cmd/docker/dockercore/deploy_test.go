package dockercore_test

import (
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/db/sqlc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestDeploy(t *testing.T) {
	tests := []struct {
		name string
		opts dockercore.DeployOpts

		want    *sqlc.Docker
		wantErr bool
	}{
		{
			name:    "deploy with invalid name",
			opts:    dockercore.DeployOpts{Name: ""},
			wantErr: true,
		},
		{
			name: "deploy (default path)",
			opts: dockercore.DeployOpts{Name: "docker-deploy-test"},
			want: &sqlc.Docker{
				Name:           "docker-deploy-test",
				Directory:      "", // ignore
				ApiUrl:         "http://localhost:33000/api/v1",
				GuiUrl:         "http://localhost:32000",
				BackofficeUrl:  "http://localhost:34000",
				ApiPort:        33000,
				GuiPort:        32000,
				BackofficePort: 34000,
			},
		},
		{
			name: "deploy with custom path",
			opts: dockercore.DeployOpts{
				Name: "docker-deploy-test",
				Path: "./../../../.test/",
			},
			want: &sqlc.Docker{
				Name:           "docker-deploy-test",
				Directory:      `.*/\.test/.*`,
				ApiUrl:         "http://localhost:33000/api/v1",
				GuiUrl:         "http://localhost:32000",
				BackofficeUrl:  "http://localhost:34000",
				ApiPort:        33000,
				GuiPort:        32000,
				BackofficePort: 34000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dockercore.Deploy(tt.opts)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("Deploy() unexpected error: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Fatalf("Deploy() = nil error, wantErr = true")
			}

			t.Cleanup(func() {
				if delErr := dockercore.Delete(dockercore.DeleteOpts{Name: tt.opts.Name}); delErr != nil {
					t.Fatalf("cleanup failed: %v", delErr)
				}
			})

			diff := cmp.Diff(tt.want, got, cmpopts.IgnoreFields(sqlc.Docker{}, "Directory"))
			if diff != "" {
				t.Fatalf("Deploy() mismatch (-want +got):\n%s", diff)
			}

			if tt.want.Directory != "" {
				re, regexErr := regexp.Compile(tt.want.Directory)
				if regexErr != nil {
					t.Fatalf("bad Directory regex %q: %v", tt.want.Directory, regexErr)
				}
				if !re.MatchString(got.Directory) {
					t.Fatalf("Directory %q does not match %q", got.Directory, tt.want.Directory)
				}
			}

			testEndpoint(tt.want.ApiUrl+"/ui", t)
			testEndpoint(tt.want.GuiUrl, t)
			testEndpoint(tt.want.BackofficeUrl, t)
		})
	}
}

func testEndpoint(url string, t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("failed to call url '%s': %v", url, err)
	}
	t.Cleanup(func() {
		_ = resp.Body.Close()
	})

	if resp.StatusCode != 200 {
		t.Fatalf("url '%s' answered with non 200 satus code: %d", url, resp.StatusCode)
	}
}
