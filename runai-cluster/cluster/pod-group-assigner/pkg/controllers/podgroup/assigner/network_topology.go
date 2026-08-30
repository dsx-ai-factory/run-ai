package assigner

import (
	"context"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/utils"
)

const (
	// TopologySourceAnnotationKey marks a PodGroup whose topology constraint was stamped by the
	// pod-group-assigner (rather than requested by the user). Its presence lets reconcile tell a
	// system-owned constraint - which it may refresh or clear - apart from a user-owned one, which
	// it must never touch.
	TopologySourceAnnotationKey = "topology.kai/source"
	// TopologySourceSystem is the value written under TopologySourceAnnotationKey.
	TopologySourceSystem = "system"
)

// desiredNetworkTopology computes the topology constraint the pod-group-assigner wants on the
// PodGroup for the given picked nodepool, and whether that constraint is system-sourced, following
// the rules below. The caller diffs the result against the PodGroup's current topology to decide
// whether an update is needed.
//
//  1. The user owns the constraint (it is set and the source annotation is absent) -> leave as is.
//  2. The picked nodepool has a NetworkTopologyName -> set a preferred constraint at the
//     topology's lowest level and mark it system-sourced.
//  3. The nodepool has no topology and the constraint was previously system-sourced -> clear it.
//  4. Otherwise -> nothing to do.
func (pga *PodGroupAssigner) desiredNetworkTopology(ctx context.Context, podGroup *kaiv2alpha2.PodGroup, nodePoolName string) (
	desiredConstraint kaiv2alpha2.TopologyConstraint, desiredSystemSourced bool, err error) {
	previousConstraint := podGroup.Spec.TopologyConstraint
	previousSystemSourced := podGroup.Annotations[TopologySourceAnnotationKey] == TopologySourceSystem

	// Case 1: the user requested a topology themselves - never override it.
	if !isTopologyConstraintEmpty(previousConstraint) && !previousSystemSourced {
		return previousConstraint, false, nil
	}

	networkTopologyName, err := utils.GetNodePoolPreferredNetworkTopologyName(ctx, pga.Client, nodePoolName)
	if err != nil {
		// Lookup failed - keep the existing topology instead of clearing it, and retry next reconcile.
		return previousConstraint, previousSystemSourced, err
	}

	if networkTopologyName != "" {
		// Case 2: add the picked nodepool's topology as a preferred constraint at the lowest level.
		lowestLevel, err := utils.GetNetworkTopologyLowestLevel(ctx, pga.Client, networkTopologyName)
		if err != nil {
			// Keep the existing topology on a lookup failure (see above).
			return previousConstraint, previousSystemSourced, err
		}

		desiredConstraint = kaiv2alpha2.TopologyConstraint{
			Topology:               networkTopologyName,
			PreferredTopologyLevel: lowestLevel,
		}
		return desiredConstraint, true, nil
	}

	// Case 3: the nodepool no longer has a topology - clear a constraint we previously added.
	if previousSystemSourced {
		return kaiv2alpha2.TopologyConstraint{}, false, nil
	}

	// Case 4: no topology to add and nothing of ours to clear.
	return previousConstraint, false, nil
}

// applyTopologyToPodGroup writes the desired topology constraint onto the PodGroup and keeps the
// source annotation in sync: present when the constraint is system-sourced, removed otherwise.
func applyTopologyToPodGroup(podGroup *kaiv2alpha2.PodGroup, constraint kaiv2alpha2.TopologyConstraint, systemSourced bool) {
	podGroup.Spec.TopologyConstraint = constraint

	if systemSourced {
		if podGroup.Annotations == nil {
			podGroup.Annotations = map[string]string{}
		}
		podGroup.Annotations[TopologySourceAnnotationKey] = TopologySourceSystem
		return
	}

	delete(podGroup.Annotations, TopologySourceAnnotationKey)
}

func isTopologyConstraintEmpty(constraint kaiv2alpha2.TopologyConstraint) bool {
	return constraint.Topology == "" &&
		constraint.PreferredTopologyLevel == "" &&
		constraint.RequiredTopologyLevel == ""
}
