/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
)

func IsMigResource(rName string) bool {
	return strings.HasPrefix(rName, constants.AnnotationMigProfilePrefix)
}

func HasMigDeviceRequest(pod *v1.Pod) bool {
	for _, container := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
		for resource := range container.Resources.Limits {
			if IsMigResource(resource.String()) {
				return true
			}
		}
		for resource := range container.Resources.Requests {
			if IsMigResource(resource.String()) {
				return true
			}
		}
	}
	return false
}

// ExtractGpuAndMemoryFromMigResourceName Returns memory in GB
func ExtractGpuAndMemoryFromMigResourceName(migResourceName string) (int, int, error) {
	return extractGpuAndMemoryFrom(`^nvidia.com/mig-(\d+)g\.(\d+)gb$`, migResourceName)
}

func extractGpuAndMemoryFrom(searchInPattern string, migString string) (int, int, error) {
	matches := regexp.MustCompile(searchInPattern).FindStringSubmatch(migString)
	if len(matches) < 3 {
		return -1, -1, fmt.Errorf("failed to extract gpu/memory from %v", migString)
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

func calculateAllocatedMigPortion(
	ctx context.Context, pod *v1.Pod, kubeClient client.Client,
) (migDeviceType string, migPortion float64, calculationError error) {
	if len(pod.Spec.NodeName) == 0 {
		return "", 0, fmt.Errorf(
			"cannot calculate allocated mig portion for non scheduled pod (no node name). Pod %s/%s",
			pod.Namespace, pod.Name)
	}

	deviceJsonStr := pod.Annotations[migDeviceAnnotationName]
	var deviceJson map[string]any
	err := json.Unmarshal([]byte(deviceJsonStr), &deviceJson)
	if err != nil {
		return "", 0, fmt.Errorf("failed to decode mig-device json from annotation %s, pod %s/%s. %w",
			migDeviceAnnotationName, pod.Namespace, pod.Name, err)
	}

	deviceTypeName := fmt.Sprintf("%v", deviceJson[deviceNameKey])
	migCompute, _, err := extractGpuAndMemoryFromMigResourceName(deviceTypeName)
	if err != nil {
		return "", 0, fmt.Errorf("failed to extract mig device type from %s in annotation %s, pod %s/%s. %w",
			deviceTypeName, migDeviceAnnotationName, pod.Namespace, pod.Name, err)
	}
	gpuProduct, err := getNodeGpuType(ctx, pod, kubeClient)
	if err != nil {
		return "", 0, fmt.Errorf("failed to extract node gpu type for node %s. pod %s/%s. %w",
			pod.Spec.NodeName, pod.Namespace, pod.Name, err)
	}
	wholeGpuNumOfPortions, err := getDeviceComputePortions(gpuProduct)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get device portion for node %s. pod %s/%s. %w",
			pod.Spec.NodeName, pod.Namespace, pod.Name, err)
	}

	migGPUs := float64(migCompute) / wholeGpuNumOfPortions
	return deviceTypeName, migGPUs, nil
}

func extractGpuAndMemoryFromMigResourceName(migResourceName string) (migComputeSize int, migMemorySize int, err error) {
	matches := regexp.MustCompile(migExtractionRegex).FindStringSubmatch(migResourceName)
	if len(matches) < migRegexNumOfOutputs {
		return -1, -1, fmt.Errorf("failed to extract gpu/memory from %v", migResourceName)
	}

	var castErr error
	migComputeSize, castErr = strconv.Atoi(matches[1])
	if castErr != nil {
		return -1, -1, fmt.Errorf("failed parsing %v to integer", matches[2])
	}
	migMemorySize, castErr = strconv.Atoi(matches[2])
	if castErr != nil {
		return -1, -1, fmt.Errorf("failed parsing %v to integer", matches[2])
	}

	return migComputeSize, migMemorySize, nil
}

func getNodeGpuType(ctx context.Context, pod *v1.Pod, kubeClient client.Client) (string, error) {
	podNode := v1.Node{}
	err := kubeClient.Get(ctx, client.ObjectKey{Name: pod.Spec.NodeName}, &podNode)
	if err != nil {
		return "", err
	}
	nodeGpuType, hasLabel := podNode.Labels[constants.NvidiaGPUProduct]
	if !hasLabel || len(nodeGpuType) == 0 {
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
