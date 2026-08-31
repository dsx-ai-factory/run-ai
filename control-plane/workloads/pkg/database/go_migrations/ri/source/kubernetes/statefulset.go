/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kubernetes

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/go_migrations/ri/source"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

func GetK8sStatefulSetRIConfiguration() source.KartaConfiguration {
	return source.KartaConfiguration{
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildK8sStatefulSetV1(),
		},
	}
}

func BuildK8sStatefulSetV1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "statefulset",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "apps",
						Version: "v1",
						Kind:    "StatefulSet",
					},
					StatusDefinition: buildStatefulSetStatusDefinition(),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: ptr.To(".spec.template"),
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To(".spec.replicas // 1"),
					},
				},
			},
		},
	}
}

func buildStatefulSetStatusDefinition() *kartav1alpha1.StatusDefinition {
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
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     ".status.readyReplicas == (.spec.replicas // 1)",
						ExpectedResult: "true",
					},
				},
			},
			Initializing: []kartav1alpha1.StatusMatcher{
				{
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     "(.status.readyReplicas // 0) < (.spec.replicas // 1)",
						ExpectedResult: "true",
					},
				},
			},
		},
	}
}
