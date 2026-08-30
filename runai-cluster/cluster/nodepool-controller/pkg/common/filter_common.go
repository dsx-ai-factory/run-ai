package common

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
)

func FilterNodeUpdatesForNPController(nodeOld, nodeNew *corev1.Node) bool {
	// we are only interested in Reconcile if the node has changed phase, Unschedulable or labels
	if nodeOld.Status.Phase != nodeNew.Status.Phase {
		return true
	}

	if nodeOld.Spec.Unschedulable != nodeNew.Spec.Unschedulable {
		return true
	}

	if !reflect.DeepEqual(nodeOld.Labels, nodeNew.Labels) {
		return true
	}

	return false
}
