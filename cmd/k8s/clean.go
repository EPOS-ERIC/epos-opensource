package k8s

import (
	"fmt"
	"os"

	"github.com/epos-eu/epos-opensource/cmd/k8s/k8score"
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/spf13/cobra"
)

var CleanCmd = &cobra.Command{
	Use:   "clean [env-name]",
	Short: "Clean the data of an environment.",
	Long: `Clean the data of an environment without redeploying. 
After clean all custom data populated in the environment will be lost. 
This action is irreversible.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: validArgsFunction,
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if !cleanForce {
			display.Warn("This will permanently delete all data in environment '%s'. This action cannot be undone.", name)
			confirmed, err := common.Confirm("Are you sure you want to continue? (y/n):")
			if err != nil {
				display.Error("Failed to read confirmation: %v", err)
				os.Exit(1)
			}
			if !confirmed {
				display.Info("Clean operation cancelled.")
				return
			}
		}

		kube, err := k8score.Clean(k8score.CleanOpts{
			Name: name,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		display.Urls(kube.GuiUrl, kube.ApiUrl, kube.BackofficeUrl, fmt.Sprintf("epos-opensource k8s clean %s", name))
	},
}

func init() {
	CleanCmd.Flags().BoolVarP(&cleanForce, "force", "f", false, "Force clean without confirmation prompt")
}
