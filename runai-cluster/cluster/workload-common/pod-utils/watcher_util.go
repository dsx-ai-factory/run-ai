package pod_utils

import (
	"github.com/samber/lo"
	"reflect"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func PodShouldUpdate(e event.UpdateEvent) bool {
	oldPod := e.ObjectOld.(*corev1.Pod)
	newPod := e.ObjectNew.(*corev1.Pod)
	return IsPodStatusAnnotationChanged(oldPod, newPod) ||
		IsPodPhaseChanged(oldPod, newPod) ||
		IsContainerStatusChanged(oldPod, newPod) ||
		IsPodConditionsChanged(oldPod, newPod)
}

func IsPodStatusAnnotationChanged(oldPod *corev1.Pod, newPod *corev1.Pod) bool {
	return oldPod.GetAnnotations()[constants.PodStatusAnnotation] !=
		newPod.GetAnnotations()[constants.PodStatusAnnotation]
}

// IsPodPhaseChanged checks if the pod's status.phase has changed
func IsPodPhaseChanged(oldPod *corev1.Pod, newPod *corev1.Pod) bool {
	return oldPod.Status.Phase != newPod.Status.Phase
}

// IsContainerStatusChanged checks if any container's or init container's ready status or state has changed
func IsContainerStatusChanged(oldPod *corev1.Pod, newPod *corev1.Pod) bool {
	// Combine containers and init containers into single slices for comparison
	oldStatuses := append(oldPod.Status.ContainerStatuses, oldPod.Status.InitContainerStatuses...)
	newStatuses := append(newPod.Status.ContainerStatuses, newPod.Status.InitContainerStatuses...)

	// If the number of container statuses changed
	if len(oldStatuses) != len(newStatuses) {
		return true
	}

	// Create maps for easier comparison
	oldStatusMap := lo.SliceToMap(oldStatuses, func(status corev1.ContainerStatus) (string, corev1.ContainerStatus) {
		return status.Name, status
	})

	newStatusMap := lo.SliceToMap(newStatuses, func(status corev1.ContainerStatus) (string, corev1.ContainerStatus) {
		return status.Name, status
	})
	// Compare each container status
	for name, newStatus := range newStatusMap {
		oldStatus, exists := oldStatusMap[name]
		if !exists {
			return true // New container appeared
		}

		// Check if ready status changed
		if oldStatus.Ready != newStatus.Ready {
			return true
		}

		// Check if container state changed
		if !reflect.DeepEqual(oldStatus.State, newStatus.State) {
			return true
		}
	}

	return false
}

// IsPodConditionsChanged checks if the pod's status.conditions have changed
func IsPodConditionsChanged(oldPod *corev1.Pod, newPod *corev1.Pod) bool {
	if len(oldPod.Status.Conditions) != len(newPod.Status.Conditions) {
		return true
	}

	oldCondMap := lo.SliceToMap(oldPod.Status.Conditions, func(cond corev1.PodCondition) (corev1.PodConditionType, corev1.PodCondition) {
		return cond.Type, cond
	})

	for _, newCond := range newPod.Status.Conditions {
		oldCond, exists := oldCondMap[newCond.Type]
		if !exists {
			return true // New condition appeared
		}
		if oldCond.Status != newCond.Status || oldCond.Reason != newCond.Reason || oldCond.Message != newCond.Message {
			return true
		}
	}

	return false
}
