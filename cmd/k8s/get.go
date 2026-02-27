package k8s

import (
	"fmt"
	"os"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var k8sGetOutputPath string

var GetCmd = &cobra.Command{
	Use:               "get [env-name]",
	Short:             "Get the currently applied K8s environment configuration.",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: validArgsFunction,
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		env, err := k8s.GetEnv(name, context)
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		configYAML, err := yaml.Marshal(env.Config)
		if err != nil {
			display.Error("failed to marshal k8s config: %v", err)
			os.Exit(1)
		}

		if k8sGetOutputPath != "" {
			if err := os.WriteFile(k8sGetOutputPath, configYAML, 0o644); err != nil {
				display.Error("failed to write config to %q: %v", k8sGetOutputPath, err)
				os.Exit(1)
			}

			display.Done("Applied config written to: %s", k8sGetOutputPath)
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
	GetCmd.Flags().StringVar(&context, "context", "", "Kubectl context to use. Uses current context if not set")
	GetCmd.Flags().StringVar(&k8sGetOutputPath, "output", "", "Write the applied configuration YAML to a file")
}
