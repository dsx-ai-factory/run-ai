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

func GetDevWorkspaceRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Build",
		Priority:       "high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildDevWorkspaceV1alpha2(),
		},
	}
}

func BuildDevWorkspaceV1alpha2() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "devworkspace",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "workspace.devfile.io",
						Version: "v1alpha2",
						Kind:    "DevWorkspace",
					},
					StatusDefinition: buildDevWorkspaceStatusDefinition(),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: ptr.To(".spec.template"),
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To("1"),
					},
				},
			},
		},
	}
}

func buildDevWorkspaceStatusDefinition() *kartav1alpha1.StatusDefinition {
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
				{ByPhase: "Starting"},
			},
			Running: []kartav1alpha1.StatusMatcher{
				{ByPhase: "Running"},
			},
			Failed: []kartav1alpha1.StatusMatcher{
				{ByPhase: "Failed"},
			},
		},
	}
}
