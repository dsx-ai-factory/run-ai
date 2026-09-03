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

func GetKubeflowScheduledWorkflowRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildKubeflowScheduledWorkflowV1alpha1(),
		},
	}
}

func BuildKubeflowScheduledWorkflowV1alpha1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "scheduledworkflow",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "kubeflow.org",
						Version: "v1alpha1",
						Kind:    "ScheduledWorkflow",
					},
					StatusDefinition: buildScheduledWorkflowStatusDefinition(),
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To("1"),
					},
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: ptr.To(".spec.workflow"),
					},
				},
			},
		},
	}
}

func buildScheduledWorkflowStatusDefinition() *kartav1alpha1.StatusDefinition {
	return &kartav1alpha1.StatusDefinition{
		ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
			Path:             ".status.conditions",
			TypeFieldName:    "type",
			StatusFieldName:  "status",
			MessageFieldName: ptr.To("message"),
		},
		StatusMappings: kartav1alpha1.StatusMappings{
			Running: []kartav1alpha1.StatusMatcher{
				{
					ByConditions: []kartav1alpha1.ExpectedCondition{
						{Type: "Enabled", Status: ptr.To("True")},
					},
				},
			},
		},
	}
}
