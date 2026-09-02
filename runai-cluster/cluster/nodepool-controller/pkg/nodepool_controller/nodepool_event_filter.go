/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"
	"strings"

	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/utils"

	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func ControllerPredicateFuncs() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return filterUpdateEvents(e.ObjectOld, e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return true
		},
	}
}

func filterUpdateEvents(objectOld, objectNew client.Object) bool {
	switch objectNew.(type) {
	case *monitorv1.ServiceMonitor:
		return true
	case *kaiv1.SchedulingShard:
		return true
	case *corev1.Node:
		return common.FilterNodeUpdatesForNPController(objectOld.(*corev1.Node), objectNew.(*corev1.Node))
	case *v1alpha1.NodePool:
		return true
	case *v1alpha1.Project:
		// only reconcile when the project's nodepool references change - ignore unrelated updates
		return !getProjectNodePoolRefs(objectOld.(*v1alpha1.Project)).Equal(getProjectNodePoolRefs(objectNew.(*v1alpha1.Project)))
	}
	return false
}

func getProjectNodePoolRefs(project *v1alpha1.Project) sets.Set[string] {
	refs := sets.New[string]()
	for _, queue := range project.Spec.Queues {
		refs.Insert(queue.Nodepool)
	}
	return refs
}

// projectWatchObject returns the org-unit CRD the controller watches for nodepool-reference changes.
func projectWatchObject() client.Object {
	return &v1alpha1.Project{}
}

// MapProjectToNodePoolEvents reconciles all nodepools that are being deleted whenever a project's
// nodepool references change. The project object itself can't tell us which references were removed
// (map functions only see a single object), and project references only affect deleting nodepools,
// so all of them are enqueued.
func (npc *NodePoolController) MapProjectToNodePoolEvents(ctx context.Context, object client.Object) (requests []reconcile.Request) {
	deletingNodePools, err := npc.listNodePoolsWithFieldSelector(ctx,
		fields.OneTermEqualSelector(common.IsDeletingPhaseField, "true"))
	if err != nil {
		// the reconcile loop of a deleting nodepool requeues with an error while it is blocked,
		// so a missed event here is eventually recovered
		log.Error().Msgf("Failed listing deleting nodepools for project <%v> event, err: %v",
			object.GetName(), err.Error())
		return requests
	}

	for _, nodePool := range deletingNodePools.Items {
		appendNodePoolNameToRequests(&requests, nodePool.Name)
	}
	return requests
}

func (npc *NodePoolController) MapNodeToNodePoolEvent(ctx context.Context, object client.Object) (requests []reconcile.Request) {
	node, ok := object.(*corev1.Node)
	if !ok {
		log.Error().Msgf("Cannot convert object to *corev1.Node: %v", object)
		return requests
	}

	nodePoolName := npc.getNodePoolNameForNode(node)

	if nodePoolName == config.Get().ExcludedNodepoolName {
		return requests
	}

	nodePool := &v1alpha1.NodePool{}
	err := npc.Client.Get(ctx, types.NamespacedName{Name: nodePoolName}, nodePool)
	if err != nil {
		log.Info().Msgf("Warning - Node pool <%v> not found for node <%v>, using default instead, error: %v",
			nodePoolName, node.Name, err.Error())
		nodePoolName = config.Get().DefaultNodepoolName
	}

	log.Debug().Msgf("Got node pool <%v> for node <%v>", nodePoolName, node.Name)
	appendNodePoolNameToRequests(&requests, nodePoolName)

	requests = append(requests, npc.mapNodeToOtherDestNodePools(ctx, node, nodePoolName)...)

	return requests
}

// MapNrtToNodePoolEvent maps a NodeResourceTopology change to its node's nodepool(s) so the
// NRT-health condition is re-evaluated. The NRT object is named after the node.
func (npc *NodePoolController) MapNrtToNodePoolEvent(ctx context.Context, object client.Object) []reconcile.Request {
	node := &corev1.Node{}
	if err := npc.Client.Get(ctx, types.NamespacedName{Name: object.GetName()}, node); err != nil {
		log.Info().Msgf("Warning - node <%v> not found for NRT event, error: %v", object.GetName(), err.Error())
		return nil
	}
	return npc.MapNodeToNodePoolEvent(ctx, node)
}

func (npc *NodePoolController) getNodePoolNameForNode(node *corev1.Node) (nodePoolName string) {
	nodePoolName = utils.GetNodePoolNameFromLabels(node.Labels)
	return nodePoolName
}

// mapNodeToOtherDestNodePools maps the node to its destination node pools if it is unschedulable by us,
// by searching for node pools that match the node's labels;
// or if it is not unschedulable by us, it checks if the node appears in any node pool's status message.
func (npc *NodePoolController) mapNodeToOtherDestNodePools(ctx context.Context, node *corev1.Node,
	nodePoolAlreadyInRequests string) []reconcile.Request {
	_, ok := node.Labels[config.Get().UnschedulableLabelKey]
	isNodeUnschedulableByUs := ok && node.Spec.Unschedulable

	nodeNameWithSuffixes := nodeNameWithSuffixesToSearchFor(node.Name)

	nodePools := &v1alpha1.NodePoolList{}
	err := npc.Client.List(ctx, nodePools)
	if err != nil {
		log.Error().Msgf("Failed to list node pools, error: %v", err)
		return []reconcile.Request{}
	}

	requests := []reconcile.Request{}
	for _, nodePool := range nodePools.Items {
		if nodePool.Name == nodePoolAlreadyInRequests {
			continue
		}

		// if the node name appears in the node pool status message, and the node is now not unschedulable by us,
		// we want to reconcile the node pool
		if !isNodeUnschedulableByUs && isNodeInNodepoolStatusMessage(nodeNameWithSuffixes, nodePool.Status) {
			log.Debug().Msgf("Node <%v> is not unschedulable by us, but it appears in node pool <%v> status message, reconciling node pool",
				node.Name, nodePool.Name)
			appendNodePoolNameToRequests(&requests, nodePool.Name)
			continue
		}

		// else - we want to continue checking only if the node is unschedulable by us
		if !isNodeUnschedulableByUs {
			continue
		}

		if nodePool.Spec.LabelKey == "" || nodePool.Spec.LabelValue == "" || nodePool.DeletionTimestamp != nil {
			continue
		}

		if node.Labels[nodePool.Spec.LabelKey] == nodePool.Spec.LabelValue {
			log.Debug().Msgf("Node <%v> is Unschedulable by us and matches label key <%v> and value <%v> of node pool <%v>",
				node.Name, nodePool.Spec.LabelKey, nodePool.Spec.LabelValue, nodePool.Name)
			appendNodePoolNameToRequests(&requests, nodePool.Name)
		}
	}

	if isNodeUnschedulableByUs && len(requests) == 0 {
		log.Debug().Msgf("Node <%v> is unschedulable by us but does not match any node pool labels, using default node pool <%v>",
			node.Name, config.Get().DefaultNodepoolName)
		appendNodePoolNameToRequests(&requests, config.Get().DefaultNodepoolName)
	}

	return requests
}

func isNodeInNodepoolStatusMessage(nodeNameWithSuffixes []string, nodePoolStatus v1alpha1.NodePoolStatus) bool {
	for _, nodeNameWithSuffix := range nodeNameWithSuffixes {
		if strings.Contains(nodePoolStatus.Message, nodeNameWithSuffix) {
			return true
		}
	}
	return false
}

func nodeNameWithSuffixesToSearchFor(nodeName string) []string {
	// the node name can appear in the status message with " ", "," or "." right after it.
	nodeNameInMessagePossibleSuffixes := []string{" ", ",", "."}
	nodeNamesWithSuffixes := make([]string, len(nodeNameInMessagePossibleSuffixes))
	for i, suffix := range nodeNameInMessagePossibleSuffixes {
		nodeNamesWithSuffixes[i] = nodeName + suffix
	}
	return nodeNamesWithSuffixes
}

func appendNodePoolNameToRequests(requests *[]reconcile.Request, nodePoolName string) {
	*requests = append(*requests, reconcile.Request{
		NamespacedName: types.NamespacedName{Name: nodePoolName}})
}
