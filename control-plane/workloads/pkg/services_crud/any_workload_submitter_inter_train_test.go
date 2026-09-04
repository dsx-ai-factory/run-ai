/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2/iw_tw_workloads"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	workloadMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
)

func TestBuildInterTrainWorkloadV2Metadata(t *testing.T) {
	t.Run("passes v1 category priority and preemptibility to workloads-manager metadata", func(t *testing.T) {
		name := "ws-1"
		projectId := "project-1"
		category := "Build"
		priorityClass := "high"
		preemptibility := workloadsApi.PREEMPTIBILITY_NON_PREEMPTIBLE

		spec := &workloadsApi.SupersetSpec{
			Category:       *workloadsApi.NewNullableString(&category),
			PriorityClass:  *workloadsApi.NewNullableString(&priorityClass),
			Preemptibility: *workloadsApi.NewNullablePreemptibility(&preemptibility),
		}

		metadata := buildInterTrainWorkloadV2Metadata(&name, &projectId, &workloadMgrApi.WorkloadV2Configuration{}, spec, nil)

		if metadata.Name == nil || *metadata.Name != name {
			t.Fatalf("expected name %q, got %v", name, metadata.Name)
		}
		if metadata.ProjectId == nil || *metadata.ProjectId != projectId {
			t.Fatalf("expected projectId %q, got %v", projectId, metadata.ProjectId)
		}
		if metadata.Category.Get() == nil || *metadata.Category.Get() != category {
			t.Fatalf("expected category %q, got %v", category, metadata.Category.Get())
		}
		if metadata.Priority.Get() == nil || *metadata.Priority.Get() != priorityClass {
			t.Fatalf("expected priority %q, got %v", priorityClass, metadata.Priority.Get())
		}
		if metadata.Preemptibility.Get() == nil || *metadata.Preemptibility.Get() != workloadMgrApi.PREEMPTIBILITY_NON_PREEMPTIBLE {
			t.Fatalf("expected preemptibility %q, got %v", workloadMgrApi.PREEMPTIBILITY_NON_PREEMPTIBLE, metadata.Preemptibility.Get())
		}
	})

	t.Run("leaves metadata unset when v1 did not set the values", func(t *testing.T) {
		metadata := buildInterTrainWorkloadV2Metadata(nil, nil, nil, &workloadsApi.SupersetSpec{}, nil)

		if metadata.Category.IsSet() {
			t.Fatalf("expected category to be unset")
		}
		if metadata.Priority.IsSet() {
			t.Fatalf("expected priority to be unset")
		}
		if metadata.Preemptibility.IsSet() {
			t.Fatalf("expected preemptibility to be unset")
		}
	})

	t.Run("uses composed job values when defaults were applied by the v1 composer", func(t *testing.T) {
		category := "Build"
		priorityClass := "high"
		preemptibility := "non-preemptible"
		target := iw_tw_workloads.NewMinimalBatchJobTarget()
		target.Job.Labels = map[string]string{composer.WorkloadCategoryLabel: category}
		target.Job.Spec.Template.Labels = map[string]string{composer.LabelPreemptibilityName: preemptibility}
		target.Job.Spec.Template.Spec.PriorityClassName = priorityClass

		metadata := buildInterTrainWorkloadV2Metadata(nil, nil, nil, &workloadsApi.SupersetSpec{}, target)

		if metadata.Category.Get() == nil || *metadata.Category.Get() != category {
			t.Fatalf("expected category %q, got %v", category, metadata.Category.Get())
		}
		if metadata.Priority.Get() == nil || *metadata.Priority.Get() != priorityClass {
			t.Fatalf("expected priority %q, got %v", priorityClass, metadata.Priority.Get())
		}
		if metadata.Preemptibility.Get() == nil || *metadata.Preemptibility.Get() != workloadMgrApi.PREEMPTIBILITY_NON_PREEMPTIBLE {
			t.Fatalf("expected preemptibility %q, got %v", workloadMgrApi.PREEMPTIBILITY_NON_PREEMPTIBLE, metadata.Preemptibility.Get())
		}
	})
}
