package docker

import (
	"fmt"
	"os"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker"
	"github.com/spf13/cobra"
)

var dockerGetOutputPath string

var GetCmd = &cobra.Command{
	Use:               "get [env-name]",
	Short:             "Get the currently applied Docker environment configuration.",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: validArgsFunction,
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		env, err := docker.GetEnv(name)
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		configYAML, err := env.Bytes()
		if err != nil {
			display.Error("failed to marshal docker config: %v", err)
			os.Exit(1)
		}

		if dockerGetOutputPath != "" {
			if err := os.WriteFile(dockerGetOutputPath, configYAML, 0o644); err != nil {
				display.Error("failed to write config to %q: %v", dockerGetOutputPath, err)
				os.Exit(1)
			}

			display.Done("Applied config written to: %s", dockerGetOutputPath)
			return
		}

		_, err = fmt.Fprint(display.Stdout, string(configYAML))
		if err != nil {
			display.Error("failed to print config: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	GetCmd.Flags().StringVar(&dockerGetOutputPath, "output", "", "Write the applied configuration YAML to a file")
}
