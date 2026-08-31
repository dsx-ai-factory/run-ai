/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package data

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	clustercommonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/common/resourceutils"
	"github.com/run-ai/runai/runai-cluster/cluster/common/resourceutils/resources"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	nodepoolconverter "github.com/run-ai/runai/runai-cluster/common/node-pool-utils/converter"
	nodepoolutils "github.com/run-ai/runai/runai-cluster/common/node-pool-utils/utils"
	topowner "github.com/run-ai/runai/runai-cluster/common/top-owner"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	convertBytesToMB = 1000 * 1000
)

type GpuInfo struct {
	GpuIndex      int
	GpuType       string
	GpuInstanceId string
}

type ResourcesInfo struct {
	Gpus               float64
	GpuMemoryMegaBytes float64
	CpuMilliCores      float64
	CpuMemoryBytes     float64
}

type PhaseLabel string

const (
	runningPhase   PhaseLabel = "Running"
	pendingPhase   PhaseLabel = "Pending"
	suspendedPhase PhaseLabel = "Suspended"
)

type PhaseMetricLabelsInfo struct {
	Phase          PhaseLabel
	DetailedStatus string
}

type PodDetailedPhase int

const (
	// pending means the task is pending in the apiserver.
	pending PodDetailedPhase = 1 << iota

	// binding means the scheduler send Bind request to apiserver.
	binding

	// bound means the task/Pod bounds to a host.
	bound

	// running means a task is running on the host.
	running

	// releasing means a task/pod is deleted.
	releasing

	// succeeded means that all containers in the pod have voluntarily terminated
	// with a container exit code of 0, and the system is not going to restart any of these containers.
	succeeded

	// Failed means that all containers in the pod have terminated, and at least one container has
	// terminated in a failure (exited with a non-zero exit code or was stopped by the system).
	failed

	// Unknown means the status of task/pod is unknown to the scheduler.
	unknown
)

var activeUsedStatuses = map[PodDetailedPhase]bool{
	binding:   true,
	bound:     true,
	running:   true,
	releasing: true}
var boundStatuses = map[PodDetailedPhase]bool{
	bound:     true,
	running:   true,
	releasing: true}

func GetWorkloadIdForPod(pod *corev1.Pod) (string, error) {
	return topowner.GetTopOwnerWorkloadIdFromLabels(pod.Labels)
}

func GetPodPhaseMetricLabelsInfo(pod *corev1.Pod, podDetailedPhase PodDetailedPhase) (*PhaseMetricLabelsInfo, error) {
	podCalculatedStatus, found := pod.Annotations[clustercommonconstants.PodStatusAnnotation]
	if !found {
		return nil, fmt.Errorf("failed finding %s annotation on pod <%s/%s>",
			clustercommonconstants.PodStatusAnnotation, pod.Namespace, pod.Name)
	}

	isPodBound := isBound(podDetailedPhase)

	phase := PhaseLabel("")
	if podCalculatedStatus == string(pendingPhase) {
		phase = pendingPhase
	} else if podCalculatedStatus == string(suspendedPhase) {
		phase = suspendedPhase
	} else if isPodBound {
		phase = runningPhase
	}

	return &PhaseMetricLabelsInfo{
		Phase:          phase,
		DetailedStatus: podCalculatedStatus,
	}, nil
}

func GetNodePoolOfPod(ctx context.Context, pod *corev1.Pod, podDetailedPhase PodDetailedPhase,
	readerClient client.Reader) (string, error) {
	if isBound(podDetailedPhase) {
		return nodepoolutils.GetNodePoolNameFromLabels(pod.Labels, v1alpha1.NodePoolNameLabel, v1alpha1.DefaultNodePoolName), nil
	}

	nodePoolsSources := nodepoolconverter.NewNodePoolsSources(pod.Spec.Affinity, readerClient, nodepoolconverter.ReturnAffinityConversionErrors, pod.Annotations, nil)
	identifiers := nodepoolconverter.NodePoolIdentifiers{
		NodePoolAssignmentLabelKey: v1alpha1.NodePoolNameLabel,
		DefaultNodepoolName:        v1alpha1.DefaultNodePoolName,
		UnexistingNodepoolSentinel: v1alpha1.UnexistingRunaiNodePool,
		AnnotationNodepoolsKey:     commonconstants.AnnotationNodepools,
	}
	requestedNodePools, err := nodepoolconverter.GetRequestedNodePools(ctx, pod.ObjectMeta, "Pod", identifiers, nodePoolsSources)

	if err != nil {
		log.Error().Msgf("For pod <%s/%s>, had an error when getting node pool names, err: <%s>",
			pod.Namespace, pod.Name, err.Error())

		return "", err
	}

	if len(requestedNodePools) == 0 {
		// shouldn't happen if there is no error, GetRequestedNodePools always returns at least 1 node pool, but still
		return "", fmt.Errorf("failed getting node pool name for pod <%s/%s>", pod.Namespace, pod.Name)
	}

	return requestedNodePools[0], nil
}

func GetPodAllocatedResources(pod *corev1.Pod, podDetailedPhase PodDetailedPhase) *ResourcesInfo {
	if !isActive(podDetailedPhase) {
		return &ResourcesInfo{}
	}

	podTemplateSpec := &corev1.PodTemplateSpec{
		ObjectMeta: pod.ObjectMeta,
		Spec:       pod.Spec,
	}

	allocatedResources := getPodAllocatedResources(podTemplateSpec)
	return parseResources(allocatedResources)
}

func GetPodResources(pod *corev1.Pod) (*ResourcesInfo, *ResourcesInfo) {
	podTemplateSpec := &corev1.PodTemplateSpec{
		ObjectMeta: pod.ObjectMeta,
		Spec:       pod.Spec,
	}

	requests, limits := getPodResources(podTemplateSpec)
	return parseResources(requests), parseResources(limits)
}

func parseResources(resources corev1.ResourceList) *ResourcesInfo {
	gpus := resources.Name(clustercommonconstants.GpuResourceName, resource.DecimalSI)
	gpus.Add(*resources.Name(clustercommonconstants.NvidiaGpuResourceName, resource.DecimalSI))
	gpus.Add(*resources.Name(clustercommonconstants.AmdGpuResourceName, resource.DecimalSI))
	gpuMemory := resources.Name(clustercommonconstants.GpuMemoryResourceName, resource.DecimalSI)

	return &ResourcesInfo{
		Gpus:               gpus.AsApproximateFloat64(),
		GpuMemoryMegaBytes: gpuMemory.AsApproximateFloat64() / convertBytesToMB,
		CpuMilliCores:      float64(resources.Cpu().MilliValue()),
		CpuMemoryBytes:     resources.Memory().AsApproximateFloat64(),
	}
}

func getPodAllocatedResources(podTemplate *corev1.PodTemplateSpec) corev1.ResourceList {
	return resourceutils.PodRunaiAllocatedResources(podTemplate.Spec.Containers, podTemplate.Annotations)
}

func getPodResources(podTemplate *corev1.PodTemplateSpec) (corev1.ResourceList, corev1.ResourceList) {
	return resources.PodRunaiRequestsAndLimits(podTemplate.Spec.Containers,
		podTemplate.Spec.InitContainers, podTemplate.Annotations)
}

func isBound(podDetailedPhase PodDetailedPhase) bool {
	if _, found := boundStatuses[podDetailedPhase]; found {
		return true
	}
	return false
}

func isActive(podDetailedPhase PodDetailedPhase) bool {
	if _, found := activeUsedStatuses[podDetailedPhase]; found {
		return true
	}
	return false
}

func GetPodDetailedPhase(pod *corev1.Pod) PodDetailedPhase {
	switch pod.Status.Phase {
	case corev1.PodRunning:
		if pod.DeletionTimestamp != nil {
			return releasing
		}

		return running
	case corev1.PodPending:
		if len(pod.Spec.NodeName) == 0 {
			return pending
		}
		if pod.DeletionTimestamp != nil {
			return releasing
		}
		return bound
	case corev1.PodSucceeded:
		return succeeded
	case corev1.PodFailed:
		return failed
	}

	return unknown
}
