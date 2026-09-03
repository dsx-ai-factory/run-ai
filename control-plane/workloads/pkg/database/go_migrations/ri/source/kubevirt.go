/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package source

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/utils/ptr"
)

func GetKubevirtVMIRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Train",
		Priority:       "low",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildKubevirtVMIV1(),
		},
	}
}

func BuildKubevirtVMIV1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "virtualmachineinstance",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "kubevirt.io",
						Version: "v1",
						Kind:    "VirtualMachineInstance",
					},
					StatusDefinition: buildVMIStatusDefinition(),
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To("1"),
					},
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
							ResourcesPath:         ptr.To("spec.domain.resources"),
							NodeAffinityPath:      ptr.To("spec.affinity.nodeAffinity"),
							PodAffinityPath:       ptr.To("spec.affinity.podAffinity"),
							PriorityClassNamePath: ptr.To("spec.priorityClassName"),
							SchedulerNamePath:     ptr.To("spec.schedulerName"),
						},
						MetadataPath: ptr.To("metadata"),
					},
				},
			},
		},
	}
}

func buildVMIStatusDefinition() *kartav1alpha1.StatusDefinition {
	return &kartav1alpha1.StatusDefinition{
		PhaseDefinition: &kartav1alpha1.PhaseDefinition{
			Path: ".status.phase",
		},
		ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
			Path:             ".status.conditions",
			TypeFieldName:    "type",
			StatusFieldName:  "status",
			MessageFieldName: ptr.To("message"),
		},
		StatusMappings: kartav1alpha1.StatusMappings{
			Initializing: []kartav1alpha1.StatusMatcher{
				{ByPhase: "Scheduled"},
			},
			Running: []kartav1alpha1.StatusMatcher{
				{ByPhase: "Running"},
			},
			Completed: []kartav1alpha1.StatusMatcher{
				{ByPhase: "Succeeded"},
			},
			Failed: []kartav1alpha1.StatusMatcher{
				{ByPhase: "Failed"},
			},
		},
	}
}
