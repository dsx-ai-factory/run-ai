/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package podgroup

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/config"
	nodepoolutils "github.com/dsx-ai-factory/run-ai/runai-cluster/common/node-pool-utils/utils"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *PodGroupReconciler) MapNodePoolToPodGroupEvent(_ context.Context, object client.Object) (requests []reconcile.Request) {
	// Only the name is needed, so this works for a node pool of either CRD without knowing
	// which one is being watched.
	nodePoolName := object.GetName()

	matchNodePoolRequirement, err := nodepoolutils.GetNodePoolLabelSelector(nodePoolName, config.Config().NodePoolLabelKey, config.Config().DefaultNodepoolName)
	if err != nil {
		log.Error().Msgf("Failed getting node pool selector for node pool <%s>, err: <%s>", nodePoolName, err.Error())
		return requests
	}

	podGroups, err := r.listPodGroupsWithSelector(*matchNodePoolRequirement)
	if err != nil {
		return requests
	}

	for _, podGroup := range podGroups.Items {
		podGroupNamespacedName := types.NamespacedName{
			Name:      podGroup.Name,
			Namespace: podGroup.Namespace,
		}
		log.Info().Msgf("For Node Pool <%s> of phase <%s>, found pod group <%v> - triggering Reconcile",
			nodePoolName, nodePoolPhase(object), podGroupNamespacedName)

		requests = append(requests, reconcile.Request{
			NamespacedName: podGroupNamespacedName})
	}

	return requests
}

func (r *PodGroupReconciler) FilterPodGroupControllerEvents(object client.Object) bool {
	switch typedObject := object.(type) {
	case *v1alpha1.NodePool:
		return isNodePoolPhaseRelevantForController(string(typedObject.Status.Phase))
	case *kaiv2alpha2.PodGroup:
		return true
	}

	return false
}

func (r *PodGroupReconciler) listPodGroupsWithSelector(labelRequirement labels.Requirement) (*kaiv2alpha2.PodGroupList, error) {
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(labelRequirement)

	podGroups := &kaiv2alpha2.PodGroupList{}

	err := r.cachedClient.List(
		context.Background(),
		podGroups,
		&client.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		log.Error().Msgf("Failed listing pod groups with label selector <%s>, err: <%s>",
			labelSelector.String(), err.Error())

		return nil, err
	}

	return podGroups, nil
}

// nodePoolPhase reads the phase off a KAI NodePool for logging. Returns an empty string for
// anything else.
func nodePoolPhase(object client.Object) string {
	switch typedObject := object.(type) {
	case *v1alpha1.NodePool:
		return string(typedObject.Status.Phase)
	}
	return ""
}

func isNodePoolPhaseRelevantForController(phase string) bool {
	// we are only interested in Reconcile if the node pool is Deleting or Unschedulable
	return phase == string(v1alpha1.NodePoolDeleting) || phase == string(v1alpha1.NodePoolUnschedulable)
}
