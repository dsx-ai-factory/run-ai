/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"encoding/json"
	"reflect"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/metrics"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NodePoolsHandler struct {
	*CommonHandler
}

type NodePoolObjectInspector bool

func NewNodePoolsHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	commonHandler, err := NewCommonHandler(runaiBackendClient, runaiClient, cachedClient,
		&v1alpha1.NodePool{}, NodePoolObjectInspector(true), HandlingOptions{})
	if err != nil {
		return nil, err
	}

	return &NodePoolsHandler{
		CommonHandler: commonHandler,
	}, nil
}

func (h *NodePoolsHandler) Handle() error {
	log.Debug().Msgf("Starting to sync nodepools")

	backendNodePools, err := h.getNodePools()
	if err != nil {
		return err
	}

	clusterNodePools, nodePoolNamesToDelete, err := extractNodePools(backendNodePools)
	if err != nil {
		// less likely, because it's only json marshaling, but if err occurred - deleted the backend-deleted nodepools
		_, innerErr := h.DeleteObjects(nodePoolNamesToDelete)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}

		log.Error().Msgf("Failed to handle NodePools, error: %s", err.Error())
		return err
	}
	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionCreateOrUpdate, len(clusterNodePools))
	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionDelete, len(nodePoolNamesToDelete))

	return h.ReconcileObjects(clusterNodePools, false)
}

func (h *NodePoolsHandler) MetricLabel() string {
	return "node_pools"
}

func (h *NodePoolsHandler) getNodePools() ([]openapi.Nodepool, error) {
	response, err := h.backendClient.ClusterServiceClient.GetNodePoolsByCluster(
		h.backendClient.Context(),
		config.Config().ClusterUuid)

	if err != nil {
		log.Error().Msgf("Failed to get Nodepools from cluster-service: err: %v", err)
		return []openapi.Nodepool{}, err
	}
	nodePoolsModels := response.GetNodepools()

	return nodePoolsModels, nil
}

// extractNodePools
// returns:
// * clusterNodePools
// * nodePoolNamesToDelete
// * error
func extractNodePools(nodepools []openapi.Nodepool) ([]client.Object, []types.NamespacedName, error) {
	return extractNodePoolsWithConverter(nodepools, func(nodepool *openapi.Nodepool) (client.Object, error) {
		return convertNodePoolFields(nodepool)
	})
}

func convertNodePoolFields(nodepoolModel *openapi.Nodepool) (*v1alpha1.NodePool, error) {
	crdNodePool := &v1alpha1.NodePool{}
	crdNodePool.Spec = v1alpha1.NodePoolSpec{
		LabelKey:                        nodepoolModel.LabelKey,
		LabelValue:                      nodepoolModel.LabelValue,
		GPUNetworkAccelerationDetection: ptr.To(convertNullableNodepoolGPUNetworkAccelerationDetection(nodepoolModel.GpuNetworkAccelerationDetection)),
		GPUNetworkAccelerationLabelKey:  ptr.Deref(nodepoolModel.GpuNetworkAccelerationLabelKey, ""),
		NetworkTopologyName:             ptr.Deref(nodepoolModel.NetworkTopologyName.Get(), ""),
	}
	setSchedulingConfigurationFields(crdNodePool, nodepoolModel.SchedulingConfiguration)
	crdNodePool.Name = nodepoolModel.Name
	return crdNodePool, nil
}

func setSchedulingConfigurationFields(crdNodePool *v1alpha1.NodePool, schedulingConfigurationFields openapi.NullableSchedulingConfiguration) {
	schedulingConfiguration := schedulingConfigurationFields.Get()
	if schedulingConfiguration == nil {
		crdNodePool.Spec.PlacementStrategy = v1alpha1.PlacementStrategy{
			Gpu:       string(GPUDefault),
			GpuDevice: string(GPUDefault),
			Cpu:       string(CPUDefault),
		}
		crdNodePool.Spec.MinGuaranteedRuntime = ""
		crdNodePool.Spec.NumaAwareScheduling = false
		return
	}

	gpuValue := GPUDefault
	gpuDeviceValue := GPUDefault
	cpuValue := CPUDefault

	if gpuPods := schedulingConfiguration.GpuPodsPlacementStrategy.Get(); gpuPods != nil {
		if node := gpuPods.Node.Get(); node != nil {
			gpuValue = *node
		}
		if device := gpuPods.Device.Get(); device != nil {
			gpuDeviceValue = *device
		}
	}
	if cpuPods := schedulingConfiguration.CpuPodsPlacementStrategy.Get(); cpuPods != nil {
		if node := cpuPods.Node.Get(); node != nil {
			cpuValue = *node
		}
	}

	crdNodePool.Spec.PlacementStrategy = v1alpha1.PlacementStrategy{
		Gpu:       string(gpuValue),
		GpuDevice: string(gpuDeviceValue),
		Cpu:       string(cpuValue),
	}
	crdNodePool.Spec.MinGuaranteedRuntime = GetFromNullable(schedulingConfiguration.MinGuaranteedRuntime, "")
	crdNodePool.Spec.NumaAwareScheduling = GetFromNullable(schedulingConfiguration.NumaAwareScheduling, false)

	setTimeBasedFairShareFields(crdNodePool, schedulingConfiguration.TimeBasedFairShare)
}

func setTimeBasedFairShareFields(crdNodePool *v1alpha1.NodePool, timeBasedFairShareFields openapi.NullableTimeBasedFairShare) {
	timeBasedFairShare := timeBasedFairShareFields.Get()
	if timeBasedFairShare == nil || !GetFromNullable(timeBasedFairShare.Enabled, false) {
		crdNodePool.Spec.TimeBasedFairShare = v1alpha1.TimeBasedFairShare{Enabled: false}
		return
	}

	crdNodePool.Spec.TimeBasedFairShare = v1alpha1.TimeBasedFairShare{
		Enabled: true,
	}

	crdNodePool.Spec.TimeBasedFairShare.HistoricalUsageWeight = ptr.To(GetFromNullable(timeBasedFairShare.HistoricalUsageWeight, DefaultTimeBasedFairShareHistoricalUsageWeight))
	crdNodePool.Spec.TimeBasedFairShare.DecayHalfLife = ptr.To(GetFromNullable(timeBasedFairShare.DecayHalfLife, DefaultTimeBasedFairShareDecayHalfLife))
	crdNodePool.Spec.TimeBasedFairShare.WindowDuration = ptr.To(GetFromNullable(timeBasedFairShare.WindowDuration, DefaultTimeBasedFairShareWindowDuration))
	windowTypeStr := GetFromNullable(timeBasedFairShare.WindowType, DefaultTimeBasedFairShareWindowType)
	crdNodePool.Spec.TimeBasedFairShare.WindowType = ptr.To(v1alpha1.TimeBasedFairShareWindowType(windowTypeStr))
	tumblingWindowStartTime := timeBasedFairShare.TumblingWindowStartTime.Get()
	if tumblingWindowStartTime != nil {
		crdNodePool.Spec.TimeBasedFairShare.TumblingWindowStartTime = ptr.To(metav1.NewTime(*tumblingWindowStartTime))
	}
}

func (n NodePoolObjectInspector) DeepCopySpec(from, to client.Object) {
	fromNodePool, ok := from.(*v1alpha1.NodePool)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return
	}
	toNodePool, ok := to.(*v1alpha1.NodePool)
	if !ok {
		return
	}

	fromNodePool.Spec.DeepCopyInto(&toNodePool.Spec)
}

func (n NodePoolObjectInspector) IsUpToDate(left, right client.Object) bool {
	leftNodePool, ok := left.(*v1alpha1.NodePool)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return false
	}
	rightNodePool, ok := right.(*v1alpha1.NodePool)
	if !ok {
		return false
	}

	leftSpecJson, err := json.Marshal(leftNodePool.Spec)
	if err != nil {
		log.Error().Msgf("Failed to json marshal nodepool spec, error: %s", err.Error())
		return false
	}
	rightSpecJson, err := json.Marshal(rightNodePool.Spec)
	if err != nil {
		log.Error().Msgf("Failed to json marshal nodepool spec, error: %s", err.Error())
		return false
	}

	return reflect.DeepEqual(leftSpecJson, rightSpecJson)
}
