/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"strconv"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"k8s.io/utils/ptr"
)

type NodepoolAdapter v1alpha1.NodePool
type NodepoolsAdapter []v1alpha1.NodePool

type NodepoolConditionAdapter v1alpha1.NodePoolCondition
type NodepoolConditionsAdapter []v1alpha1.NodePoolCondition

// NewNodepoolAdapter creates a new NodepoolAdapter from a kai.resources Nodepool
func NewNodepoolAdapter(nodepool *v1alpha1.NodePool) *NodepoolAdapter {
	adapter := NodepoolAdapter(*nodepool)
	return &adapter
}

// NewNodepoolsAdapter creates a new NodepoolsAdapter from a kai.resources Nodepool
func NewNodepoolsAdapter(nodepools []v1alpha1.NodePool) *NodepoolsAdapter {
	adapter := NodepoolsAdapter(nodepools)
	return &adapter
}

// NewNodepoolConditionAdapter creates a new NodepoolConditionAdapter from a kai.resources Nodepool
func NewNodepoolConditionAdapter(condition *v1alpha1.NodePoolCondition) *NodepoolConditionAdapter {
	adapter := NodepoolConditionAdapter(*condition)
	return &adapter
}

// NewNodepoolConditionsAdapter creates a new NodepoolConditionsAdapter from a kai.resources Nodepool
func NewNodepoolConditionsAdapter(condition []v1alpha1.NodePoolCondition) *NodepoolConditionsAdapter {
	adapter := NodepoolConditionsAdapter(condition)
	return &adapter
}

// ToServerScheme converts a slice of Nodepool to a slice of NodepoolSyncFields
func (ntas *NodepoolsAdapter) ToServerScheme() ([]openapi.NodepoolSyncFields, error) {
	nodepools := make([]openapi.NodepoolSyncFields, 0, len(*ntas))

	for _, nodepool := range *ntas {
		nodepoolSync, err := NewNodepoolAdapter(&nodepool).
			ToServerScheme()
		if err != nil {
			return nil, err
		}
		nodepools = append(nodepools, *nodepoolSync)
	}

	return nodepools, nil
}

func (condition *NodepoolConditionAdapter) ToServerScheme(nodes []v1alpha1.NodeInNodePool) *openapi.NodepoolConditionDetails {
	relatedNodesNames := []string{}
	if condition.Type == v1alpha1.NodeTopologyMismatch {
		for _, node := range nodes {
			if node.TopologyMismatch {
				relatedNodesNames = append(relatedNodesNames, node.Name)
			}
		}
	}
	return &openapi.NodepoolConditionDetails{
		Type:    string(condition.Type),
		Reason:  condition.Reason,
		Message: ptr.To(condition.Message),
		Status:  ptr.To(string(condition.Status)),
		Nodes:   relatedNodesNames,
	}
}

func (conditions *NodepoolConditionsAdapter) ToServerScheme(nodes []v1alpha1.NodeInNodePool) []openapi.NodepoolConditionDetails {
	conditionsSync := make([]openapi.NodepoolConditionDetails, 0, len(*conditions))

	for _, condition := range *conditions {
		conditionSync := NewNodepoolConditionAdapter(&condition).
			ToServerScheme(nodes)
		conditionsSync = append(conditionsSync, *conditionSync)
	}

	return conditionsSync
}

func (nodePool *NodepoolAdapter) ToServerScheme() (*openapi.NodepoolSyncFields, error) {
	nodepoolStatus := nodePool.Status

	phase, err := openapi.NewNodepoolPhaseFromValue(string(nodepoolStatus.Phase))
	if err != nil {
		return nil, err
	}

	nodepoolConditionsAdapter := NewNodepoolConditionsAdapter(nodePool.Status.Conditions)
	npSyncFields := &openapi.NodepoolSyncFields{
		Name:         nodePool.Name,
		Phase:        phase,
		PhaseMessage: ptr.To(nodePool.Status.Message),
		Status: &openapi.NodepoolSyncFieldsStatus{
			Nodes:                          getNodeNames(nodepoolStatus.Nodes),
			GpuNetworkAccelerationDetected: ptr.To(nodePool.gpuNetworkAccelerationDetected()),
			Conditions:                     nodepoolConditionsAdapter.ToServerScheme(nodepoolStatus.Nodes),
		},
	}
	nodePool.setSchedulingConfigurationStatus(npSyncFields.Status)

	return npSyncFields, nil
}

func (nodePool *NodepoolAdapter) ToDeleteRequest() *openapi.NodepoolSyncFields {
	return &openapi.NodepoolSyncFields{
		Name:  nodePool.Name,
		Phase: ptr.To(openapi.NODEPOOLPHASE_DELETED),
	}
}

func (nodePool *NodepoolAdapter) gpuNetworkAccelerationDetected() bool {
	detected, err := strconv.ParseBool(nodePool.Annotations[v1alpha1.AnnotationGPUNetworkAccelerationDetected])
	if err != nil {
		return false
	}
	return detected
}

func getNodeNames(nodesInNodePool []v1alpha1.NodeInNodePool) []string {
	var nodeNames []string
	if len(nodesInNodePool) == 0 {
		return nodeNames
	}

	for _, node := range nodesInNodePool {
		nodeNames = append(nodeNames, node.Name)
	}

	return nodeNames
}

// setSchedulingConfigurationStatus reports the minimum guaranteed runtime. On the
// kai.resources CRD it lives under schedulingShardConfig.minRuntime, which splits it into
// a preempt and a reclaim value; the migration writes the single legacy value to both, so
// the preempt value is read back here.
func (nodePool *NodepoolAdapter) setSchedulingConfigurationStatus(apiStatus *openapi.NodepoolSyncFieldsStatus) {
	shardConfig := nodePool.Spec.SchedulingShardConfig
	if shardConfig == nil || shardConfig.MinRuntime == nil || shardConfig.MinRuntime.PreemptMinRuntime == nil {
		return
	}

	minGuaranteedRuntime := *shardConfig.MinRuntime.PreemptMinRuntime
	if minGuaranteedRuntime == "" {
		return
	}

	apiStatus.SchedulingConfiguration = *openapi.NewNullableSchedulingConfiguration(
		&openapi.SchedulingConfiguration{
			MinGuaranteedRuntime: *openapi.NewNullableString(&minGuaranteedRuntime),
		},
	)
}
