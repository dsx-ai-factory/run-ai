/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
)

// runaiDeletionBlockers is the runai-specific set of project-deletion blockers,
// reproducing the historically hard-coded Workload/DataVolume/Secret/PVC set.
// KRM's operator writes them into the ConfigMap its project-controller reads, so
// this is the single source of truth for the runai packaging.
//
// Blockers sharing a displayName (Workloads, Secrets) are grouped by the
// controller into a single project condition.
//
// A blocker the controller lacks RBAC to list is an error that blocks deletion
// outright, so the chart must keep granting get/list/watch on every kind here.
func runaiDeletionBlockers() []kaiv1alpha1.DeleteBlocker {
	const (
		runaiGroup = "run.ai"
		v2alpha1   = "v2alpha1"
		v1alpha1   = "v1alpha1"
	)
	assetKinds := []string{"password", "access-key", "generic", "docker-registry"}

	// projectScopeExpressions excludes department / cluster-wide / tenant-wide scoped
	// resources, matching common.getProjectScopeLabelSelector in project-controller.
	projectScopeExpressions := []metav1.LabelSelectorRequirement{
		{Key: clustercommonconstants.DepartmentScopeLabel, Operator: metav1.LabelSelectorOpDoesNotExist},
		{Key: clustercommonconstants.ClusterWideResourceLabel, Operator: metav1.LabelSelectorOpDoesNotExist},
		{Key: clustercommonconstants.TenantScopeLabel, Operator: metav1.LabelSelectorOpDoesNotExist},
	}

	projectScopeWithType := func(typeLabel string) *metav1.LabelSelector {
		exprs := append(projectScopeExpressions, metav1.LabelSelectorRequirement{
			Key: typeLabel, Operator: metav1.LabelSelectorOpIn, Values: assetKinds,
		})
		return &metav1.LabelSelector{MatchExpressions: exprs}
	}

	return []kaiv1alpha1.DeleteBlocker{
		{DisplayName: "Workloads", Group: runaiGroup, Version: v2alpha1, Kind: "TrainingWorkload"},
		{DisplayName: "Workloads", Group: runaiGroup, Version: v2alpha1, Kind: "InferenceWorkload"},
		{DisplayName: "Workloads", Group: runaiGroup, Version: v2alpha1, Kind: "DistributedWorkload"},
		{DisplayName: "Workloads", Group: runaiGroup, Version: v2alpha1, Kind: "InteractiveWorkload"},
		{DisplayName: "Workloads", Group: runaiGroup, Version: v1alpha1, Kind: "ExternalWorkload"},
		{DisplayName: "Workloads", Group: runaiGroup, Version: v1alpha1, Kind: "DistributedInferenceWorkload"},
		{DisplayName: "Workloads", Group: runaiGroup, Version: "v1", Kind: "RunaiJob"},
		{DisplayName: "Workloads", Group: runaiGroup, Version: v1alpha1, Kind: "WorkloadRunner"},
		{DisplayName: "DataVolumes", Group: runaiGroup, Version: v1alpha1, Kind: "DataVolume"},
		{DisplayName: "Secrets", Group: "", Version: "v1", Kind: "Secret",
			LabelSelector: projectScopeWithType(clustercommonconstants.RunaiResourceLabel)},
		{DisplayName: "Secrets", Group: "", Version: "v1", Kind: "Secret",
			LabelSelector: projectScopeWithType(clustercommonconstants.RunaiAssetLabel)},
		{DisplayName: "Pvcs", Group: "", Version: "v1", Kind: "PersistentVolumeClaim",
			LabelSelector: &metav1.LabelSelector{MatchExpressions: projectScopeExpressions}},
	}
}
