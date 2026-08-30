package patcher

import (
	"context"

	"strconv"

	metametrics "github.com/run-ai/runai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	podStatusAnnotationName             = "runai-calculated-status"
	PodAllocatedGPUsAnnotationName      = "runai-allocated-gpus"
	PodAllocatedGpuMemoryAnnotationName = "runai-allocated-gpu-memory"
	PodAllocatedMigGPUsAnnotationName   = "runai-allocated-mig-gpus"
)

func ShouldUpdatePodAnnotations(pod *v1.Pod, podMetrics metametrics.PodMetadata) bool {
	if len(pod.Annotations) == 0 {
		return true
	}

	for k, v := range transformToPodAnnotations(podMetrics) {
		if pod.Annotations[k] != v {
			return true
		}
	}

	return false
}

func UpdatePodAnnotations(ctx context.Context, pod *v1.Pod, podMetrics metametrics.PodMetadata,
	kubeClient client.Client) error {
	if len(pod.Annotations) == 0 {
		pod.Annotations = map[string]string{}
	}

	newPod := pod.DeepCopy()
	for k, v := range transformToPodAnnotations(podMetrics) {
		newPod.Annotations[k] = v
	}

	return kubeClient.Patch(ctx, newPod, client.MergeFrom(pod))
}

func transformToPodAnnotations(podMetrics metametrics.PodMetadata) map[string]string {
	return map[string]string{
		podStatusAnnotationName:             podMetrics.Status,
		PodAllocatedGPUsAnnotationName:      strconv.FormatFloat(podMetrics.GPUs, 'f', -1, 64),
		PodAllocatedGpuMemoryAnnotationName: strconv.FormatInt(podMetrics.GpuMemory, 10),
		PodAllocatedMigGPUsAnnotationName:   strconv.FormatFloat(podMetrics.MigGPUs, 'f', -1, 64),
	}
}
