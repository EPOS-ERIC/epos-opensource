package k8s

import "github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"

type Env struct {
	config.EnvConfig

	Name    string
	Context string
}
