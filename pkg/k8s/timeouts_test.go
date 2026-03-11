package k8s

import (
	"testing"
	"time"
)

func TestResolveCommandTimeout(t *testing.T) {
	tests := []struct {
		name    string
		input   time.Duration
		want    time.Duration
		wantErr bool
	}{
		{
			name:  "Uses default when timeout is unset",
			input: 0,
			want:  defaultCommandTimeout,
		},
		{
			name:  "Uses provided timeout when positive",
			input: 2 * time.Minute,
			want:  2 * time.Minute,
		},
		{
			name:    "Fails when timeout is negative",
			input:   -1 * time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveCommandTimeout(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolveCommandTimeout() error = nil, want error")
				}

				return
			}

			if err != nil {
				t.Fatalf("resolveCommandTimeout() error = %v, want nil", err)
			}

			if got != tt.want {
				t.Fatalf("resolveCommandTimeout() = %s, want %s", got, tt.want)
			}
		})
	}
}
