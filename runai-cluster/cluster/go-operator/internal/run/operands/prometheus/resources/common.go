package resources

import (
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
)

func shouldUseOpenshiftBuiltInPrometheus(config *runaiv1.RunaiConfig) bool {
	return config.Spec.Internal.Global.Platform.KubernetesFlavor == runv1_internal.KubernetesFlavorOpenShift && *config.Spec.Prometheus.Openshift.UseBuiltInPrometheus
}
