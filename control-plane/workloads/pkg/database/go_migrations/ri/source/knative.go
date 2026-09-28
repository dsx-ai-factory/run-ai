/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package source

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/utils/ptr"
)

func GetKNativeServiceRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildKNativeServiceV1(),
		},
	}
}

func BuildKNativeServiceV1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "knativeservice",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "serving.knative.dev",
						Version: "v1",
						Kind:    "Service",
					},
					StatusDefinition: buildKNativeStatusDefinition(),
				},
				AdditionalChildKinds: []kartav1alpha1.GroupVersionKind{
					{
						Group:   "apps",
						Kind:    "Deployment",
						Version: "v1",
					},
				},
				ChildComponents: []kartav1alpha1.ComponentDefinition{
					{
						Name: "revision",
						Kind: &kartav1alpha1.GroupVersionKind{
							Group:   "serving.knative.dev",
							Version: "v1",
							Kind:    "Revision",
						},
						OwnerRef: ptr.To("knativeservice"),
						SpecDefinition: &kartav1alpha1.SpecDefinition{
							PodTemplateSpecPath: ptr.To(".spec.template"),
						},
						ScaleDefinition: &kartav1alpha1.ScaleDefinition{
							MinReplicasPath: ptr.To(".spec.template.metadata.annotations[\"autoscaling.knative.dev/min-scale\"] // 1 | tonumber"),
							MaxReplicasPath: ptr.To(".spec.template.metadata.annotations[\"autoscaling.knative.dev/max-scale\"] | tonumber?"),
						},
					},
				}},
			Instructions: kartav1alpha1.OptimizationInstructions{
				GangScheduling: &kartav1alpha1.GangSchedulingInstruction{
					PodGroups: []kartav1alpha1.PodGroupDefinition{
						{
							Name: "revision",
							Members: []kartav1alpha1.PodGroupMemberDefinition{
								{
									ComponentName: "revision",
									GroupByKeyPaths: []string{
										".metadata.labels[\"serving.knative.dev/revision\"]",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func buildKNativeStatusDefinition() *kartav1alpha1.StatusDefinition {
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
						{Type: "Ready", Status: ptr.To("True")},
						{Type: "RoutesReady", Status: ptr.To("True")},
						{Type: "ConfigurationsReady", Status: ptr.To("True")},
					},
				},
			},
		},
	}
}
