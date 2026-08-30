package mnnvl

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// hasGPUResources reports whether the container requests an NVIDIA GPU.
// mnnvl is NVIDIA-DRA-specific, so this check is intentionally vendor-bound.
func hasGPUResources(resources corev1.ResourceRequirements) bool {
	for resourceName := range resources.Requests {
		if isNvidiaGPUResource(string(resourceName)) {
			return true
		}
	}
	for resourceName := range resources.Limits {
		if isNvidiaGPUResource(string(resourceName)) {
			return true
		}
	}
	return false
}

func isNvidiaGPUResource(name string) bool {
	return strings.Contains(name, "nvidia.com/gpu")
}
