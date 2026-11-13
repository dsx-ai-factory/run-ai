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
	"sort"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/resourceutils/resources"
	"k8s.io/apimachinery/pkg/api/resource"

	v1 "k8s.io/api/core/v1"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

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

	// Blackwell profiles may carry a media extensions suffix: "+me" / "+me.all" in the profile name (annotation) and
	// ".me" / ".me.all" in the resource name, since the NVIDIA device plugin replaces "+" with "."; "-me" excludes them.
	migExtractionRegex      = `^nvidia.com/mig-(\d+)g\.(\d+)gb(?:[+.-][a-z]+(?:\.[a-z]+)?)?$`
	migRegexNumOfOutputs    = 3
	migDeviceAnnotationName = "runai-mig-device"
	deviceNameKey           = "name"

	nodeGpuTypeLabelName   = "nvidia.com/gpu.product"
	A100                   = "A100"
	H100                   = "H100"
	H200                   = "H200"
	B200                   = "B200" // also matches GB200
	B300                   = "B300"
	RTXPro5000             = "RTX-PRO-5000"
	RTXPro6000             = "RTX-PRO-6000"
	defaultComputePortions = float64(7)
	rtxPro5000Portions     = float64(2)
	rtxPro6000Portions     = float64(4)

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
	return int64(getGpuMemoryRequestInMb(pod))
}

func getGpuMemoryRequestInMb(pod *v1.Pod) float64 {
	memory, err := resources.ReadGPUMemoryAnnotations(pod.Annotations)
	if err != nil {
		log.Log.Error(err, "Failed to parse GPU request", "namespace", pod.Namespace, "pod", pod.Name)
		return 0
	}
	// Status annotations use decimal MB even when the request uses binary units.
	return memory.Request.AsApproximateFloat64() / ByteToMb
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

// getAllocatedMigPortion returns the fraction of a whole GPU allocated to the pod through MIG devices: the compute
// slices of every allocated profile, times its count, divided by the compute slices of the node's GPU product.
func getAllocatedMigPortion(ctx context.Context, pod *v1.Pod, kubeClient client.Client) (float64, error) {
	if len(pod.Spec.NodeName) == 0 { // Can't match metric if the pod isn't scheduled to a pod
		return 0, nil
	}

	devices, err := getMigDevices(pod)
	if err != nil {
		return 0, err
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

	migGPUs := float64(0)
	for _, device := range devices {
		migCompute, _, err := extractGpuAndMemoryFromMigResourceName(device.name)
		if err != nil {
			return 0, fmt.Errorf("failed to extract mig device type from %s in annotation %s, pod %s/%s. %w",
				device.name, migDeviceAnnotationName, pod.Namespace, pod.Name, err)
		}
		migGPUs += device.count * float64(migCompute) / wholeGpuNumOfPortions
	}
	return migGPUs, nil
}

// migDevice is a MIG profile resource name (e.g. nvidia.com/mig-1g.24gb) and how many of them the pod uses.
type migDevice struct {
	name  string
	count float64
}

// getMigDevices returns the MIG devices of the pod from the runai-mig-device annotation. The scheduler may leave the
// annotation empty for pods that request a MIG profile directly in their resources (mig.strategy=mixed), so an empty
// annotation, or one without a device name, falls back to the MIG resources requested by the pod.
func getMigDevices(pod *v1.Pod) ([]migDevice, error) {
	deviceJsonStr := pod.Annotations[migDeviceAnnotationName]
	if deviceJsonStr == "" {
		log.Log.V(1).Info("No mig-device annotation, using the requested MIG resources",
			"namespace", pod.Namespace, "pod", pod.Name)
		return getRequestedMigResources(pod)
	}

	var deviceJson map[string]any
	if err := json.Unmarshal([]byte(deviceJsonStr), &deviceJson); err != nil {
		return nil, fmt.Errorf("failed to decode mig-device json from annotation %s, pod %s/%s. %w",
			migDeviceAnnotationName, pod.Namespace, pod.Name, err)
	}
	deviceName, ok := deviceJson[deviceNameKey].(string)
	if !ok || deviceName == "" {
		log.Log.V(1).Info("mig-device annotation has no device name, using the requested MIG resources",
			"namespace", pod.Namespace, "pod", pod.Name)
		return getRequestedMigResources(pod)
	}
	return []migDevice{{name: deviceName, count: 1}}, nil
}

// getRequestedMigResources returns every MIG profile requested by the pod with its quantity, in a stable order.
func getRequestedMigResources(pod *v1.Pod) ([]migDevice, error) {
	// Extended resources may be set only in limits; the API server copies them to requests.
	req, limits := resourcehelper.PodRequestsAndLimits(pod)
	for _, resourceList := range []v1.ResourceList{req, limits} {
		names := make([]string, 0, len(resourceList))
		for rName, quantity := range resourceList {
			if isMigResource(string(rName)) && !quantity.IsZero() {
				names = append(names, string(rName))
			}
		}
		if len(names) == 0 {
			continue
		}
		sort.Strings(names)
		devices := make([]migDevice, 0, len(names))
		for _, name := range names {
			quantity := resourceList[v1.ResourceName(name)]
			devices = append(devices, migDevice{name: name, count: float64(quantity.Value())})
		}
		return devices, nil
	}
	return nil, fmt.Errorf("pod %s/%s has no %s annotation and no mig resource request",
		pod.Namespace, pod.Name, migDeviceAnnotationName)
}

// matchAllMetrics recalibrates the gpu metrics in podMetrics, so they will match to each other.
// For example, if the pod requested 0.5 a gpu and has been scheduled to a node with gpus of 4G each, matchAllMetrics
// will set podMetrics.GpuMemory to be 2000 (Mi).
func matchAllMetrics(ctx context.Context, pod *v1.Pod, podMetrics *metadata.PodMetadata,
	kubeClient client.Client) error {
	requestExpressedInGpus := podMetrics.GPUs > 0
	requestedGpuMemory := getGpuMemoryRequestInMb(pod)
	requestExpressedInGpuMemory := requestedGpuMemory > 0

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
		podMetrics.GPUs = requestedGpuMemory / singleGpuMemory
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

// getDeviceComputePortions returns the number of MIG compute slices of a whole GPU, per the NVIDIA MIG user guide
// (supported MIG profiles).
func getDeviceComputePortions(gpuProduct string) (float64, error) {
	for _, product := range []string{A100, H100, H200, B200, B300} {
		if strings.Contains(gpuProduct, product) {
			return defaultComputePortions, nil
		}
	}
	if strings.Contains(gpuProduct, RTXPro6000) {
		return rtxPro6000Portions, nil
	}
	if strings.Contains(gpuProduct, RTXPro5000) {
		return rtxPro5000Portions, nil
	}
	return -1, fmt.Errorf("cannot calculate device portion for gpu product %s", gpuProduct)
}

func isMigResource(rName string) bool {
	return strings.HasPrefix(rName, migResourcePrefix)
}

func trimMigDeviceNamePrefix(migResourceFullName string) string {
	return strings.TrimPrefix(migResourceFullName, migResourcePrefix)
}
