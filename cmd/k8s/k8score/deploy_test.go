package k8score

import (
	"testing"
)

func TestDeployOpts_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    DeployOpts
		wantErr bool
	}{
		{
			name:    "Environment does not exist",
			opts:    DeployOpts{Name: "does-not-exist"},
			wantErr: true,
		},
		{
			name:    "Invalid Name for environment",
			opts:    DeployOpts{Name: "787/(//&&$%£$£$£&)(/((=??=)))"},
			wantErr: true,
		},
		{
			name:    "Invalid Path to Custom environment",
			opts:    DeployOpts{Name: "test", Path: "/Invalid/path.exe"},
			wantErr: true,
		},
		{
			name:    "Invalid Path to  Enviroment file",
			opts:    DeployOpts{Name: "test", EnvFile: "/this/is/a/dir"},
			wantErr: true,
		},
		{
			name:    "Invalid protocol",
			opts:    DeployOpts{Name: "test", Protocol: "thisSiteIsSafeDontWorry"},
			wantErr: true,
		},
		{
			name:    "Invalid Path to Manifest dir",
			opts:    DeployOpts{Name: "test", ManifestDir: "/this/is/a/dir/test.txt"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("Deploy Kubernetes Validation Test Failed = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
