package topology_controller

import (
	grovev1alpha1 "github.com/ai-dynamo/grove/operator/api/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	groveResource = "clustertopologybindings"
	groveGroup    = "grove.io"
	groveCRDName  = groveResource + "." + groveGroup

	// kaiSchedulerName is the SchedulerName the grove operator uses
	// to identify the KAI scheduler backend in ClusterTopologyBindingStatus.
	kaiSchedulerName = "kai-scheduler"
)

// groveToKaiTopologyName returns the KAI Topology name that the given
// grove ClusterTopologyBinding maps to, by looking up the kai-scheduler entry
// in Status.SchedulerTopologyStatuses. Returns "" if no such entry exists.
func groveToKaiTopologyName(groveTopology client.Object) string {
	ct, ok := groveTopology.(*grovev1alpha1.ClusterTopologyBinding)
	if !ok {
		return ""
	}
	for _, status := range ct.Status.SchedulerTopologyStatuses {
		if status.SchedulerName == kaiSchedulerName {
			return status.TopologyReference
		}
	}
	return ""
}
