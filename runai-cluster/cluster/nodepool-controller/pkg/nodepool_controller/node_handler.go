/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"
	"encoding/json"
	"strings"

	kaischedulerv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/utils"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	GPUNetworkAccelerationLabelKey = "gpuNetworkAccelerationLabelKey"

	MarkedUnschedulableByNodePoolControllerLabelValue = "true"
)

func (npc *NodePoolController) findAndChangeNodePoolForNode(ctx context.Context, node *corev1.Node,
	oldNodePoolName string, availableNodePools []v1alpha1.NodePool) (schedulable bool, err error) {
	return npc.changeNodePoolForNodeGeneric(ctx, node, "", oldNodePoolName, availableNodePools)
}

func (npc *NodePoolController) ChangeNodePoolForNode(ctx context.Context, node *corev1.Node,
	newNodePoolName, oldNodePoolName string) (schedulable bool, err error) {
	return npc.changeNodePoolForNodeGeneric(ctx, node, newNodePoolName, oldNodePoolName, []v1alpha1.NodePool{})
}

func (npc *NodePoolController) changeNodePoolForNodeGeneric(ctx context.Context, node *corev1.Node,
	newNodePoolName, oldNodePoolName string, availableNodePools []v1alpha1.NodePool) (schedulable bool, err error) {
	schedulable, err = npc.validateNodeStatusWhenChangingNodePool(ctx, node, oldNodePoolName)
	if err != nil {
		log.Error().Msgf("Failed to change nodepools: failed validating node <%v> status when changing nodepools from <%v>, err: %v",
			node.Name, oldNodePoolName, err.Error())
		return
	}

	if schedulable {
		if newNodePoolName == "" {
			newNodePoolName = findMatchingNodePoolFromList(node, availableNodePools)
		}

		log.Info().Msgf("Changing nodepool assignment of node <%v> from: <%v> to: <%v>",
			node.Name, oldNodePoolName, newNodePoolName)
		err = npc.updateNodeWithNodePoolNameIfNeeded(ctx, node, newNodePoolName)
	}
	return
}

func (npc *NodePoolController) validateDefaultNodePoolLabel(ctx context.Context, node *corev1.Node) error {
	return npc.updateNodeWithNodePoolNameIfNeeded(ctx, node, config.Get().DefaultNodepoolName)
}

func (npc *NodePoolController) updateNodeWithNodePoolNameIfNeeded(ctx context.Context, node *corev1.Node, nodePoolName string) error {
	labelsToAdd := map[string]string{}
	labelsToRemove := []string{}

	nodePoolLabelKey := config.Get().NodePoolNameLabel
	foundNodePoolNameLabelValue, found := node.Labels[nodePoolLabelKey]
	if nodePoolName == config.Get().DefaultNodepoolName {
		if !found {
			return nil
		}
		labelsToRemove = append(labelsToRemove, nodePoolLabelKey)
	} else {
		if found && foundNodePoolNameLabelValue == nodePoolName {
			return nil
		}
		labelsToAdd[nodePoolLabelKey] = nodePoolName
	}

	err := npc.UpdateNodeLabels(ctx, node, labelsToAdd, labelsToRemove...)
	if err != nil {
		log.Error().Msgf("Failed updating node <%v> with nodepool name <%v>, err: %v",
			node.Name, nodePoolName, err.Error())
		return err
	}

	log.Info().Msgf("Successfully updated node <%v> with nodepool name <%v>",
		node.Name, nodePoolName)
	return nil
}

func (npc *NodePoolController) reconcileNodeAnnotations(ctx context.Context, node *corev1.Node, nodePool *v1alpha1.NodePool) error {
	nvlink := nvlinkAnnotationValue(nodePool)
	policy, err := npc.getTopologyManagerPolicyFromNodeNRT(ctx, node, nodePool)
	if err != nil {
		return err
	}

	desired := map[string]any{
		GPUNetworkAccelerationLabelKey:           nilIfEmpty(nvlink),
		v1alpha1.AnnotationTopologyManagerPolicy: nilIfEmpty(policy),
	}
	if annotationsUpToDate(node, desired) {
		return nil
	}
	return npc.UpdateAnnotations(ctx, node, desired)
}

func nvlinkAnnotationValue(nodePool *v1alpha1.NodePool) string {
	if getGPUNetworkAccelerationDetection(nodePool) == v1alpha1.AutoGPUNetworkAccelerationDetection {
		return getGPUNetworkAccelerationLabelKey(nodePool)
	}
	return ""
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func annotationsUpToDate(node *corev1.Node, desired map[string]any) bool {
	for key, wanted := range desired {
		current, found := node.Annotations[key]
		if wanted == nil {
			if found {
				return false
			}
			continue
		}
		if !found || current != wanted {
			return false
		}
	}
	return true
}

func (npc *NodePoolController) validateNodeLabeledWithUnschedulable(
	ctx context.Context, node *corev1.Node, unschedulable bool) error {
	if unschedulable {
		_, found := node.Labels[config.Get().UnschedulableLabelKey]
		if found {
			return nil
		}

		return npc.UpdateNodeLabels(ctx, node, map[string]string{config.Get().UnschedulableLabelKey: MarkedUnschedulableByNodePoolControllerLabelValue})
	}

	_, found := node.Labels[config.Get().UnschedulableLabelKey]
	if !found {
		return nil
	}

	return npc.UpdateNodeLabels(ctx, node, map[string]string{}, config.Get().UnschedulableLabelKey)
}

// handleNodesOfNodePool - implements required logic for nodes that are part of the nodepool
// 1. Node annotations (NVLink, Topology Manager policy)
// 2. NRT-health condition
func (npc *NodePoolController) handleNodesOfNodePool(ctx context.Context,
	nodePool *v1alpha1.NodePool, alteredNodes []*corev1.Node) ([]*corev1.Node, error) {
	nodes := npc.filterNodePoolNodes(alteredNodes, nodePool)

	if len(nodes) == 0 {
		return []*corev1.Node{}, nil
	}
	var err error

	resultNodes := []*corev1.Node{}
	for i := range nodes {
		node := nodes[i]
		innerErr := npc.reconcileNodeAnnotations(ctx, node, nodePool)
		err = utils.AppendErrIfNotNil(err, innerErr)

		innerErr = npc.reconcileNrtHealthCondition(ctx, node, nodePool)
		err = utils.AppendErrIfNotNil(err, innerErr)

		resultNodes = append(resultNodes, node)
	}

	return resultNodes, err
}

// hasMismatchNodeAgainstTopology
// Validates if the node's labels and NodeLevels of topology have a mismatch.
func (npc *NodePoolController) hasMismatchNodeAgainstTopology(node *corev1.Node, topology kaischedulerv1alpha1.Topology) bool {
	// Extract NodeLevels from the topology
	nodeLevels := npc.getNodeLevelsFromTopology(topology)

	// Validate node labels against NodeLevels
	missingNodeLabels := npc.getMissingNodeLabelsAgainstNodeLevels(node, nodeLevels)

	if len(missingNodeLabels) > 0 {
		log.Info().Msgf("Node <%v> has topology mismatch with topology <%v>, following labels are missing <%s>", node.Name, topology.Name, strings.Join(missingNodeLabels, ", "))
	}

	return len(missingNodeLabels) > 0
}

// getNodeLevelsFromTopology extracts NodeLevels from a topology resource
// This is a placeholder function that should be implemented when topology types are available
func (npc *NodePoolController) getNodeLevelsFromTopology(topology kaischedulerv1alpha1.Topology) []string {
	labels := []string{}
	for _, level := range topology.Spec.Levels {
		labels = append(labels, level.NodeLabel)
	}
	return labels
}

// getMissingNodeLabelsAgainstNodeLevels gets the NodeLevels that are missing on node labels
func (npc *NodePoolController) getMissingNodeLabelsAgainstNodeLevels(node *corev1.Node, topologyLevels []string) []string {
	missingLabels := []string{}
	nodeLabels := node.GetLabels()
	for _, nodeLevel := range topologyLevels {
		_, ok := nodeLabels[nodeLevel]
		if !ok {
			missingLabels = append(missingLabels, nodeLevel)
		}
	}
	return missingLabels
}

func (npc *NodePoolController) getTopologyMismatchNodes(ctx context.Context,
	nodePool *v1alpha1.NodePool, nodes []*corev1.Node) ([]*corev1.Node, error) {
	topology, fetchErr := npc.getTopologyByName(ctx, nodePool.Spec.PreferredNetworkTopologyName)
	if fetchErr != nil {
		return []*corev1.Node{}, fetchErr
	}
	if topology == nil {
		return []*corev1.Node{}, nil
	}
	mismatchNodes := []*corev1.Node{}
	for _, node := range nodes {
		if npc.hasMismatchNodeAgainstTopology(node, *topology) {
			mismatchNodes = append(mismatchNodes, node)
		}
	}
	return mismatchNodes, nil
}

func (npc *NodePoolController) patchNodeUnschedulable(ctx context.Context, node *corev1.Node, unschedulable bool) (err error) {
	patchBytes, err := updateNodeUnschedulablePatchBytes(unschedulable)
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - remove label from node <%v>, err: %v", node.Name, err)
		return err
	}

	return npc.patchObject(ctx, node, patchBytes)
}

func (npc *NodePoolController) UpdateAnnotations(ctx context.Context, node client.Object, annotations map[string]any) error {
	metadata, err := json.Marshal(map[string]interface{}{"metadata": map[string]interface{}{
		"annotations": annotations}})
	if err != nil {
		return err
	}
	err = npc.patchObject(ctx, node, metadata)
	if err != nil {
		log.Error().Msgf("Failed updating annotations, annotations <%v>, for node <%s> err: %v",
			annotations, node.GetName(), err.Error())
		return err
	}
	return nil
}

func (npc *NodePoolController) UpdateNodeLabels(ctx context.Context, node *corev1.Node,
	labelsToAdd map[string]string, labelsToRemove ...string) (err error) {
	patchBytes, err := updateNodeLabelsPatchBytes(labelsToAdd, labelsToRemove...)
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - update labels of node <%s>, err: %v", node.Name, err)
		return err
	}

	err = npc.patchObject(ctx, node, patchBytes)
	if err != nil {
		log.Error().Msgf("Failed updating labels, adding: <%+v>, removing: <%v>, of node <%s>, err: %v",
			labelsToAdd, labelsToRemove, node.Name, err.Error())
		return err
	}

	log.Debug().Msgf("Successfully added labels <%+v>, removed labels <%v> from node <%s>",
		labelsToAdd, labelsToRemove, node.Name)
	return nil
}

func (npc *NodePoolController) patchObject(ctx context.Context, obj client.Object, patchBytes []byte) error {
	patch := client.RawPatch(types.MergePatchType, patchBytes)
	return npc.Client.Patch(ctx, obj, patch)
}

func findMatchingNodePoolFromList(node *corev1.Node, allNodePools []v1alpha1.NodePool) (foundNodePoolName string) {
	for _, nodePool := range allNodePools {
		nodeLabelValue, found := node.Labels[nodePool.Spec.LabelKey]
		if !found || nodeLabelValue != nodePool.Spec.LabelValue {
			continue
		}

		log.Debug().Msgf("For node <%v>, found label key <%v> and value <%v> of nodepool <%v>",
			node.Name, nodePool.Spec.LabelKey, nodePool.Spec.LabelValue, nodePool.Name)
		return nodePool.Name
	}

	log.Debug().Msgf("Node's <%v> labels dont match any nodepool; using default nodepool: <%v>",
		node.Name, config.Get().DefaultNodepoolName)
	return config.Get().DefaultNodepoolName
}

func updateNodeLabelsPatchBytes(labelsToAdd map[string]string, labelsToRemove ...string) ([]byte, error) {
	labels := map[string]interface{}{}
	for key, value := range labelsToAdd {
		labels[key] = value
	}
	for _, labelToRemove := range labelsToRemove {
		labels[labelToRemove] = nil
	}

	return json.Marshal(map[string]interface{}{"metadata": map[string]interface{}{
		"labels": labels}})
}

func updateNodeUnschedulablePatchBytes(unschedulable bool) ([]byte, error) {
	labels := map[string]interface{}{}
	if unschedulable {
		labels[config.Get().UnschedulableLabelKey] = MarkedUnschedulableByNodePoolControllerLabelValue
	} else {
		labels[config.Get().UnschedulableLabelKey] = nil
	}

	return json.Marshal(map[string]interface{}{"spec": map[string]interface{}{
		"unschedulable": unschedulable},
		"metadata": map[string]interface{}{
			"labels": labels}})
}
