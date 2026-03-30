package completion

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestSharedValuesCompletionFunction(t *testing.T) {
	tests := []struct {
		name       string
		toComplete string
		values     []string
		getterErr  error
		want       []string
	}{
		{
			name:       "matches prefix",
			toComplete: "dev",
			values:     []string{"dev", "dev-eu", "prod"},
			want:       []string{"dev", "dev-eu"},
		},
		{
			name:       "empty prefix returns all values",
			toComplete: "",
			values:     []string{"alpha", "beta"},
			want:       []string{"alpha", "beta"},
		},
		{
			name:       "getter error returns no matches",
			toComplete: "dev",
			getterErr:  errors.New("boom"),
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, directive := SharedValuesCompletion(tt.toComplete, func() ([]string, error) {
				return tt.values, tt.getterErr
			})

			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Fatalf("SharedValuesCompletionFunction() directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("SharedValuesCompletionFunction() = %v, want %v", got, tt.want)
			}
		})
	}
}
