package resources

import (
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

// shouldCreateServiceMonitors checks if service monitors should be created based on the RunaiConfig
func shouldCreateServiceMonitors(runaiConfig *runv1.RunaiConfig) bool {
	return runaiConfig.Spec.Knative.CreateServiceMonitors != nil && *runaiConfig.Spec.Knative.CreateServiceMonitors
}
