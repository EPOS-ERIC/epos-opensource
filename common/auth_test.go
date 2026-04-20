package common

import "testing"

func TestTrimAuthURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace string",
			input:    "   ",
			expected: "",
		},
		{
			name:     "already auth root",
			input:    "https://auth.example.com",
			expected: "https://auth.example.com",
		},
		{
			name:     "auth root with trailing slash",
			input:    "https://auth.example.com/",
			expected: "https://auth.example.com",
		},
		{
			name:     "userinfo endpoint",
			input:    "https://auth.example.com/oauth2/userinfo",
			expected: "https://auth.example.com",
		},
		{
			name:     "userinfo endpoint with trailing slash",
			input:    "https://auth.example.com/oauth2/userinfo/",
			expected: "https://auth.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrimAuthURL(tt.input)

			if got != tt.expected {
				t.Fatalf("TrimAuthURL() = %q, expected %q", got, tt.expected)
			}
		})
	}
}
