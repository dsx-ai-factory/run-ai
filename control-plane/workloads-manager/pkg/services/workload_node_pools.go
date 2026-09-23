/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"strings"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	kartaresource "github.com/run-ai/karta/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// nodePoolsKarta returns the Karta used to read pod node pools, or nil when unavailable.
func nodePoolsKarta(ctx context.Context, workloadType *workloadsApi.WorkloadTypeConfig, version string) *kartav1alpha1.Karta {
	karta, err := decodeKartaForVersion(workloadType, version)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("no Karta to read workload node pools: %s", err.Message)
		return nil
	}
	return karta
}

// requestedWorkloadNodePools returns the node pools selected by the workload's
// pod templates, or nil when none can be read.
func requestedWorkloadNodePools(
	ctx context.Context,
	karta *kartav1alpha1.Karta,
	resource map[string]interface{},
	actualNodePools []clusterApi.Nodepool,
) []string {
	if karta == nil {
		return nil
	}

	factory := kartaresource.NewComponentFactoryFromObject(karta, &unstructured.Unstructured{Object: resource})
	components, err := networkingKartaComponents(factory)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to read Karta components for node pools: %v", err)
		return nil
	}

	var requested []string
	for _, component := range components {
		instances, err := component.GetExtractedInstances(ctx)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to extract Karta component %q for node pools: %v", component.Name(), err)
			return nil
		}
		for _, instance := range instances {
			requested = mergeNodePoolNames(requested, podNodePools(instance, actualNodePools))
		}
	}
	return requested
}

// podNodePools follows the cluster pod mutator's precedence: node-pools
// annotation, then node-pool label, then required node affinity.
func podNodePools(instance kartaresource.ExtractedInstance, actualNodePools []clusterApi.Nodepool) []string {
	var labels, annotations map[string]string
	var nodeAffinity *corev1.NodeAffinity
	switch {
	case instance.PodTemplateSpec != nil:
		labels, annotations = instance.PodTemplateSpec.Labels, instance.PodTemplateSpec.Annotations
		if instance.PodTemplateSpec.Spec.Affinity != nil {
			nodeAffinity = instance.PodTemplateSpec.Spec.Affinity.NodeAffinity
		}
	case instance.PodSpec != nil:
		if instance.PodSpec.Affinity != nil {
			nodeAffinity = instance.PodSpec.Affinity.NodeAffinity
		}
	case instance.FragmentedPodSpec != nil:
		labels, annotations = instance.FragmentedPodSpec.Labels, instance.FragmentedPodSpec.Annotations
		nodeAffinity = instance.FragmentedPodSpec.NodeAffinity
	}

	if annotated := knownNodePools(strings.Fields(annotations[commonconstants.AnnotationNodepools]), actualNodePools); len(annotated) > 0 {
		return annotated
	}
	if labeled := knownNodePools([]string{labels[clusterconstants.NodePoolNameLabel]}, actualNodePools); len(labeled) > 0 {
		return labeled
	}
	return nodePoolsFromNodeAffinity(nodeAffinity, actualNodePools)
}

func knownNodePools(names []string, actualNodePools []clusterApi.Nodepool) []string {
	var known []string
	for _, name := range names {
		for _, nodePool := range actualNodePools {
			if nodePool.GetName() == name {
				known = mergeNodePoolNames(known, []string{name})
				break
			}
		}
	}
	return known
}

func nodePoolsFromNodeAffinity(nodeAffinity *corev1.NodeAffinity, actualNodePools []clusterApi.Nodepool) []string {
	if nodeAffinity == nil || nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		return nil
	}

	nodePoolByLabel := make(map[string]map[string]string)
	for _, nodePool := range actualNodePools {
		if nodePool.GetPhase() == clusterApi.NODEPOOLPHASE_DELETING || nodePool.GetLabelKey() == "" {
			continue
		}
		if nodePoolByLabel[nodePool.GetLabelKey()] == nil {
			nodePoolByLabel[nodePool.GetLabelKey()] = make(map[string]string)
		}
		nodePoolByLabel[nodePool.GetLabelKey()][nodePool.GetLabelValue()] = nodePool.GetName()
	}

	var requested []string
	for _, term := range nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		for _, expression := range term.MatchExpressions {
			if nodePools, matched := nodePoolsForSelectorRequirement(expression, nodePoolByLabel); matched {
				// The cluster resolver uses only the first node-pool expression per term.
				requested = mergeNodePoolNames(requested, nodePools)
				break
			}
		}
	}
	return requested
}
