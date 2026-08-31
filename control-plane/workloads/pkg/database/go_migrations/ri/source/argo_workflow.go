/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package source

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

func GetArgoWorkflowRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildArgoWorkflowV1alpha1(),
		},
	}
}

func BuildArgoWorkflowV1alpha1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "workflow",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "argoproj.io",
						Version: "v1alpha1",
						Kind:    "Workflow",
					},
					StatusDefinition: buildArgoWorkflowStatusDefinition(),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
							PodAffinityPath:   ptr.To(".spec.affinity.podAffinity"),
							NodeAffinityPath:  ptr.To(".spec.affinity.nodeAffinity"),
							SchedulerNamePath: ptr.To(".spec.schedulerName"),
						},
						MetadataPath: ptr.To("spec.workflowMetadata"),
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To(".spec.parallelism"),
					},
				},
				ChildComponents: []kartav1alpha1.ComponentDefinition{{
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "",
						Version: "v1",
						Kind:    "Pod",
					},
					Name: "templates[].name",
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
							PodAffinityPath:       ptr.To("spec.templates[].affinity.podAffinity"),
							NodeAffinityPath:      ptr.To("spec.templates[].affinity.nodeAffinity"),
							AnnotationsPath:       ptr.To("spec.templates[].metadata.annotations"),
							LabelsPath:            ptr.To("spec.templates[].metadata.labels"),
							PriorityClassNamePath: ptr.To("spec.templates[].priorityClassName"),
							SchedulerNamePath:     ptr.To("spec.templates[].schedulerName"),
						},
						MetadataPath: ptr.To("spec.templates[].metadata"),
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To(".spec.templates[].parallelism"),
					},
					OwnerRef: ptr.To("workflow"),
				},
				},
			},
		},
	}
}

func buildArgoWorkflowStatusDefinition() *kartav1alpha1.StatusDefinition {
	return &kartav1alpha1.StatusDefinition{
		PhaseDefinition: &kartav1alpha1.PhaseDefinition{
			Path: ".status.phase",
		},
		ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
			Path:            ".status.conditions",
			TypeFieldName:   "type",
			StatusFieldName: "status",
		},
		StatusMappings: kartav1alpha1.StatusMappings{
			Running: []kartav1alpha1.StatusMatcher{
				{ByPhase: "Running"},
			},
			Completed: []kartav1alpha1.StatusMatcher{
				{ByPhase: "Succeeded"},
			},
			Failed: []kartav1alpha1.StatusMatcher{
				{ByPhase: "Failed"},
				{ByPhase: "Error"},
			},
		},
	}
}
