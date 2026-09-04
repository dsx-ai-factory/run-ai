/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package commonparsers

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/resourceutils/resources"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HandleK8sPodControllerSpec - a common handler for k8s pod controllers (deployment, daemonset, statefulset, etc.)
// only fills the externalworkload spec; status should be handled by the specific handler.
// returns the number of requested pods.
func HandleK8sPodControllerSpec(ctx context.Context, topOwnerObjectMeta metav1.ObjectMeta, topOwnerGroupKind schema.GroupKind, specReplicas *int32,
	podTemplate *corev1.PodTemplateSpec, cachedReader client.Reader, draParser *dra.DRAParser, externalWorkload *v1alpha1.ExternalWorkload) int32 {
	FillExternalWorkloadDataFromPodTemplateSpec(ctx, topOwnerObjectMeta, topOwnerGroupKind, podTemplate, nil, cachedReader, draParser, externalWorkload)

	numberOfRequestedPods := getNumberOfRequestedPodsForK8sPodController(specReplicas)
	externalWorkload.Spec.RequestedPods.Replicas = ptr.To(numberOfRequestedPods)

	MultiplyRequestedResourcesPerNumberOfRequestedPods(numberOfRequestedPods, externalWorkload)
	return numberOfRequestedPods
}

// FillExternalWorkloadDataFromPodTemplateSpec fills the ExternalWorkload spec from a top-owner pod template.
// specName identifies which logical pod template the affinity belongs to (e.g. "head", "master").
// Pass nil for single-template workloads (Deployment, StatefulSet, Job, etc.) to default to
// v1alpha1.DefaultNodeAffinitySpecName ("worker"). Multi-template workloads (Ray head/worker,
// kubeflow master/worker, etc.) pass an explicit name so each template lands on a distinct entry.
func FillExternalWorkloadDataFromPodTemplateSpec(ctx context.Context, topOwnerObjectMeta metav1.ObjectMeta, topOwnerGroupKind schema.GroupKind,
	podTemplate *corev1.PodTemplateSpec, specName *string, cachedReader client.Reader, draParser *dra.DRAParser, externalWorkload *v1alpha1.ExternalWorkload) {
	fillContainersInfo(ctx, podTemplate, externalWorkload)
	resolveDRAGpus(ctx, podTemplate, draParser, externalWorkload)

	externalWorkload.Spec.PriorityClassName, externalWorkload.Spec.Preemptibility = CalculatePriorityClassNameAndPreemptibility(ctx, cachedReader,
		topOwnerObjectMeta.Labels, topOwnerGroupKind, podTemplate.Spec.PriorityClassName)

	FillRequestedNodePools(ctx, topOwnerObjectMeta, topOwnerGroupKind.Kind, podTemplate.Spec.Affinity, cachedReader, externalWorkload)

	nodeAffinitySpecName := ptr.Deref(specName, v1alpha1.DefaultNodeAffinitySpecName)
	UpsertNodeAffinityFromPodTemplate(nodeAffinitySpecName, podTemplate, externalWorkload)

	SetStorageRefsFromPodTemplate(ctx, podTemplate, cachedReader, externalWorkload)
}

func FillExternalWorkloadDataFromPodTemplateSpecIfEmpty(ctx context.Context, topOwnerObjectMeta metav1.ObjectMeta, topOwnerGroupKind schema.GroupKind,
	podTemplate *corev1.PodTemplateSpec, specName *string, cachedReader client.Reader, draParser *dra.DRAParser, externalWorkload *v1alpha1.ExternalWorkload) {
	fillContainersInfoIfEmpty(ctx, podTemplate, externalWorkload)
	resolveDRAGpus(ctx, podTemplate, draParser, externalWorkload)

	if externalWorkload.Spec.PriorityClassName == "" || externalWorkload.Spec.Preemptibility == "" {
		priorityClassName, preemptibility := CalculatePriorityClassNameAndPreemptibility(ctx, cachedReader,
			topOwnerObjectMeta.Labels, topOwnerGroupKind, podTemplate.Spec.PriorityClassName)
		if externalWorkload.Spec.PriorityClassName == "" {
			externalWorkload.Spec.PriorityClassName = priorityClassName
		}
		if externalWorkload.Spec.Preemptibility == "" {
			externalWorkload.Spec.Preemptibility = preemptibility
		}
	}

	FillRequestedNodePools(ctx, topOwnerObjectMeta, topOwnerGroupKind.Kind, podTemplate.Spec.Affinity, cachedReader, externalWorkload)

	nodeAffinitySpecName := ptr.Deref(specName, v1alpha1.DefaultNodeAffinitySpecName)
	if !hasNodeAffinityEntry(externalWorkload, nodeAffinitySpecName) {
		UpsertNodeAffinityFromPodTemplate(nodeAffinitySpecName, podTemplate, externalWorkload)
	}

	UpsertStorageRefsFromPodTemplate(ctx, podTemplate, cachedReader, externalWorkload)
}

// UpsertNodeAffinityFromPodTemplate copies the pod template's nodeAffinity into the
// matching specName entry on externalWorkload.Spec.NodeAffinities, replacing any
// previous entry with the same SpecName. When the pod template has no node affinity,
// any existing entry for that SpecName is removed so stale data is not retained.
func UpsertNodeAffinityFromPodTemplate(specName string, podTemplate *corev1.PodTemplateSpec, externalWorkload *v1alpha1.ExternalWorkload) {
	if specName == "" {
		specName = v1alpha1.DefaultNodeAffinitySpecName
	}

	var na *corev1.NodeAffinity
	if podTemplate != nil && podTemplate.Spec.Affinity != nil && podTemplate.Spec.Affinity.NodeAffinity != nil {
		na = podTemplate.Spec.Affinity.NodeAffinity.DeepCopy()
	}

	for i := range externalWorkload.Spec.NodeAffinities {
		if externalWorkload.Spec.NodeAffinities[i].SpecName != specName {
			continue
		}
		if na == nil {
			externalWorkload.Spec.NodeAffinities = append(externalWorkload.Spec.NodeAffinities[:i],
				externalWorkload.Spec.NodeAffinities[i+1:]...)
			return
		}
		externalWorkload.Spec.NodeAffinities[i].NodeAffinity = na
		return
	}
	// new specName
	if na != nil {
		externalWorkload.Spec.NodeAffinities = append(externalWorkload.Spec.NodeAffinities,
			v1alpha1.NodeAffinityEntry{SpecName: specName, NodeAffinity: na})
	}
}

func hasNodeAffinityEntry(externalWorkload *v1alpha1.ExternalWorkload, specName string) bool {
	if specName == "" {
		specName = v1alpha1.DefaultNodeAffinitySpecName
	}
	for i := range externalWorkload.Spec.NodeAffinities {
		if externalWorkload.Spec.NodeAffinities[i].SpecName == specName {
			return true
		}
	}
	return false
}

// SetStorageRefsFromPodTemplate replaces externalWorkload.Spec.StorageRefs with the refs
// resolved from a single pod template, dropping any left over from a previous reconcile
// (e.g. a detached PVC/Secret/ConfigMap or a deleted Ray worker group). It is the first
// storage-ref touch of a reconcile; multi-template workloads (e.g. Ray) then call
// UpsertStorageRefsFromPodTemplate for their remaining templates to union those in.
func SetStorageRefsFromPodTemplate(ctx context.Context, podTemplate *corev1.PodTemplateSpec,
	cachedReader client.Reader, externalWorkload *v1alpha1.ExternalWorkload) {
	externalWorkload.Spec.StorageRefs = nil
	UpsertStorageRefsFromPodTemplate(ctx, podTemplate, cachedReader, externalWorkload)
}

// UpsertStorageRefsFromPodTemplate resolves the PVCs, Secrets, and ConfigMaps referenced
// by the pod template (volumes and imagePullSecrets) to their live UIDs and unions them
// into externalWorkload.Spec.StorageRefs, deduping across templates so multi-template
// workloads (e.g. Ray head/worker) contribute a single combined set. It only appends, so
// the first template of a reconcile must go through SetStorageRefsFromPodTemplate to clear
// stale refs. Resolving here keeps the lookup in the EWI so cluster-sync passes the UID
// straight through.
func UpsertStorageRefsFromPodTemplate(ctx context.Context, podTemplate *corev1.PodTemplateSpec,
	cachedReader client.Reader, externalWorkload *v1alpha1.ExternalWorkload) {
	resolved := resolveStorageRefUIDs(ctx, cachedReader, externalWorkload.Namespace, extractStorageRefs(podTemplate))
	combined := append(externalWorkload.Spec.StorageRefs, resolved...)
	externalWorkload.Spec.StorageRefs = dedupeStorageRefs(combined)
}

// newObjectForKind returns an empty object of the given kind, or nil for an
// unrecognized kind so the caller can skip it instead of resolving against the
// wrong resource type.
func newObjectForKind(kind v1alpha1.StorageRefKind) client.Object {
	switch kind {
	case v1alpha1.StorageRefKindPVC:
		return &corev1.PersistentVolumeClaim{}
	case v1alpha1.StorageRefKindSecret:
		return &corev1.Secret{}
	case v1alpha1.StorageRefKindConfigMap:
		return &corev1.ConfigMap{}
	default:
		return nil
	}
}

// resolveStorageRefUIDs resolves each name+kind reference to its live Kubernetes UID via
// the cached reader. References whose resource is not found are skipped so the workload
// still syncs without them; they are picked up on a later reconcile once the resource exists.
func resolveStorageRefUIDs(ctx context.Context, cachedReader client.Reader, namespace string,
	refs []v1alpha1.StorageRef) []v1alpha1.StorageRef {
	resolved := make([]v1alpha1.StorageRef, 0, len(refs))
	for _, ref := range refs {
		obj := newObjectForKind(ref.Kind)
		if obj == nil {
			log.Warn().Msgf("Skipping storage ref <%s/%s> in namespace <%s>: unsupported kind",
				ref.Kind, ref.Name, namespace)
			continue
		}
		if err := cachedReader.Get(ctx, client.ObjectKey{Namespace: namespace, Name: ref.Name}, obj); err != nil {
			if client.IgnoreNotFound(err) != nil {
				log.Warn().Msgf("Failed to resolve storage ref <%s/%s> in namespace <%s>, error: %s",
					ref.Kind, ref.Name, namespace, err.Error())
			}
			continue
		}
		ref.UID = string(obj.GetUID())
		resolved = append(resolved, ref)
	}
	return resolved
}

// extractStorageRefs returns the deduped name+kind references in a single pod template.
// UID is resolved separately by resolveStorageRefUIDs.
func extractStorageRefs(podTemplate *corev1.PodTemplateSpec) []v1alpha1.StorageRef {
	if podTemplate == nil {
		return nil
	}
	var refs []v1alpha1.StorageRef
	for _, vol := range podTemplate.Spec.Volumes {
		switch {
		case vol.PersistentVolumeClaim != nil:
			refs = appendStorageRef(refs, v1alpha1.StorageRefKindPVC, vol.PersistentVolumeClaim.ClaimName)
		case vol.Secret != nil:
			refs = appendStorageRef(refs, v1alpha1.StorageRefKindSecret, vol.Secret.SecretName)
		case vol.ConfigMap != nil:
			refs = appendStorageRef(refs, v1alpha1.StorageRefKindConfigMap, vol.ConfigMap.Name)
		}
	}
	for _, ips := range podTemplate.Spec.ImagePullSecrets {
		refs = appendStorageRef(refs, v1alpha1.StorageRefKindSecret, ips.Name)
	}
	return dedupeStorageRefs(refs)
}

// appendStorageRef appends a name+kind reference, skipping empty names.
func appendStorageRef(refs []v1alpha1.StorageRef, kind v1alpha1.StorageRefKind, name string) []v1alpha1.StorageRef {
	if name == "" {
		return refs
	}
	return append(refs, v1alpha1.StorageRef{Kind: kind, Name: name})
}

// dedupeStorageRefs returns refs with duplicate name+kind pairs removed, preserving first-seen order.
func dedupeStorageRefs(refs []v1alpha1.StorageRef) []v1alpha1.StorageRef {
	if len(refs) == 0 {
		return nil
	}
	seen := make(map[v1alpha1.StorageRef]bool, len(refs))
	deduped := make([]v1alpha1.StorageRef, 0, len(refs))
	for _, ref := range refs {
		if _, ok := seen[ref]; !ok {
			seen[ref] = true
			deduped = append(deduped, ref)
		}
	}
	return deduped
}

func fillContainersInfoIfEmpty(ctx context.Context, podTemplate *corev1.PodTemplateSpec, externalWorkload *v1alpha1.ExternalWorkload) {
	if externalWorkload.Spec.ImageName == "" {
		CalculateImageInfo(podTemplate, externalWorkload)
	}

	if len(externalWorkload.Spec.RequestedResources.Limits) == 0 &&
		len(externalWorkload.Spec.RequestedResources.Requests) == 0 {
		requestedResources := GetPodRequestedResources(podTemplate)
		externalWorkload.Spec.RequestedResources = requestedResources
	}
}

func fillContainersInfo(ctx context.Context, podTemplate *corev1.PodTemplateSpec, externalWorkload *v1alpha1.ExternalWorkload) {
	requestedResources := GetPodRequestedResources(podTemplate)

	CalculateImageInfo(podTemplate, externalWorkload)

	externalWorkload.Spec.RequestedResources = requestedResources
}

func resolveDRAGpus(ctx context.Context, podTemplate *corev1.PodTemplateSpec, draParser *dra.DRAParser, externalWorkload *v1alpha1.ExternalWorkload) {
	if draParser != nil && len(podTemplate.Spec.ResourceClaims) > 0 {
		gpuCount := draParser.ResolveDRAGpuCount(ctx, externalWorkload.Namespace, podTemplate.Spec.ResourceClaims)
		dra.AddDRAGpuToResources(gpuCount, &externalWorkload.Spec.RequestedResources)
	}
}

func CalculateImageInfo(podTemplate *corev1.PodTemplateSpec, externalWorkload *v1alpha1.ExternalWorkload) {
	calculatedImage := ""
	for _, container := range podTemplate.Spec.Containers {
		if calculatedImage == "" {
			calculatedImage = container.Image
		} else if calculatedImage != container.Image {
			// containers have different images - use empty image
			calculatedImage = ""
			break
		}
	}
	externalWorkload.Spec.ImageName = calculatedImage
}

func GetPodRequestedResources(podTemplate *corev1.PodTemplateSpec) corev1.ResourceRequirements {
	requests, limits := resources.PodRunaiRequestsAndLimits(podTemplate.Spec.Containers,
		podTemplate.Spec.InitContainers, podTemplate.Annotations)
	return corev1.ResourceRequirements{Limits: limits, Requests: requests}
}

// getNumberOfRequestedPodsForK8sPodController - calculates the number of requested pods for an object
// that controls pods - deployment, replicaset, statefulset, etc..
func getNumberOfRequestedPodsForK8sPodController(replicas *int32) int32 {
	// default for k8s pod controllers is 1
	return utils.DerefInt32(replicas, 1)
}
