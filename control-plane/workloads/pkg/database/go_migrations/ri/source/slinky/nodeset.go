/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package slinky

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/go_migrations/ri/source"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

func GetSlinkyNodeSetRIConfiguration() source.KartaConfiguration {
	return source.KartaConfiguration{
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildSlinkyNodeSetV1beta1(),
		},
	}
}

func BuildSlinkyNodeSetV1beta1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "nodeset",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "slinky.slurm.net",
						Version: "v1beta1",
						Kind:    "NodeSet",
					},
					PodSelector: &kartav1alpha1.PodSelector{
						ComponentTypeSelector: &kartav1alpha1.ComponentTypeSelector{
							KeyPath: ".metadata.labels[\"app.kubernetes.io/name\"]",
							Value:   ptr.To("slurmd"),
						},
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To(".spec.replicas"),
					},
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
							AnnotationsPath:       ptr.To(".spec.template.metadata.annotations"),
							ContainerPath:         ptr.To("(.spec.slurmd // .spec.template.spec.containers[0])"),
							ImagePath:             ptr.To("(.spec.slurmd.image // .spec.template.spec.containers[0].image)"),
							LabelsPath:            ptr.To(".spec.template.metadata.labels"),
							NodeAffinityPath:      ptr.To(".spec.template.spec.affinity.nodeAffinity"),
							PodAffinityPath:       ptr.To(".spec.template.spec.affinity.podAffinity"),
							PriorityClassNamePath: ptr.To(".spec.template.spec.priorityClassName"),
							ResourcesPath:         ptr.To("(.spec.slurmd.resources // .spec.template.spec.containers[0].resources)"),
							SchedulerNamePath:     ptr.To(".spec.template.spec.schedulerName"),
						},
					},
					StatusDefinition: buildNodeSetStatusDefinition(),
				},
			},
		},
	}
}

func buildNodeSetStatusDefinition() *kartav1alpha1.StatusDefinition {
	return &kartav1alpha1.StatusDefinition{
		StatusMappings: kartav1alpha1.StatusMappings{
			Initializing: []kartav1alpha1.StatusMatcher{
				{
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     "(.status.replicas // 0) > 0 and (.status.readyReplicas // 0) == 0",
						ExpectedResult: "true",
					},
				},
			},
			Running: []kartav1alpha1.StatusMatcher{
				{
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     "(.status.readyReplicas // 0) > 0",
						ExpectedResult: "true",
					},
				},
			},
		},
	}
}
