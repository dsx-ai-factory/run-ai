package mnnvlutils

import (
	"strconv"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
)

// getUnsupportedNodePools returns a list of nodepools that doesn't support MNNVL configuration and, if exist
func getUnsupportedNodePools(nodePools []kaiv1alpha1.NodePool) []string {
	if len(nodePools) == 0 {
		return nil
	}

	var unsupported []string

	for _, nodePool := range nodePools {
		if !isNodePoolValidForMNNVL(nodePool) {
			unsupported = append(unsupported, nodePool.Name)
		}
	}

	return unsupported
}

// isNodePoolValidForMNNVL checks if a NodePool is valid for MNNVL feature
// A NodePool is valid for MNNVL if:
// 1. GPU network acceleration was detected on it
// 2. It has a valid GPU network acceleration label key
//
// On the KAI CRD both live in annotations rather than status/spec. An absent or
// unparseable detection annotation counts as not detected: acceleration relaxes MNNVL
// placement, so failing open would schedule workloads onto unsupported nodes.
func isNodePoolValidForMNNVL(nodePool kaiv1alpha1.NodePool) bool {
	detected, err := strconv.ParseBool(nodePool.Annotations[kaiv1alpha1.AnnotationGPUNetworkAccelerationDetected])
	if err != nil {
		return false
	}
	return detected && gpuNetworkAccelerationLabelKey(nodePool) != ""
}

// gpuNetworkAccelerationLabelKey reads the configured detection label key, falling back to
// the default the legacy CRD schema applied to its spec field.
func gpuNetworkAccelerationLabelKey(nodePool kaiv1alpha1.NodePool) string {
	if labelKey := nodePool.Annotations[kaiv1alpha1.AnnotationGPUNetworkAccelerationLabelKey]; labelKey != "" {
		return labelKey
	}
	return kaiv1alpha1.DefaultGPUNetworkAccelerationLabelKey
}
