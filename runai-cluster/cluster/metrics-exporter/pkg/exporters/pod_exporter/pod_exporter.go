/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/metrics_exporter/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	podAllocatedGPUsAnnotationName    = "runai-allocated-gpus"
	podAllocatedMigGPUsAnnotationName = "runai-allocated-mig-gpus"
	runaiOpportunisticAnnotationName  = "runai-opportunistic"
	podStatusAnnotationName           = "runai-calculated-status"
	allocatedMigDeviceAnnotationName  = "runai-mig-device"
	numberOfDevicesAnnotationName     = "gpu-fraction-num-devices"

	receivedResourceTypeAnnotationName = "received-resource-type"

	unknownDevice                 = "unknown"
	nodeSchedulerControlledDevice = "all"
	opportunisticPodGpuIndex      = "opportunistic"
)

type PodExporter struct {
	ctx    context.Context
	client client.Client
	pod    *v1.Pod
}

func NewPodExporter(ctx context.Context, k8sClient client.Client, pod *v1.Pod) *PodExporter {
	return &PodExporter{
		ctx:    ctx,
		client: k8sClient,
		pod:    pod,
	}
}

func (pe *PodExporter) GetName() string {
	return pe.pod.Name
}

func (pe *PodExporter) GetNamespace() string {
	return pe.pod.Namespace
}

func (pe *PodExporter) GetUUID() string {
	return string(pe.pod.UID)
}

func (pe *PodExporter) GetNodeName() string {
	return pe.pod.Spec.NodeName
}

func (pe *PodExporter) GetMigDeviceName() (string, error) {
	device, err := pe.getMIGDeviceInfo()
	if err != nil {
		return "", err
	}
	return device.Name, nil
}

func (pe *PodExporter) GetGpuIndex() (string, error) {
	if !pe.IsBound() {
		return "", nil
	}

	if pe.isMIG() {
		device, err := pe.getMIGDeviceInfo()
		if err != nil {
			return "", err
		}
		return device.GPUIndex, nil
	}

	if pe.isOpportunisticAllocation() {
		return opportunisticPodGpuIndex, nil
	}

	gpuIndex, err := pe.extractGPUIndex()
	if err != nil {
		return "", err
	}

	gpuIndex = stripCDIDevicePrefix(gpuIndex)

	if gpuIndex == nodeSchedulerControlledDevice {
		gpuIndex = unknownDevice
	}

	return gpuIndex, nil
}

// stripCDIDevicePrefix extracts GPU identifiers from CDI (Container Device Interface) device names.
// CDI format: "k8s.device-plugin.nvidia.com/gpu=GPU-<uuid>" → "GPU-<uuid>"
// Non-CDI values like "3" or "GPU-<uuid>" pass through unchanged.
// Handles comma-separated multi-device strings.
func stripCDIDevicePrefix(gpuIndex string) string {
	parts := strings.Split(gpuIndex, ",")
	for i, part := range parts {
		if idx := strings.LastIndex(part, "="); idx != -1 {
			parts[i] = part[idx+1:]
		}
	}
	return strings.Join(parts, ",")
}

func (pe *PodExporter) GetGpuInstanceId() (string, error) {
	if !pe.isMIG() {
		return "", nil
	}
	device, err := pe.getMIGDeviceInfo()
	if err != nil {
		return "", err
	}
	return strconv.Itoa(device.GpuInstanceId), nil
}

func (pe *PodExporter) GetAllocatedGpus() float64 {
	if !pe.isActive() {
		return float64(0)
	}

	gpus, err := utils.ParseFloatFromAnnotation(pe.pod.Annotations, podAllocatedGPUsAnnotationName)
	if err != nil {
		log.Printf("failed to get allocated gpus for pod %s/%s, error: %s",
			pe.GetNamespace(), pe.GetName(), err.Error())
		return float64(0)
	}
	migDevices, err := utils.ParseFloatFromAnnotation(pe.pod.Annotations, podAllocatedMigGPUsAnnotationName)
	if err != nil {
		log.Printf("failed to get allocated MIG devices for pod %s/%s, error: %s",
			pe.GetNamespace(), pe.GetName(), err.Error())
		return float64(0)
	}
	return gpus + migDevices
}

func (pe *PodExporter) GetAllocatedCpus() float64 {
	if !pe.isActive() {
		return float64(0)
	}

	cpus := float64(0)
	for _, container := range pe.pod.Spec.Containers {
		cpuRequest := container.Resources.Requests[v1.ResourceCPU]
		cpus += float64(cpuRequest.MilliValue())
	}
	return cpus
}

func (pe *PodExporter) GetAllocatedMemory() float64 {
	if !pe.isActive() {
		return float64(0)
	}
	memory := float64(0)
	for _, container := range pe.pod.Spec.Containers {
		memoryRequest := container.Resources.Requests[v1.ResourceMemory]
		memory += float64(memoryRequest.Value())
	}
	return memory
}

func (pe *PodExporter) GetPhase() (string, error) {
	podCalculatedStatus, found := pe.pod.Annotations[podStatusAnnotationName]
	if !found {
		return "", fmt.Errorf("could not find %s annotation", podStatusAnnotationName)
	}
	return podCalculatedStatus, nil
}

func (pe *PodExporter) IsBound() bool {
	podStatus := getPodStatus(pe.pod)
	for _, status := range boundStatuses {
		if podStatus == status {
			return true
		}
	}
	return false
}

func (pe *PodExporter) GetGpusFromResourceRequest() float64 {
	gpus := float64(0)
	for _, container := range pe.pod.Spec.Containers {
		gpuRequest := container.Resources.Requests[clustercommonconstants.NvidiaGpuResourceName]
		gpus += float64(gpuRequest.Value())
		amdGpuRequest := container.Resources.Requests[constants.AmdGpuResourceName]
		gpus += float64(amdGpuRequest.Value())
	}
	return gpus
}

func (pe *PodExporter) GetTopOwnerWorkloadId() (string, error) {
	topOwnerUid, found := pe.pod.Labels[constants.TopOwnerWorkloadIdLabel]
	if found && topOwnerUid != "" {
		return topOwnerUid, nil
	}

	topOwnerUid, found = pe.pod.Labels[constants.TopOwnerUidLabel]
	if found && topOwnerUid != "" {
		return topOwnerUid, nil
	}

	return "", fmt.Errorf("could not find %s or %s labels", constants.TopOwnerWorkloadIdLabel, constants.TopOwnerUidLabel)
}

func (pe *PodExporter) GetTopOwnerWorkloadName() string {
	return pe.pod.Labels[constants.ControlPlaneWorkloadNameLabelForWorkload]
}

func (pe *PodExporter) GetGpuNumDevices() (string, error) {
	if !pe.isActive() {
		return "0", nil
	}

	if pe.isFraction() {
		if numDevices, found := pe.pod.Annotations[numberOfDevicesAnnotationName]; found {
			return numDevices, nil
		}
		return "1", nil
	}

	return fmt.Sprintf("%.0f", pe.GetGpusFromResourceRequest()), nil
}

func (pe *PodExporter) isActive() bool {
	podStatus := getPodStatus(pe.pod)
	for _, status := range activeUsedStatuses {
		if podStatus == status {
			return true
		}
	}
	return false
}

func (pe *PodExporter) getMIGDeviceInfo() (*utils.MigDeviceInfo, error) {
	if !pe.isMIG() {
		return nil, fmt.Errorf("asked for MIG Device info for non-mig pod")
	}
	migData, found := pe.pod.Annotations[allocatedMigDeviceAnnotationName]
	if !found {
		return nil, fmt.Errorf("could not find %s annotation", allocatedMigDeviceAnnotationName)
	}
	var migDeviceInfo utils.MigDeviceInfo
	err := json.Unmarshal([]byte(migData), &migDeviceInfo)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal %s annotation value, error: %w",
			allocatedMigDeviceAnnotationName, err)
	}
	return &migDeviceInfo, nil
}

func (pe *PodExporter) getGpuGroup() string {
	gpuGroupValue, found := pe.pod.Labels[clustercommonconstants.GPUGroup]
	if !found {
		return ""
	}
	return gpuGroupValue
}

func (pe *PodExporter) isMIG() bool {
	return pe.getResourceReceivedType() == clustercommonconstants.ReceivedTypeMigInstance
}

func (pe *PodExporter) isFraction() bool {
	return pe.getResourceReceivedType() == clustercommonconstants.ReceivedTypeFraction
}

func (pe *PodExporter) getResourceReceivedType() string {
	receivedType, found := pe.pod.Annotations[receivedResourceTypeAnnotationName]
	if !found {
		return ""
	}
	return receivedType
}

func (pe *PodExporter) isOpportunisticAllocation() bool {
	_, found := pe.pod.Annotations[runaiOpportunisticAnnotationName]
	return found
}

// extractGPUIndex returns the nvidiaVisibleDevices set for the container by runai.
// If no value is set, return with "", nil
func (pe *PodExporter) extractGPUIndex() (string, error) {
	nvidiaVisibleDevicesValue, err := pe.extractGPUIndexFromEnv()
	if err != nil {
		return "", err
	}
	if nvidiaVisibleDevicesValue != "" {
		return nvidiaVisibleDevicesValue, nil
	}

	nvidiaVisibleDevicesValue, err = pe.extractGpuIndexFromDirectEnvConfigMap()
	if err != nil {
		return "", err
	}

	return nvidiaVisibleDevicesValue, nil
}

func (pe *PodExporter) extractGPUIndexFromEnv() (string, error) {
	if len(pe.pod.Spec.Containers) == 0 {
		return "", nil
	}

	for _, envVar := range pe.pod.Spec.Containers[0].Env {
		if envVar.Name != clustercommonconstants.NvidiaVisibleDevices {
			continue
		}
		if envVar.Value != "" {
			return unknownDevice, nil
		}

		if envVar.ValueFrom == nil {
			return "", fmt.Errorf("failed to read value from %s env var", clustercommonconstants.NvidiaVisibleDevices)
		}

		if envVar.ValueFrom.ConfigMapKeyRef == nil {
			return "", fmt.Errorf("failed to read config map reference from %s env var", clustercommonconstants.NvidiaVisibleDevices)
		}

		var configMap v1.ConfigMap
		err := pe.client.Get(pe.ctx, client.ObjectKey{
			Namespace: pe.pod.Namespace,
			Name:      envVar.ValueFrom.ConfigMapKeyRef.Name,
		}, &configMap)

		if err != nil {
			return "", fmt.Errorf("failed to get config map %s. error: %w",
				envVar.ValueFrom.ConfigMapKeyRef.Name, err)
		} else {
			if cmGPUIndex, found := configMap.Data[envVar.ValueFrom.ConfigMapKeyRef.Key]; found {
				return cmGPUIndex, nil
			} else {
				return "", fmt.Errorf("failed to get VALUE from config map.configMap data: %v, error: %w",
					configMap.Data, err)
			}
		}
	}
	return "", nil
}

func (pe *PodExporter) extractGpuIndexFromDirectEnvConfigMap() (string, error) {
	configMapEnvVarSource := map[string]bool{}
	for _, container := range append(pe.pod.Spec.InitContainers, pe.pod.Spec.Containers...) {
		for _, envFrom := range container.EnvFrom {
			if envFrom.ConfigMapRef != nil {
				configMapEnvVarSource[envFrom.ConfigMapRef.Name] = true
			}
		}
	}

	for configMapEnvVarName := range configMapEnvVarSource {
		var configMap v1.ConfigMap
		err := pe.client.Get(pe.ctx, client.ObjectKey{
			Namespace: pe.pod.Namespace,
			Name:      configMapEnvVarName,
		}, &configMap)

		if err != nil {
			return "", fmt.Errorf("extract NVIDIA_VISIBLE_DEVICES from configmap envFrom error: "+
				"failed to get config map %s. error: %w", configMapEnvVarName, err)
		}

		for envKey, envValue := range configMap.Data {
			if envKey == clustercommonconstants.NvidiaVisibleDevices {
				return envValue, nil
			}
		}
	}
	return "", nil
}
