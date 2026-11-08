/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dra

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	draclient "k8s.io/dynamic-resource-allocation/client"
	"k8s.io/dynamic-resource-allocation/resourceclaim"
	"k8s.io/utils/ptr"
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

// ErrClaimDeleted reports that a claim or template of a finished Pod no longer exists.
// Kubernetes can delete generated claims once their Pod finishes, so this is permanent and callers must not retry.
var ErrClaimDeleted = errors.New("the claim or template of a finished Pod no longer exists")

// ResolvePodRequestedGpuCount returns the NVIDIA GPU count requested through DRA.
// It reads claims referenced directly by the Pod, or claims created from
// templates. Before a generated claim is named in the Pod's status, it reads
// the template. Each claim is counted once, and lookup errors are returned.
// A claim or template that is missing for an active Pod is a lookup error to retry; for a Pod that
// finished (Succeeded or Failed) the error wraps ErrClaimDeleted, which must not be retried.
func (p *DRAParser) ResolvePodRequestedGpuCount(ctx context.Context, pod *corev1.Pod) (int64, error) {
	var totalGpuCount int64
	countedClaims := map[string]struct{}{}
	for _, claimRef := range pod.Spec.ResourceClaims {
		claimKey, claimSpec, err := p.fetchPodClaimSpec(ctx, pod, claimRef)
		if err != nil {
			return 0, err
		}
		if claimSpec == nil {
			continue
		}
		if _, found := countedClaims[claimKey]; found {
			continue
		}
		countedClaims[claimKey] = struct{}{}
		totalGpuCount += extractGpuCountFromClaimSpec(*claimSpec)
	}
	return totalGpuCount, nil
}

// ResolveDRAGpuCount resolves the total NVIDIA GPU count from pod-level resource claims.
// Deprecated: prefer ResolvePodRequestedGpuCount, which de-duplicates claims and returns errors.
func (p *DRAParser) ResolveDRAGpuCount(ctx context.Context, namespace string, podResourceClaims []corev1.PodResourceClaim) int64 {
	var totalGpuCount int64
	for _, prc := range podResourceClaims {
		totalGpuCount += p.ResolveNamedClaimGpuCount(ctx, namespace, prc.ResourceClaimName, prc.ResourceClaimTemplateName)
	}
	return totalGpuCount
}

// fetchPodClaimSpec reads the requested resource specification from a claim
// or template. It also returns a key used to prevent counting the same request twice.
func (p *DRAParser) fetchPodClaimSpec(ctx context.Context, pod *corev1.Pod, claimRef corev1.PodResourceClaim) (string, *resourcev1.ResourceClaimSpec, error) {
	claimName, mustCheckOwner, err := resourceclaim.Name(pod, &claimRef)
	if err != nil && !errors.Is(err, resourceclaim.ErrClaimNotFound) {
		return "", nil, err
	}

	if err == nil {
		if claimName == nil {
			// Kubernetes intentionally did not create a claim for this reference.
			return "", nil, nil
		}
		// A claim that is named but not readable yet (for example, not in the cache) is returned as an error so the caller retries.
		claim, getErr := p.draClient.ResourceClaims(pod.Namespace).Get(ctx, *claimName, metav1.GetOptions{})
		if getErr != nil {
			return "", nil, claimReadError(pod, "ResourceClaim", *claimName, getErr)
		}
		if mustCheckOwner {
			if ownerErr := resourceclaim.IsForPod(pod, claim); ownerErr != nil {
				return "", nil, ownerErr
			}
		}
		return "claim/" + *claimName, &claim.Spec, nil
	}

	// The Pod's status does not name a generated claim yet, so the template describes the request.
	if templateName := ptr.Deref(claimRef.ResourceClaimTemplateName, ""); templateName != "" {
		template, getErr := p.draClient.ResourceClaimTemplates(pod.Namespace).Get(ctx, templateName, metav1.GetOptions{})
		if getErr != nil {
			return "", nil, claimReadError(pod, "ResourceClaimTemplate", templateName, getErr)
		}
		return "template/" + claimRef.Name, &template.Spec.Spec, nil
	}

	return "", nil, nil
}

// claimReadError classifies a failed read of a claim or template. A missing object of a finished Pod is
// permanent (ErrClaimDeleted). Every other failure, including a missing object of an active Pod, is a lookup error to retry.
func claimReadError(pod *corev1.Pod, kind, name string, err error) error {
	if apierrors.IsNotFound(err) && (pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed) {
		return fmt.Errorf("%s %s/%s: %w", kind, pod.Namespace, name, ErrClaimDeleted)
	}
	return fmt.Errorf("getting %s %s/%s: %w", kind, pod.Namespace, name, err)
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
