package utils

import runaiv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

func GetNodePoolNameFromLabels(labels map[string]string) string {
	nodePoolName, found := labels[runaiv1alpha1.NodePoolNameLabel]
	if !found || nodePoolName == "" {
		nodePoolName = runaiv1alpha1.DefaultNodePoolName
	}

	return nodePoolName
}
