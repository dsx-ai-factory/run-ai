/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"encoding/json"
	"reflect"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/metrics"
	multierror "github.com/hashicorp/go-multierror"
	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config"
	kainodepool "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// numaPluginName is the scheduler plugin key that enables NUMA-aware scheduling.
const numaPluginName = "numa"

type KaiNodePoolsHandler struct {
	*CommonHandler
}

type KaiNodePoolObjectInspector bool

func NewKaiNodePoolsHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	commonHandler, err := NewCommonHandler(runaiBackendClient, runaiClient, cachedClient,
		&kainodepool.NodePool{}, KaiNodePoolObjectInspector(true), HandlingOptions{})
	if err != nil {
		return nil, err
	}

	return &KaiNodePoolsHandler{
		CommonHandler: commonHandler,
	}, nil
}

func (h *KaiNodePoolsHandler) Handle() error {
	log.Debug().Msgf("Starting to sync kai nodepools")

	backendNodePools, err := h.getNodePools()
	if err != nil {
		return err
	}

	clusterNodePools, nodePoolNamesToDelete, err := extractKaiNodePools(backendNodePools)
	if err != nil {
		// less likely, because it's only json marshaling, but if err occurred - deleted the backend-deleted nodepools
		_, innerErr := h.DeleteObjects(nodePoolNamesToDelete)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}

		log.Error().Msgf("Failed to handle kai NodePools, error: %s", err.Error())
		return err
	}
	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionCreateOrUpdate, len(clusterNodePools))
	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionDelete, len(nodePoolNamesToDelete))

	return h.ReconcileObjects(clusterNodePools, false)
}

func (h *KaiNodePoolsHandler) MetricLabel() string {
	return "node_pools_kai"
}

func (h *KaiNodePoolsHandler) getNodePools() ([]openapi.Nodepool, error) {
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

// extractKaiNodePools
// returns:
// * clusterNodePools
// * nodePoolNamesToDelete
// * error
func extractKaiNodePools(nodepools []openapi.Nodepool) ([]client.Object, []types.NamespacedName, error) {
	return extractNodePoolsWithConverter(nodepools, func(nodepool *openapi.Nodepool) (client.Object, error) {
		return convertKaiNodePoolFields(nodepool)
	})
}

func convertKaiNodePoolFields(nodepoolModel *openapi.Nodepool) (*kainodepool.NodePool, error) {
	crdNodePool := &kainodepool.NodePool{}
	crdNodePool.Name = nodepoolModel.Name
	crdNodePool.Spec = kainodepool.NodePoolSpec{
		LabelKey:                     nodepoolModel.LabelKey,
		LabelValue:                   nodepoolModel.LabelValue,
		PreferredNetworkTopologyName: ptr.Deref(nodepoolModel.NetworkTopologyName.Get(), ""),
	}

	crdNodePool.Annotations = kaiNodePoolGpuNetworkAccelerationAnnotations(nodepoolModel)

	schedulingConfiguration := nodepoolModel.SchedulingConfiguration.Get()
	crdNodePool.Spec.SchedulingShardConfig = convertKaiSchedulingShardConfig(schedulingConfiguration)

	return crdNodePool, nil
}

func kaiNodePoolGpuNetworkAccelerationAnnotations(nodepoolModel *openapi.Nodepool) map[string]string {
	annotations := map[string]string{
		kainodepool.AnnotationGPUNetworkAccelerationDetection: string(convertNullableNodepoolGPUNetworkAccelerationDetection(nodepoolModel.GpuNetworkAccelerationDetection)),
	}

	// Left unset (not defaulted) when empty: the controller's detection logic is
	// expected to apply the nvidia.com/gpu.clique fallback itself, the same way it
	// does today for the legacy CRD's spec field. Defaulting here would only cover
	// CP-managed nodepools, not ones applied directly without control-plane/agent.
	if labelKey := ptr.Deref(nodepoolModel.GpuNetworkAccelerationLabelKey, ""); labelKey != "" {
		annotations[kainodepool.AnnotationGPUNetworkAccelerationLabelKey] = labelKey
	}

	return annotations
}

func convertKaiSchedulingShardConfig(schedulingConfiguration *openapi.SchedulingConfiguration) *kainodepool.SchedulingShardConfig {
	if schedulingConfiguration == nil {
		return nil
	}

	shardConfig := &kainodepool.SchedulingShardConfig{
		PlacementStrategy:  convertKaiPlacementStrategy(schedulingConfiguration),
		MinRuntime:         convertKaiMinRuntime(GetFromNullable(schedulingConfiguration.MinGuaranteedRuntime, "")),
		Plugins:            convertKaiPlugins(schedulingConfiguration),
		TimeBasedFairShare: convertKaiTimeBasedFairShare(schedulingConfiguration.TimeBasedFairShare),
	}

	return shardConfig
}

func convertKaiPlacementStrategy(schedulingConfiguration *openapi.SchedulingConfiguration) *kaiv1.PlacementStrategy {
	gpuNodeStrategy, _ := kaiGpuPlacementStrategies(schedulingConfiguration)
	cpuNodeStrategy := CPUDefault
	if cpuPods := schedulingConfiguration.CpuPodsPlacementStrategy.Get(); cpuPods != nil {
		if node := cpuPods.Node.Get(); node != nil {
			cpuNodeStrategy = *node
		}
	}

	return &kaiv1.PlacementStrategy{
		GPU: ptr.To(string(gpuNodeStrategy)),
		CPU: ptr.To(string(cpuNodeStrategy)),
	}
}

// kaiGpuPlacementStrategies returns the node-level and device-level GPU placement
// strategies, defaulting each independently the same way the legacy handler does.
func kaiGpuPlacementStrategies(schedulingConfiguration *openapi.SchedulingConfiguration) (nodeStrategy, deviceStrategy openapi.PlacementStrategy) {
	nodeStrategy = GPUDefault
	deviceStrategy = ""

	gpuPods := schedulingConfiguration.GpuPodsPlacementStrategy.Get()
	if gpuPods == nil {
		return nodeStrategy, deviceStrategy
	}

	if node := gpuPods.Node.Get(); node != nil {
		nodeStrategy = *node
	}
	if device := gpuPods.Device.Get(); device != nil {
		deviceStrategy = *device
	}

	return nodeStrategy, deviceStrategy
}

func convertKaiMinRuntime(minGuaranteedRuntime string) *kaiv1.MinRuntime {
	if minGuaranteedRuntime == "" {
		return nil
	}

	return &kaiv1.MinRuntime{
		PreemptMinRuntime: ptr.To(minGuaranteedRuntime),
		ReclaimMinRuntime: ptr.To(minGuaranteedRuntime),
	}
}

// convertKaiPlugins builds the whole plugin override map.
func convertKaiPlugins(schedulingConfiguration *openapi.SchedulingConfiguration) map[string]kaiv1.PluginConfig {
	plugins := convertKaiDevicePlacementPlugins(schedulingConfiguration)

	if GetFromNullable(schedulingConfiguration.NumaAwareScheduling, false) {
		if plugins == nil {
			plugins = map[string]kaiv1.PluginConfig{}
		}
		plugins[numaPluginName] = kaiv1.PluginConfig{Enabled: ptr.To(true)}
	}

	return plugins
}

// convertKaiDevicePlacementPlugins sets the plugins the KAI scheduler otherwise
// derives from the node-level GPU strategy alone (see getDevicePlacementPluginOverrides
// in the nodepool-controller), explicitly reflecting the device-level strategy
// whenever one was requested.
func convertKaiDevicePlacementPlugins(schedulingConfiguration *openapi.SchedulingConfiguration) map[string]kaiv1.PluginConfig {
	_, gpuDeviceStrategy := kaiGpuPlacementStrategies(schedulingConfiguration)

	if gpuDeviceStrategy == "" {
		return nil
	}

	isDeviceBinpack := gpuDeviceStrategy == openapi.PLACEMENTSTRATEGY_BINPACK
	return map[string]kaiv1.PluginConfig{
		"gpupack":         {Enabled: ptr.To(isDeviceBinpack)},
		"gpuspread":       {Enabled: ptr.To(!isDeviceBinpack)},
		"gpusharingorder": {Enabled: ptr.To(isDeviceBinpack)},
	}
}

func convertKaiTimeBasedFairShare(timeBasedFairShareField openapi.NullableTimeBasedFairShare) *kainodepool.TimeBasedFairShare {
	timeBasedFairShare := timeBasedFairShareField.Get()
	if timeBasedFairShare == nil || !GetFromNullable(timeBasedFairShare.Enabled, false) {
		return nil
	}

	tbfs := &kainodepool.TimeBasedFairShare{
		Enabled:               ptr.To(true),
		HistoricalUsageWeight: ptr.To(GetFromNullable(timeBasedFairShare.HistoricalUsageWeight, DefaultTimeBasedFairShareHistoricalUsageWeight)),
		HalfLifePeriod:        ptr.To(GetFromNullable(timeBasedFairShare.DecayHalfLife, DefaultTimeBasedFairShareDecayHalfLife)),
		Window: &kainodepool.FairShareWindow{
			Type: ptr.To(string(GetFromNullable(timeBasedFairShare.WindowType, DefaultTimeBasedFairShareWindowType))),
			Size: ptr.To(GetFromNullable(timeBasedFairShare.WindowDuration, DefaultTimeBasedFairShareWindowDuration)),
		},
	}

	if tumblingWindowStartTime := timeBasedFairShare.TumblingWindowStartTime.Get(); tumblingWindowStartTime != nil {
		tbfs.Window.TumblingStartTime = ptr.To(metav1.NewTime(*tumblingWindowStartTime))
	}

	return tbfs
}

func (k KaiNodePoolObjectInspector) DeepCopySpec(from, to client.Object) {
	fromNodePool, ok := from.(*kainodepool.NodePool)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return
	}
	toNodePool, ok := to.(*kainodepool.NodePool)
	if !ok {
		return
	}

	fromNodePool.Spec.DeepCopyInto(&toNodePool.Spec)
	toNodePool.Annotations = make(map[string]string, len(fromNodePool.Annotations))
	for k, v := range fromNodePool.Annotations {
		toNodePool.Annotations[k] = v
	}
}

func (k KaiNodePoolObjectInspector) IsUpToDate(left, right client.Object) bool {
	leftNodePool, ok := left.(*kainodepool.NodePool)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return false
	}
	rightNodePool, ok := right.(*kainodepool.NodePool)
	if !ok {
		return false
	}

	leftSpecJson, err := json.Marshal(leftNodePool.Spec)
	if err != nil {
		log.Error().Msgf("Failed to json marshal kai nodepool spec, error: %s", err.Error())
		return false
	}
	rightSpecJson, err := json.Marshal(rightNodePool.Spec)
	if err != nil {
		log.Error().Msgf("Failed to json marshal kai nodepool spec, error: %s", err.Error())
		return false
	}

	return reflect.DeepEqual(leftSpecJson, rightSpecJson) &&
		reflect.DeepEqual(leftNodePool.Annotations, rightNodePool.Annotations)
}
