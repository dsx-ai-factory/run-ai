/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nvidia

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/go_migrations/ri/source"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

func GetNIMServiceRIConfiguration() source.KartaConfiguration {
	return source.KartaConfiguration{
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildNIMServiceV1alpha1(),
		},
	}
}

func BuildNIMServiceV1alpha1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "nimservice",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "apps.nvidia.com",
						Version: "v1alpha1",
						Kind:    "NIMService",
					},
					StatusDefinition: buildNIMServiceStatusDefinition(),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
							ResourcesPath:     ptr.To(".spec.resources"),
							PodAffinityPath:   ptr.To(".spec.affinity.podAffinity"),
							NodeAffinityPath:  ptr.To(".spec.affinity.nodeAffinity"),
							AnnotationsPath:   ptr.To(".spec.annotations"),
							LabelsPath:        ptr.To(".spec.labels"),
							SchedulerNamePath: ptr.To(".spec.schedulerName"),
						},
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To(".spec.replicas // 1"),
					},
				},
			},
		},
	}
}

func buildNIMServiceStatusDefinition() *kartav1alpha1.StatusDefinition {
	return &kartav1alpha1.StatusDefinition{
		PhaseDefinition: &kartav1alpha1.PhaseDefinition{
			Path: ".status.state",
		},
		ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
			Path:             ".status.conditions",
			TypeFieldName:    "type",
			StatusFieldName:  "status",
			MessageFieldName: ptr.To("message"),
		},
		StatusMappings: kartav1alpha1.StatusMappings{
			Initializing: []kartav1alpha1.StatusMatcher{
				{ByPhase: "NotReady"},
			},
			Running: []kartav1alpha1.StatusMatcher{
				{ByPhase: "Ready"},
			},
			Failed: []kartav1alpha1.StatusMatcher{
				{ByPhase: "Failed"},
			},
		},
	}
}
