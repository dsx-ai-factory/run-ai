/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

const NvidiaProductLabel = "nvidia.com/gpu.product"

// RequestedAccelerator returns the effective override for compute composition.
// A model-only request explicitly selects the default accelerator; this matters
// when checking for duplicate quantities in extendedResources.
func RequestedAccelerator(model *string, accelerator *workloadsApi.AcceleratorSelection) *workloadsApi.AcceleratorSelection {
	if model != nil && accelerator == nil {
		return &workloadsApi.AcceleratorSelection{}
	}
	return accelerator
}

// ResolveGPUSelection accepts the effective model label key resolved by discovery.
// Unavailable label values are valid placement requirements.
func ResolveGPUSelection(model *string, accelerator *workloadsApi.AcceleratorSelection) (string, *corev1.NodeSelectorRequirement, *workloadsApi.Error) {
	resourceName := NvidiaResourceName
	var labelKey *string
	if accelerator != nil {
		resourceName = convutils.Deref(accelerator.ResourceName.Get(), NvidiaResourceName)
		labelKey = accelerator.DeviceLabelKey.Get()
	}
	if !strings.Contains(resourceName, "/") || strings.Contains(resourceName, corev1.ResourceDefaultNamespacePrefix) ||
		strings.HasPrefix(resourceName, corev1.DefaultResourceRequestsPrefix) || len(validation.IsQualifiedName(corev1.DefaultResourceRequestsPrefix+resourceName)) > 0 {
		return "", nil, workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("accelerator.resourceName %q is not a Kubernetes extended resource name", resourceName))
	}
	if labelKey != nil && len(validation.IsQualifiedName(*labelKey)) > 0 {
		return "", nil, workloadsApi.NewError(http.StatusBadRequest, "accelerator.deviceLabelKey must be a valid Kubernetes label key")
	}
	if model == nil {
		return resourceName, nil, nil
	}
	if *model == "" || len(validation.IsValidLabelValue(*model)) > 0 {
		return "", nil, workloadsApi.NewError(http.StatusBadRequest, "gpuType must be a nonempty Kubernetes label value")
	}
	if labelKey == nil {
		if resourceName != NvidiaResourceName {
			return "", nil, workloadsApi.NewError(http.StatusBadRequest, "accelerator.deviceLabelKey is required when the GPU model label key cannot be determined")
		}
		labelKey = convutils.Ref(NvidiaProductLabel)
	}
	return resourceName, &corev1.NodeSelectorRequirement{Key: *labelKey, Operator: corev1.NodeSelectorOpIn, Values: []string{*model}}, nil
}

// ComposeGPUPlacement intersects the resolved device predicate with each existing
// alternative. A present empty selector or term matches no nodes and stays empty.
func (fc *FieldsComposer) ComposeGPUPlacement(model *string, accelerator *workloadsApi.AcceleratorSelection, pools []string, pod *corev1.PodSpec, meta *metav1.ObjectMeta) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if model == nil && accelerator == nil {
			return nil
		}
		resourceName, requirement, err := ResolveGPUSelection(model, accelerator)
		if err != nil {
			return err
		}
		if !fc.Options.NodeAffinitySupported {
			return workloadsApi.NewError(http.StatusBadRequest, "accelerator selection requires node affinity support")
		}

		// KAI accounts NVIDIA and AMD in the same GPU slot. Constrain the
		// physical resource advertised by each candidate node as well.
		knative := fc.WorkloadType == Inference
		if knative {
			// Knative installations require affinity support but may disable
			// podspec-nodeselector. Preserve existing equalities in required affinity.
			for _, key := range slices.Sorted(maps.Keys(pod.NodeSelector)) {
				intersectGPURequirement(pod, corev1.NodeSelectorRequirement{
					Key: key, Operator: corev1.NodeSelectorOpIn, Values: []string{pod.NodeSelector[key]},
				})
			}
			pod.NodeSelector = nil
		}

		requireLabel := func(req corev1.NodeSelectorRequirement) {
			intersectGPURequirement(pod, req)
			// Also protect the equality from older Pod mutators that replace
			// same-key expressions. Never remove an existing conflicting selector.
			if !knative {
				if pod.NodeSelector == nil {
					pod.NodeSelector = make(map[string]string)
				}
				if _, exists := pod.NodeSelector[req.Key]; !exists {
					pod.NodeSelector[req.Key] = req.Values[0]
				}
			}
		}
		requireLabel(corev1.NodeSelectorRequirement{
			Key: commonconstants.AcceleratorResourceNodeLabel(resourceName), Operator: corev1.NodeSelectorOpIn, Values: []string{"true"},
		})
		if requirement != nil {
			requireLabel(*requirement)
			metav1.SetMetaDataAnnotation(meta, commonconstants.AnnotationGPUTypeLabel, requirement.Key+"="+*model)
		}
		// Older admission replaces same-key scheduling requirements and may
		// skip project-pool checks. Keep the template unschedulable until
		// upgraded Pod admission has enforced the actual resource, model and pools.
		alreadyGuarded := meta.Annotations[commonconstants.AnnotationGPUPlacementResource] == resourceName
		metav1.SetMetaDataAnnotation(meta, commonconstants.AnnotationGPUPlacementResource, resourceName)
		terms := pod.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		for i := range terms {
			fields := terms[i].MatchFields
			if alreadyGuarded && len(fields) >= 2 && fields[0].Key == "metadata.name" && fields[1].Key == "metadata.name" &&
				fields[0].Operator == corev1.NodeSelectorOpIn && fields[1].Operator == corev1.NodeSelectorOpNotIn &&
				len(fields[0].Values) == 1 && slices.Equal(fields[0].Values, fields[1].Values) &&
				(fields[0].Values[0] == commonconstants.GPUPlacementAdmissionNode || fields[0].Values[0] == commonconstants.GPUPlacementAdmissionNode+"-empty") {
				continue
			}
			node := commonconstants.GPUPlacementAdmissionNode
			if len(terms[i].MatchFields) == 0 && len(terms[i].MatchExpressions) == 0 {
				// An originally empty term must remain unsatisfiable after admission.
				node += "-empty"
			}
			terms[i].MatchFields = append([]corev1.NodeSelectorRequirement{
				{Key: "metadata.name", Operator: corev1.NodeSelectorOpIn, Values: []string{node}},
				{Key: "metadata.name", Operator: corev1.NodeSelectorOpNotIn, Values: []string{node}},
			}, terms[i].MatchFields...)
		}
		if len(pools) > 0 {
			metav1.SetMetaDataAnnotation(meta, commonconstants.AnnotationNodepools, strings.Join(pools, " "))
		}
		return nil
	}
}

func intersectGPURequirement(pod *corev1.PodSpec, requirement corev1.NodeSelectorRequirement) {
	if pod.Affinity == nil {
		pod.Affinity = &corev1.Affinity{}
	}
	if pod.Affinity.NodeAffinity == nil {
		pod.Affinity.NodeAffinity = &corev1.NodeAffinity{}
	}
	affinity := pod.Affinity.NodeAffinity
	if affinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		affinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{requirement}}},
		}
		return
	}
	if len(affinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0 {
		// An explicit empty selector matches nothing. The cluster's affinity
		// merger preserves an empty term across snapshot restoration.
		affinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = []corev1.NodeSelectorTerm{{}}
	}
	for i := range affinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		term := &affinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[i]
		if len(term.MatchExpressions) == 0 && len(term.MatchFields) == 0 {
			continue
		}
		found := slices.ContainsFunc(term.MatchExpressions, func(expr corev1.NodeSelectorRequirement) bool {
			return expr.Key == requirement.Key && expr.Operator == requirement.Operator && slices.Equal(expr.Values, requirement.Values)
		})
		if !found {
			term.MatchExpressions = append(term.MatchExpressions, requirement)
		}
	}
}

// intersectNodePools combines pool alternatives with user constraints. Each
// original term stays intact, including conflicting values on the same key.
func intersectNodePools(affinity *corev1.NodeAffinity, pools []corev1.NodeSelectorRequirement) {
	required := affinity.RequiredDuringSchedulingIgnoredDuringExecution
	if required == nil {
		required = &corev1.NodeSelector{}
		for _, pool := range pools {
			required.NodeSelectorTerms = append(required.NodeSelectorTerms, corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{pool}})
		}
		affinity.RequiredDuringSchedulingIgnoredDuringExecution = required
		return
	}
	terms := make([]corev1.NodeSelectorTerm, 0, len(pools)*len(required.NodeSelectorTerms))
	for _, term := range required.NodeSelectorTerms {
		if len(term.MatchFields) == 0 && len(term.MatchExpressions) == 0 {
			terms = append(terms, term)
			continue
		}
		for _, pool := range pools {
			exprs := append(slices.Clone(term.MatchExpressions), pool)
			terms = append(terms, corev1.NodeSelectorTerm{MatchExpressions: exprs, MatchFields: term.MatchFields})
		}
	}
	required.NodeSelectorTerms = terms
}
