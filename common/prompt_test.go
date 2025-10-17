package common

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestConfirm(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		{
			name:    "confirm with y",
			input:   "y\n",
			want:    true,
			wantErr: false,
		},
		{
			name:    "confirm with yes",
			input:   "yes\n",
			want:    true,
			wantErr: false,
		},
		{
			name:    "confirm with Y uppercase",
			input:   "Y\n",
			want:    true,
			wantErr: false,
		},
		{
			name:    "confirm with YES uppercase",
			input:   "YES\n",
			want:    true,
			wantErr: false,
		},
		{
			name:    "deny with n",
			input:   "n\n",
			want:    false,
			wantErr: false,
		},
		{
			name:    "deny with no",
			input:   "no\n",
			want:    false,
			wantErr: false,
		},
		{
			name:    "deny with N uppercase",
			input:   "N\n",
			want:    false,
			wantErr: false,
		},
		{
			name:    "deny with NO uppercase",
			input:   "NO\n",
			want:    false,
			wantErr: false,
		},
		{
			name:    "invalid then confirm",
			input:   "invalid\ny\n",
			want:    true,
			wantErr: false,
		},
		{
			name:    "empty then deny",
			input:   "\nn\n",
			want:    false,
			wantErr: false,
		},
		{
			name:    "whitespace then confirm",
			input:   "   \nyes\n",
			want:    true,
			wantErr: false,
		},
		{
			name:    "multiple invalid then confirm",
			input:   "maybe\nwhat\ny\n",
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r

			go func() {
				defer w.Close()
				_, _ = io.WriteString(w, tt.input)
			}()

			got, err := Confirm("Test prompt:")

			os.Stdin = oldStdin

			if (err != nil) != tt.wantErr {
				t.Errorf("Confirm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Confirm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfirm_InvalidInputRetry(t *testing.T) {
	inputs := []string{"invalid", "maybe", "y"}
	expectedPrompts := len(inputs)

	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		defer w.Close()
		_, _ = io.WriteString(w, strings.Join(inputs, "\n")+"\n")
	}()

	got, err := Confirm("Test prompt:")

	os.Stdin = oldStdin

	if err != nil {
		t.Errorf("Confirm() unexpected error = %v", err)
		return
	}
	if !got {
		t.Errorf("Confirm() = %v, want true", got)
	}

	if expectedPrompts != 3 {
		t.Logf("Expected to process %d inputs before confirming", expectedPrompts)
	}
}
