package docker

import (
	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/db"
	"github.com/spf13/cobra"
)

func validArgsFunction(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return common.SharedValidArgsFunction(cmd, args, toComplete, func() ([]string, error) {
		dockers, err := db.GetAllDocker()
		if err != nil {
			return nil, err
		}
		names := make([]string, len(dockers))
		for i, d := range dockers {
			names[i] = d.Name
		}
		return names, nil
	})
}
