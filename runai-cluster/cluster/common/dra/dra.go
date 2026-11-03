/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dra

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	draclient "k8s.io/dynamic-resource-allocation/client"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logger.Log.WithName("DRAParser")

const nvidiaGpuDeviceClassName = "gpu.nvidia.com"

// NvidiaGpuResourceName is the extended resource name for NVIDIA GPUs.
const NvidiaGpuResourceName corev1.ResourceName = "nvidia.com/gpu"

// DRAParser resolves GPU counts from DRA ResourceClaim and ResourceClaimTemplate objects.
// It uses a version-aware DRA client that auto-detects which API version
// (v1, v1beta2, v1beta1) the cluster supports and converts all responses to resource/v1 types.
// The v1 Get calls are served from the cached reader when possible.
type DRAParser struct {
	draClient *draclient.Client
}

// NewDRAParser creates a DRAParser with a version-aware DRA client.
// The cachedReader is used to serve v1 ResourceClaim and ResourceClaimTemplate Get calls from cache.
// The kubeClient is used as fallback for version negotiation (v1beta1, v1beta2).
func NewDRAParser(cachedReader client.Reader, kubeClient kubernetes.Interface) *DRAParser {
	return &DRAParser{
		draClient: draclient.New(newCachedKubeClient(cachedReader, kubeClient)),
	}
}

// ResolveDRAGpuCount resolves the total NVIDIA GPU count from pod-level resource claims
// by fetching the referenced ResourceClaim or ResourceClaimTemplate objects from the cluster.
func (p *DRAParser) ResolveDRAGpuCount(ctx context.Context, namespace string, podResourceClaims []corev1.PodResourceClaim) int64 {
	var totalGpuCount int64
	for _, prc := range podResourceClaims {
		gpuCount := p.ResolveNamedClaimGpuCount(ctx, namespace, prc.ResourceClaimName, prc.ResourceClaimTemplateName)
		totalGpuCount += gpuCount
	}
	return totalGpuCount
}

// ResolvePodGpuCount resolves the NVIDIA GPU count a pod requests through DRA, counting every
// pod-level claim once and returning lookup errors (unlike ResolveDRAGpuCount, which logs them and returns 0).
//   - ResourceClaimName: the referenced claim is read.
//   - ResourceClaimTemplateName: the claim generated for the pod (named in pod.Status.ResourceClaimStatuses)
//     is read when it exists; before it exists (e.g. an unscheduled pod) the template is read instead.
//     Both describe the same request, so only one of them is counted.
func (p *DRAParser) ResolvePodGpuCount(ctx context.Context, pod *corev1.Pod) (int64, error) {
	var totalGpuCount int64
	counted := map[string]struct{}{}
	for _, prc := range pod.Spec.ResourceClaims {
		key, spec, err := p.podClaimSpec(ctx, pod, prc)
		if err != nil {
			return 0, err
		}
		if spec == nil {
			continue
		}
		if _, found := counted[key]; found {
			continue
		}
		counted[key] = struct{}{}
		totalGpuCount += extractGpuCountFromClaimSpec(*spec)
	}
	return totalGpuCount, nil
}

// podClaimSpec returns the claim spec of one pod-level claim and a key identifying what was read.
func (p *DRAParser) podClaimSpec(ctx context.Context, pod *corev1.Pod, prc corev1.PodResourceClaim) (string, *resourcev1.ResourceClaimSpec, error) {
	claimName := ""
	if prc.ResourceClaimName != nil {
		claimName = *prc.ResourceClaimName
	} else {
		for _, status := range pod.Status.ResourceClaimStatuses {
			if status.Name == prc.Name && status.ResourceClaimName != nil {
				claimName = *status.ResourceClaimName
				break
			}
		}
	}

	if claimName != "" {
		claim, err := p.draClient.ResourceClaims(pod.Namespace).Get(ctx, claimName, metav1.GetOptions{})
		if err == nil {
			return "claim/" + claimName, &claim.Spec, nil
		}
		// A generated claim that is not readable yet falls back to its template.
		if prc.ResourceClaimName != nil || !apierrors.IsNotFound(err) {
			return "", nil, fmt.Errorf("getting ResourceClaim %s/%s: %w", pod.Namespace, claimName, err)
		}
	}

	if prc.ResourceClaimTemplateName != nil && *prc.ResourceClaimTemplateName != "" {
		template, err := p.draClient.ResourceClaimTemplates(pod.Namespace).Get(ctx, *prc.ResourceClaimTemplateName, metav1.GetOptions{})
		if err != nil {
			return "", nil, fmt.Errorf("getting ResourceClaimTemplate %s/%s: %w", pod.Namespace, *prc.ResourceClaimTemplateName, err)
		}
		return "template/" + prc.Name, &template.Spec.Spec, nil
	}

	return "", nil, nil
}

// ResolveNamedClaimGpuCount resolves GPU count from an explicit ResourceClaim name or ResourceClaimTemplate name.
// Exactly one of claimName or templateName should be non-nil.
func (p *DRAParser) ResolveNamedClaimGpuCount(ctx context.Context, namespace string, claimName *string, templateName *string) int64 {
	if claimName != nil && *claimName != "" {
		claim, err := p.draClient.ResourceClaims(namespace).Get(ctx, *claimName, metav1.GetOptions{})
		if err != nil {
			log.Error(err, "Warning: failed to get ResourceClaim", "namespace", namespace, "name", *claimName)
			return 0
		}
		return extractGpuCountFromClaimSpec(claim.Spec)
	}

	if templateName != nil && *templateName != "" {
		template, err := p.draClient.ResourceClaimTemplates(namespace).Get(ctx, *templateName, metav1.GetOptions{})
		if err != nil {
			log.Error(err, "Warning: failed to get ResourceClaimTemplate", "namespace", namespace, "name", *templateName)
			return 0
		}
		return extractGpuCountFromClaimSpec(template.Spec.Spec)
	}

	return 0
}

// extractGpuCountFromClaimSpec extracts the NVIDIA GPU count from a ResourceClaimSpec.
// Only supports ExactCount allocation mode with deviceClassName "gpu.nvidia.com".
func extractGpuCountFromClaimSpec(spec resourcev1.ResourceClaimSpec) int64 {
	var totalGpuCount int64
	for _, request := range spec.Devices.Requests {
		if request.Exactly == nil {
			if request.FirstAvailable != nil {
				log.Info("Warning: DRA device request uses FirstAvailable mode; GPU count resolution is not supported for this mode", "request", request.Name)
			}
			continue
		}
		exact := request.Exactly
		if exact.DeviceClassName != nvidiaGpuDeviceClassName {
			continue
		}
		if exact.AllocationMode != "" && exact.AllocationMode != resourcev1.DeviceAllocationModeExactCount {
			log.Info("Warning: DRA device request uses unsupported allocation mode; skipping GPU count resolution",
				"request", request.Name, "allocationMode", exact.AllocationMode, "deviceClassName", exact.DeviceClassName)
			continue
		}
		if exact.Count == 0 {
			log.Info("Warning: DRA device request has count 0; skipping", "request", request.Name)
			continue
		}
		totalGpuCount += exact.Count
	}
	return totalGpuCount
}

// AddDRAGpuToResources adds the given GPU count to both Requests and Limits of a ResourceRequirements.
func AddDRAGpuToResources(gpuCount int64, resources *corev1.ResourceRequirements) {
	if gpuCount <= 0 {
		return
	}
	gpuQty := resource.NewQuantity(gpuCount, resource.DecimalSI)

	if resources.Requests == nil {
		resources.Requests = corev1.ResourceList{}
	}
	if resources.Limits == nil {
		resources.Limits = corev1.ResourceList{}
	}

	addToResourceList(resources.Requests, NvidiaGpuResourceName, *gpuQty)
	addToResourceList(resources.Limits, NvidiaGpuResourceName, *gpuQty)
}

// addToResourceList adds or merges a resource quantity into a ResourceList.
func addToResourceList(list corev1.ResourceList, name corev1.ResourceName, qty resource.Quantity) {
	if existing, found := list[name]; found {
		existing.Add(qty)
		list[name] = existing
	} else {
		list[name] = qty.DeepCopy()
	}
}
