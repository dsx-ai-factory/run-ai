package workloadutils

import (
	"context"
	"errors"

	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
)

func ValidateAndHandelServingConfigurationFields(ctx context.Context, responseSpec *workloadsApi.SupersetSpec, clusterId string) error {
	servingConfiguration := responseSpec.ServingConfiguration.Get()

	// we only validate the serving configuration RequestTimeoutSeconds vs the cluster version.
	// for InitializationTimeoutSeconds is handled for old cluster versions in the synchronizeInitializationTimeouts function
	// we set the InitializationTimeoutSeconds in the Autoscaling section
	if servingConfiguration != nil && servingConfiguration.RequestTimeoutSeconds.IsSet() {
		clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId)
		clusterSupportMin, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionOfInferenceServingConfiguration)
		if err != nil {
			return err
		}

		if !clusterSupportMin {
			return errors.New("RequestTimeoutSeconds is not supported in this cluster (minimal supported cluster version: 2.22)")
		}
	}

	SynchronizeInitializationTimeouts(responseSpec)
	return nil
}

// SynchronizeInitializationTimeouts synchronizes InitializationTimeoutSeconds between Autoscaling and ServingConfiguration.
// 1. If ServingConfiguration.InitializationTimeoutSeconds is set, it overrides Autoscaling.InitializationTimeoutSeconds.
// 2. Else, if Autoscaling.InitializationTimeoutSeconds is set, it's used for ServingConfiguration.InitializationTimeoutSeconds.
func SynchronizeInitializationTimeouts(responseSpec *workloadsApi.SupersetSpec) {
	autoscalingFields := responseSpec.Autoscaling.Get()
	servingConfiguration := responseSpec.ServingConfiguration.Get()

	if servingConfiguration != nil && servingConfiguration.InitializationTimeoutSeconds.IsSet() {
		// Case 1: ServingConfiguration's timeout is set; apply it to Autoscaling.
		scTimeoutVal := servingConfiguration.InitializationTimeoutSeconds.Get()
		if autoscalingFields != nil {
			autoscalingFields.InitializationTimeoutSeconds.Set(scTimeoutVal)
		} else {
			// If Autoscaling doesn't exist, create it. and set the InitializationTimeoutSeconds.
			newAutoscalingConfig := workloadsApi.AutoScaling{}
			newAutoscalingConfig.InitializationTimeoutSeconds.Set(scTimeoutVal)
			responseSpec.Autoscaling.Set(&newAutoscalingConfig)
		}
	} else {
		// Case 2: ServingConfiguration's timeout is NOT set.
		// If Autoscaling's timeout IS set, apply it to ServingConfiguration.
		if autoscalingFields != nil && autoscalingFields.InitializationTimeoutSeconds.IsSet() {
			asTimeoutVal := autoscalingFields.InitializationTimeoutSeconds.Get()
			if servingConfiguration == nil {
				// If ServingConfiguration doesn't exist, create it.
				newServingConfig := workloadsApi.ServingConfiguration{}
				newServingConfig.InitializationTimeoutSeconds.Set(asTimeoutVal)
				responseSpec.ServingConfiguration.Set(&newServingConfig)
			} else {
				// ServingConfiguration exists, set its timeout.
				servingConfiguration.InitializationTimeoutSeconds.Set(asTimeoutVal)
			}
		}
	}
}
