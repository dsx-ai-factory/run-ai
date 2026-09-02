/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"

	v1 "k8s.io/api/core/v1"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"
)

const (
	receivedResourceAnnotationName = "received-resource-type"
	receivedTypeMigInstance        = "MigInstance"
	receivedTypeFraction           = "Fraction"
	receivedTypeRegular            = "Regular"
	receivedTypeNone               = ""

	gpuFractionAnnotation = "gpu-fraction"
	gpuFractionsCountName = "gpu-fraction-num-devices"
	gpuMemoryAnnotation   = "gpu-memory"

	gpuResourceName    = "nvidia.com/gpu"
	amdGpuResourceName = "amd.com/gpu"

	nodeGpuMemoryLabelName            = "nvidia.com/gpu.memory"
	amdNodeGpuMemoryLabelPrefix       = "beta.amd.com/gpu.vram"
	amdNodeGpuMemoryLabelValuePattern = "(\\d+.*)"
	amdNodeGpuMemoryValueFactor       = 1000000

	migExtractionRegex      = `^nvidia.com/mig-(\d+)g\.(\d+)gb$`
	migRegexNumOfOutputs    = 3
	migDeviceAnnotationName = "runai-mig-device"
	deviceNameKey           = "name"

	nodeGpuTypeLabelName   = "nvidia.com/gpu.product"
	A100                   = "A100"
	H100                   = "H100"
	defaultComputePortions = float64(7)

	migResourcePrefix = "nvidia.com/mig-"

	ByteToMib = 1024 * 1024
	ByteToMb  = 1000 * 1000
	TibInMib  = 1024 * 1024
)

func CalcPodGpuResources(ctx context.Context, pod *v1.Pod, podMetrics *metadata.PodMetadata,
	kubeClient client.Client) error {
	if !IsActiveUsedStatus(GetExtendedPhase(pod)) {
		return nil
	}

	receivedResourceType := pod.Annotations[receivedResourceAnnotationName]
	switch receivedResourceType {
	case receivedTypeNone:
		return nil
	case receivedTypeRegular:
		podMetrics.GPUs = GetGpuRequest(pod)
		return matchAllMetrics(ctx, pod, podMetrics, kubeClient)
	case receivedTypeFraction:
		podMetrics.GPUs = GetGpuRequest(pod)
		podMetrics.GpuMemory = GetGpuMemoryRequest(pod)
		podMetrics.FractionsCount = GetGpuFractionsCount(pod)
		return matchAllMetrics(ctx, pod, podMetrics, kubeClient)
	case receivedTypeMigInstance:
		migGPUs, err := getAllocatedMigPortion(ctx, pod, kubeClient)
		if err != nil {
			return err
		}
		podMetrics.MigGPUs = migGPUs
	default:
		return fmt.Errorf("pod %s/%s has unknown resource type %s", pod.Namespace, pod.Name, receivedResourceType)
	}

	return nil
}

func GetGpuRequest(pod *v1.Pod) float64 {
	req, _ := resourcehelper.PodRequestsAndLimits(pod)
	gpuReq := req[gpuResourceName]
	gpusCount := float64(gpuReq.Value())
	amdGpuReq := req[amdGpuResourceName]
	gpusCount += float64(amdGpuReq.Value())

	gpuFraction, gpuFractionErr := strconv.ParseFloat(pod.Annotations[gpuFractionAnnotation], 64)
	if !(gpuFraction <= 0 || gpuFraction > 1 || gpuFractionErr != nil) {
		gpusCount += gpuFraction
	}

	return gpusCount
}

func GetGpuMemoryRequest(pod *v1.Pod) int64 {
	gpuMemory, err := strconv.ParseInt(pod.Annotations[gpuMemoryAnnotation], 10, 64)
	if err == nil && gpuMemory > 0 {
		return gpuMemory
	}
	return 0
}

func GetGpuFractionsCount(pod *v1.Pod) int64 {
	fractionsCount, err := strconv.ParseInt(pod.Annotations[gpuFractionsCountName], 10, 64)
	if err == nil && fractionsCount > 0 {
		return fractionsCount
	}
	return 1
}

func GetRequestedMigDevices(pod *v1.Pod) map[string]int64 {
	requestedMigDevices := map[string]int64{}
	for annotationName, annotationValue := range pod.Annotations {
		if isMigResource(annotationName) {
			value, err := strconv.ParseInt(annotationValue, 10, 64)
			if err != nil {
				continue
			}
			requestedMigDevices[trimMigDeviceNamePrefix(annotationName)] = value
		}
	}
	return requestedMigDevices
}

func getAllocatedMigPortion(ctx context.Context, pod *v1.Pod, kubeClient client.Client) (float64, error) {
	if len(pod.Spec.NodeName) == 0 { // Can't match metric if the pod isn't scheduled to a pod
		return 0, nil
	}

	deviceJsonStr := pod.Annotations[migDeviceAnnotationName]
	var deviceJson map[string]any
	err := json.Unmarshal([]byte(deviceJsonStr), &deviceJson)
	if err != nil {
		return 0, fmt.Errorf("failed to decode mig-device json from annotation %s, pod %s/%s. %w",
			migDeviceAnnotationName, pod.Namespace, pod.Name, err)
	}

	deviceTypeName := fmt.Sprintf("%v", deviceJson[deviceNameKey])
	migCompute, _, err := extractGpuAndMemoryFromMigResourceName(deviceTypeName)
	if err != nil {
		return 0, fmt.Errorf("failed to extract mig device type from %s in annotation %s, pod %s/%s. %w",
			deviceTypeName, migDeviceAnnotationName, pod.Namespace, pod.Name, err)
	}
	gpuProduct, err := getNodeGpuType(ctx, pod, kubeClient)
	if err != nil {
		return 0, fmt.Errorf("failed to extract node gpu type for node %s. pod %s/%s. %w",
			pod.Spec.NodeName, pod.Namespace, pod.Name, err)
	}
	wholeGpuNumOfPortions, err := getDeviceComputePortions(gpuProduct)
	if err != nil {
		return 0, fmt.Errorf("failed to get device portion for node %s. pod %s/%s. %w",
			pod.Spec.NodeName, pod.Namespace, pod.Name, err)
	}

	migGPUs := float64(migCompute) / wholeGpuNumOfPortions
	return migGPUs, nil
}

// matchAllMetrics recalibrates the gpu metrics in podMetrics, so they will match to each other.
// For example, if the pod requested 0.5 a gpu and has been scheduled to a node with gpus of 4G each, matchAllMetrics
// will set podMetrics.GpuMemory to be 2000 (Mi).
func matchAllMetrics(ctx context.Context, pod *v1.Pod, podMetrics *metadata.PodMetadata,
	kubeClient client.Client) error {
	requestExpressedInGpus := podMetrics.GPUs > 0
	requestExpressedInGpuMemory := podMetrics.GpuMemory > 0

	if len(pod.Spec.NodeName) == 0 { // Can't match metric if the pod isn't scheduled to a pod
		return nil
	}

	if !requestExpressedInGpus && !requestExpressedInGpuMemory {
		// In case no gpus has been requested, zero all the gpu metrics
		podMetrics.GPUs = 0
		podMetrics.GpuMemory = 0
		return nil
	}

	singleGpuMemory, err := getNodeSingleGpuMemory(ctx, pod, kubeClient)
	if err != nil {
		return err
	}

	if requestExpressedInGpus {
		podMetrics.GpuMemory = int64(singleGpuMemory * podMetrics.GPUs)
	} else if requestExpressedInGpuMemory {
		podMetrics.GPUs = float64(podMetrics.GpuMemory) / singleGpuMemory
	}

	if podMetrics.FractionsCount > 1 {
		podMetrics.GPUs = podMetrics.GPUs * float64(podMetrics.FractionsCount)
		podMetrics.GpuMemory = podMetrics.GpuMemory * podMetrics.FractionsCount
	}

	return nil
}

func getNodeSingleGpuMemory(ctx context.Context, pod *v1.Pod, kubeClient client.Client) (float64, error) {
	podNode := v1.Node{}
	err := kubeClient.Get(ctx, client.ObjectKey{Name: pod.Spec.NodeName}, &podNode)
	if err != nil {
		return 0, err
	}
	nodeGpuMemoryStr, found := podNode.Labels[nodeGpuMemoryLabelName]
	if !found {
		nodeGpuMemoryStr = "0"
	}

	gpuMemoryLabelValue, err := strconv.Atoi(nodeGpuMemoryStr)
	if err != nil {
		return 0, err
	}

	// This code is a fix to cover for the Gpu-feature-discovery bug
	// (issue https://github.com/NVIDIA/gpu-feature-discovery/issues/26)
	if !checkGpuMemoryIsInMib(gpuMemoryLabelValue) {
		gpuMemoryLabelValue = convertBytesToMib(gpuMemoryLabelValue)
	}

	// ultimately, the nodeGpuMemoryLabelName label value is in MiB
	singleGpuMemory := convertMibToMb(gpuMemoryLabelValue)

	singleAmdGpuMemory, err := getNodeSingleAmdGpuMemory(&podNode)
	if err != nil {
		return 0, err
	}
	singleGpuMemory += float64(singleAmdGpuMemory)

	return singleGpuMemory, nil
}

// The boundary itself counts as MiB: a 1 TiB device is conceivable, a 1 MiB one is not.
func checkGpuMemoryIsInMib(gpuMemoryValue int) bool {
	return gpuMemoryValue <= TibInMib
}

func convertBytesToMib(gpuMemoryValue int) int {
	return gpuMemoryValue / ByteToMib
}

func convertMibToMb(countInMib int) float64 {
	return float64(countInMib) * ByteToMib / ByteToMb
}

func getNodeSingleAmdGpuMemory(node *v1.Node) (int, error) {
	for label := range node.Labels {
		if !strings.HasPrefix(label, amdNodeGpuMemoryLabelPrefix) {
			continue
		}
		regex := regexp.MustCompile(amdNodeGpuMemoryLabelValuePattern)
		match := regex.FindString(label)
		if match != "" {
			res := resource.MustParse(match)
			value, success := res.AsInt64()
			if !success {
				return 0, fmt.Errorf("could not extract memory amount from amd gpu memory label %s", label)
			}
			return int(value) / amdNodeGpuMemoryValueFactor, nil
		}
	}
	return 0, nil
}

func extractGpuAndMemoryFromMigResourceName(migResourceName string) (int, int, error) {
	matches := regexp.MustCompile(migExtractionRegex).FindStringSubmatch(migResourceName)
	if len(matches) < migRegexNumOfOutputs {
		return -1, -1, fmt.Errorf("failed to extract gpu/memory from %v", migResourceName)
	}

	gpu, err := strconv.Atoi(matches[1])
	if err != nil {
		return -1, -1, fmt.Errorf("failed parsing %v to integer", matches[2])
	}
	mem, err := strconv.Atoi(matches[2])
	if err != nil {
		return -1, -1, fmt.Errorf("failed parsing %v to integer", matches[2])
	}

	return gpu, mem, nil
}

func getNodeGpuType(ctx context.Context, pod *v1.Pod, kubeClient client.Client) (string, error) {
	podNode := v1.Node{}
	err := kubeClient.Get(ctx, client.ObjectKey{Name: pod.Spec.NodeName}, &podNode)
	if err != nil {
		return "", err
	}
	nodeGpuType := podNode.Labels[nodeGpuTypeLabelName]
	if len(nodeGpuType) == 0 {
		return "", fmt.Errorf("node %s doesn't have a nvidia.com/gpu.product label."+
			" The pod %s requires mig GPU resources and have been schedualed on the node", pod.Spec.NodeName, pod.Name)
	}

	return nodeGpuType, nil
}

func getDeviceComputePortions(gpuProduct string) (float64, error) {
	if strings.Contains(gpuProduct, A100) || strings.Contains(gpuProduct, H100) {
		return defaultComputePortions, nil
	}
	return -1, fmt.Errorf("cannot calculate device portion for gpu product %s", gpuProduct)
}

func isMigResource(rName string) bool {
	return strings.HasPrefix(rName, migResourcePrefix)
}

func trimMigDeviceNamePrefix(migResourceFullName string) string {
	return strings.TrimPrefix(migResourceFullName, migResourcePrefix)
}
