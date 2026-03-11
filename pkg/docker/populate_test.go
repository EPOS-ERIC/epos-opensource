package docker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPopulateOpts_Validate(t *testing.T) {
	tmpDir := t.TempDir()

	ttlFile := filepath.Join(tmpDir, "test.ttl")
	if err := os.WriteFile(ttlFile, []byte{}, 0o600); err != nil {
		t.Fatalf("failed to create ttl file: %v", err)
	}

	notTtlFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(notTtlFile, []byte{}, 0o600); err != nil {
		t.Fatalf("failed to create txt file: %v", err)
	}

	tests := []struct {
		name    string
		opts    PopulateOpts
		wantErr bool
	}{
		{
			name:    "Environment does not exist",
			opts:    PopulateOpts{Name: "does-not-exist", Parallel: 1, TTLDirs: []string{tmpDir}},
			wantErr: true,
		},
		{
			name:    "Parallel above maximum",
			opts:    PopulateOpts{Name: "test", Parallel: 21, TTLDirs: []string{tmpDir}},
			wantErr: true,
		},
		{
			name:    "Parallel below minimum",
			opts:    PopulateOpts{Name: "test", Parallel: 0, TTLDirs: []string{tmpDir}},
			wantErr: true,
		},
		{
			name:    "Parallel negative",
			opts:    PopulateOpts{Name: "test", Parallel: -1, TTLDirs: []string{tmpDir}},
			wantErr: true,
		},
		{
			name:    "File without ttl extension",
			opts:    PopulateOpts{Name: "test", Parallel: 5, TTLDirs: []string{notTtlFile}},
			wantErr: true,
		},
		{
			name:    "Non-existent path",
			opts:    PopulateOpts{Name: "test", Parallel: 5, TTLDirs: []string{"nonexistent"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
