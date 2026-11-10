/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nvidia

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/go_migrations/ri/source"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/utils/ptr"
)

// GetSnapshotJobRIConfiguration makes a Snapshot capture its own workload, so its
// pod and GPU are attributed to it.
func GetSnapshotJobRIConfiguration() source.KartaConfiguration {
	return source.KartaConfiguration{
		Category:       "Deploy",
		Priority:       "low",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildSnapshotJobV1alpha1(),
		},
	}
}

func BuildSnapshotJobV1alpha1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "snapshotjob",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "nvidia.com",
						Version: "v1alpha1",
						Kind:    "SnapshotJob",
					},
					StatusDefinition: buildSnapshotJobStatusDefinition(),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: ptr.To(".spec.podTemplate"),
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To("1"),
					},
				},
			},
		},
	}
}

// Snapshot writes all four conditions on its first status write and never removes
// them, so False is as reliable as True. Each state matches one status: EWI ranks
// Running above Failed and Completed, so a capture that failed with its pod Ready
// would otherwise show as Running.
func buildSnapshotJobStatusDefinition() *kartav1alpha1.StatusDefinition {
	notFailed := kartav1alpha1.ExpectedCondition{Type: "Failed", Status: ptr.To("False")}

	return &kartav1alpha1.StatusDefinition{
		ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
			Path:             ".status.conditions",
			TypeFieldName:    "type",
			StatusFieldName:  "status",
			MessageFieldName: ptr.To("message"),
			ReasonFieldName:  ptr.To("reason"),
		},
		StatusMappings: kartav1alpha1.StatusMappings{
			Initializing: []kartav1alpha1.StatusMatcher{
				// Scheduling and warm-up, before the pod signals it is ready to capture.
				{ByConditions: []kartav1alpha1.ExpectedCondition{
					{Type: "Running", Status: ptr.To("False")},
					{Type: "Captured", Status: ptr.To("False")},
					notFailed,
				}},
				// Before Snapshot's first status write.
				{ByExpression: &kartav1alpha1.ExpressionMatcher{
					Expression:     "(.status.conditions // []) | length == 0",
					ExpectedResult: "true",
				}},
			},
			Running: []kartav1alpha1.StatusMatcher{
				{ByConditions: []kartav1alpha1.ExpectedCondition{
					{Type: "Running", Status: ptr.To("True")},
					notFailed,
				}},
				// Dumped, but the source Job's other containers are still finishing.
				{ByConditions: []kartav1alpha1.ExpectedCondition{
					{Type: "Captured", Status: ptr.To("True")},
					{Type: "Completed", Status: ptr.To("False")},
					notFailed,
				}},
			},
			Completed: []kartav1alpha1.StatusMatcher{
				{ByConditions: []kartav1alpha1.ExpectedCondition{
					{Type: "Completed", Status: ptr.To("True")},
				}},
			},
			Failed: []kartav1alpha1.StatusMatcher{
				{ByConditions: []kartav1alpha1.ExpectedCondition{
					{Type: "Failed", Status: ptr.To("True")},
				}},
			},
		},
	}
}
