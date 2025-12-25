package k8s

import (
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/spf13/cobra"
)

func validArgsFunction(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return common.SharedValidArgsFunction(cmd, args, toComplete, func() ([]string, error) {
		k8sEnvs, err := db.GetAllK8s()
		if err != nil {
			return nil, err
		}
		names := make([]string, len(k8sEnvs))
		for i, k := range k8sEnvs {
			names[i] = k.Name
		}
		return names, nil
	})
}
