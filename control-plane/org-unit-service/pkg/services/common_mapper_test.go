/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"k8s.io/utils/ptr"
)

// Without a parent there is nothing to inherit, so every effective value has to
// mirror the org unit's own value.
func TestMapEffectiveFieldsWithoutParentMirrorsTheOrgUnitData(t *testing.T) {
	interactiveMaxIdle := int32(11)
	trainingMaxIdle := int32(22)
	nodeTypes := []models.NodeType{{Name: "gpu-a"}}
	nodePools := pq.StringArray{"pool-a"}

	data := models.OrgUnitData{
		InteractiveJobTimeLimit:           ptr.To(int32(1)),
		InteractiveJobMaxIdleDuration:     &interactiveMaxIdle,
		InteractiveJobPreemptIdleDuration: ptr.To(int32(2)),
		TrainingJobTimeLimit:              ptr.To(int32(3)),
		TrainingJobMaxIdleDuration:        &trainingMaxIdle,
		DefaultNodePoolList:               &nodePools,
		TrainingNodeTypes:                 &nodeTypes,
		WorkspaceNodeTypes:                &nodeTypes,
	}

	got := MapEffectiveFields(uuid.Nil, nil, data, false)

	if got.InteractiveJobTimeLimitEffective != data.InteractiveJobTimeLimit {
		t.Errorf("interactive job time limit = %v, want %v", got.InteractiveJobTimeLimitEffective, data.InteractiveJobTimeLimit)
	}
	if got.InteractiveJobMaxIdleDurationEffective != data.InteractiveJobMaxIdleDuration {
		t.Errorf("interactive job max idle duration = %v, want the interactive value %v",
			got.InteractiveJobMaxIdleDurationEffective, data.InteractiveJobMaxIdleDuration)
	}
	if got.InteractiveJobPreemptIdleDurationEffective != data.InteractiveJobPreemptIdleDuration {
		t.Errorf("interactive job preempt idle duration = %v, want %v", got.InteractiveJobPreemptIdleDurationEffective, data.InteractiveJobPreemptIdleDuration)
	}
	if got.TrainingJobTimeLimitEffective != data.TrainingJobTimeLimit {
		t.Errorf("training job time limit = %v, want %v", got.TrainingJobTimeLimitEffective, data.TrainingJobTimeLimit)
	}
	if got.TrainingJobMaxIdleDurationEffective != data.TrainingJobMaxIdleDuration {
		t.Errorf("training job max idle duration = %v, want %v", got.TrainingJobMaxIdleDurationEffective, data.TrainingJobMaxIdleDuration)
	}
	if got.DefaultNodePoolListEffective != data.DefaultNodePoolList {
		t.Errorf("default node pool list = %v, want %v", got.DefaultNodePoolListEffective, data.DefaultNodePoolList)
	}
	if got.TrainingNodeTypesEffective != data.TrainingNodeTypes {
		t.Errorf("training node types = %v, want %v", got.TrainingNodeTypesEffective, data.TrainingNodeTypes)
	}
	if got.WorkspaceNodeTypesEffective != data.WorkspaceNodeTypes {
		t.Errorf("workspace node types = %v, want %v", got.WorkspaceNodeTypesEffective, data.WorkspaceNodeTypes)
	}
	for _, origin := range []struct {
		name  string
		value *uuid.UUID
	}{
		{"interactive job time limit", got.InteractiveJobTimeLimitEffectiveOrigin},
		{"interactive job max idle duration", got.InteractiveJobMaxIdleDurationEffectiveOrigin},
		{"interactive job preempt idle duration", got.InteractiveJobPreemptIdleDurationEffectiveOrigin},
		{"training job time limit", got.TrainingJobTimeLimitEffectiveOrigin},
		{"training job max idle duration", got.TrainingJobMaxIdleDurationEffectiveOrigin},
		{"default node pool list", got.DefaultNodePoolListEffectiveOrigin},
		{"training node types", got.TrainingNodeTypesEffectiveOrigin},
		{"workspace node types", got.WorkspaceNodeTypesEffectiveOrigin},
	} {
		if origin.value != nil {
			t.Errorf("%s origin = %v, want nil without a parent", origin.name, *origin.value)
		}
	}
}

// A department gets no default node pool injected, and that is the only thing
// isDepartment changes for an org unit without a parent.
func TestMapEffectiveFieldsWithoutParentKeepsAnEmptyNodePoolList(t *testing.T) {
	empty := pq.StringArray{}

	for _, tc := range []struct {
		name         string
		isDepartment bool
		want         pq.StringArray
	}{
		{name: "department", isDepartment: true, want: empty},
		{name: "project", isDepartment: false, want: pq.StringArray{defaultNodePoolName}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := MapEffectiveFields(uuid.Nil, nil, models.OrgUnitData{DefaultNodePoolList: &empty}, tc.isDepartment)
			if got.DefaultNodePoolListEffective == nil {
				t.Fatal("default node pool list is nil, want a value")
			}
			if len(*got.DefaultNodePoolListEffective) != len(tc.want) {
				t.Fatalf("default node pool list = %v, want %v", *got.DefaultNodePoolListEffective, tc.want)
			}
			for i := range tc.want {
				if (*got.DefaultNodePoolListEffective)[i] != tc.want[i] {
					t.Fatalf("default node pool list = %v, want %v", *got.DefaultNodePoolListEffective, tc.want)
				}
			}
		})
	}
}
