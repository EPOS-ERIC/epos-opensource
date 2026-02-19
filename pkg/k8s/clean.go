package k8s

import (
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/validate"
)

type CleanOpts struct {
	// Required. name of the environment
	Name string
	// Optional. Kubernetes context to use; defaults to the current kubectl context when unset.
	Context string
}

func Clean(opts CleanOpts) (*Env, error) {
	// TODO this command to maybe just be an update with setting to true the job to clean up the db?
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid clean parameters: %w", err)
	}

	// kube, err := db.GetK8sByName(opts.Name)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get k8s info for %s: %w", opts.Name, err)
	// }
	//
	// dir := kube.Directory
	// ctx := kube.Context
	// ns := opts.Name
	//
	// handleFailure := func(msg string, mainErr error) (*sqlc.K8s, error) {
	// 	display.Error("Clean failed, attempting recovery")
	// 	// if err := deployManifests(dir, ns, false, ctx, kube.TlsEnabled); err != nil {
	// 	// 	return nil, fmt.Errorf("recovery failed: %w", err)
	// 	// }
	// 	return nil, fmt.Errorf(msg, mainErr)
	// }
	//
	// // Scale down metadata database
	// display.Step("Scaling down metadata database")
	// if err := runKubectl(dir, false, ctx, "scale", "deployment", "metadata-database", "--replicas=0", "-n", ns); err != nil {
	// 	return handleFailure("failed to scale down metadata database: %w", err)
	// }
	//
	// // Scale down cache-dependent services
	// display.Step("Scaling down services with cached data")
	// services := []string{
	// 	"resources-service",
	// 	"ingestor-service",
	// 	"external-access-service",
	// 	"backoffice-service",
	// }
	// for _, svc := range services {
	// 	if err := runKubectl(dir, false, ctx, "scale", "deployment", svc, "--replicas=0", "-n", ns); err != nil {
	// 		return handleFailure("failed to scale down %s: %w", fmt.Errorf("%s: %w", svc, err))
	// 	}
	// }
	//
	// // Delete PVC
	// display.Step("Deleting PostgreSQL data volume")
	// if err := runKubectl(dir, false, ctx, "delete", "pvc", "psqldata", "-n", ns); err != nil {
	// 	return handleFailure("failed to delete PVC: %w", err)
	// }
	//
	// // Recreate PVC
	// display.Step("Recreating PostgreSQL data volume")
	// if err := runKubectl(dir, false, ctx, "apply", "-f", "pvc-psqldata.yaml", "-n", ns); err != nil {
	// 	return handleFailure("failed to recreate PVC: %w", err)
	// }
	//
	// // Wait for PVC to be bound
	// display.Step("Waiting for volume to be bound")
	// // if err := waitForPVCBound(ctx, ns, "psqldata"); err != nil {
	// // 	return handleFailure("PVC failed to bind: %w", err)
	// // }
	//
	// // Scale up in order: infra first, then services
	// display.Step("Restarting metadata database")
	// if err := runKubectl(dir, false, ctx, "scale", "deployment", "metadata-database", "--replicas=1", "-n", ns); err != nil {
	// 	return handleFailure("failed to scale up metadata database: %w", err)
	// }
	// if err := runKubectl(dir, false, ctx, "rollout", "status", "deployment/metadata-database", "-n", ns, "--timeout=30s"); err != nil {
	// 	return handleFailure("metadata database failed to start: %w", err)
	// }
	//
	// // Scale up services
	// display.Step("Restarting services")
	// for _, svc := range services {
	// 	if err := runKubectl(dir, false, ctx, "scale", "deployment", svc, "--replicas=1", "-n", ns); err != nil {
	// 		return handleFailure("failed to scale up %s: %w", fmt.Errorf("%s: %w", svc, err))
	// 	}
	// }
	// // Wait for all services
	// for _, svc := range services {
	// 	if err := runKubectl(dir, false, ctx, "rollout", "status", fmt.Sprintf("deployment/%s", svc), "-n", ns, "--timeout=60s"); err != nil {
	// 		return handleFailure("%s failed to start: %w", fmt.Errorf("%s: %w", svc, err))
	// 	}
	// }
	//
	// // Clear ingested files
	// if err := db.DeleteIngestedFilesByEnvironment("k8s", opts.Name); err != nil {
	// 	return handleFailure("failed to clear ingested files: %w", err)
	// }
	//
	// // Populate ontologies
	// if err := common.PopulateOntologies(kube.ApiUrl); err != nil {
	// 	return handleFailure("failed to populate ontologies: %w", err)
	// }
	//
	// display.Done("Cleaned environment: %s", opts.Name)
	// return kube, nil
	return nil, nil
}

func (c *CleanOpts) Validate() error {
	if err := validate.EnvironmentExistsK8s(c.Name); err != nil {
		return fmt.Errorf("no environment with the name '%s' exists: %w", c.Name, err)
	}

	return nil
}
