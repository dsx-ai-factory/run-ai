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

func GetRayClusterRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Train",
		Priority:       "low",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildRayClusterV1alpha1(),
			BuildRayClusterV1(),
		},
	}
}

func BuildRayClusterV1alpha1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "raycluster",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "ray.io",
						Version: "v1alpha1",
						Kind:    "RayCluster",
					},
					StatusDefinition: buildRayClusterStatusDefinition(),
				},
				ChildComponents: buildRayClusterChildComponents(),
			},
			Instructions: buildRayClusterInstructions(),
		},
	}
}

func BuildRayClusterV1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "raycluster",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "ray.io",
						Version: "v1",
						Kind:    "RayCluster",
					},
					StatusDefinition: buildRayClusterStatusDefinition(),
				},
				ChildComponents: buildRayClusterChildComponents(),
			},
			Instructions: buildRayClusterInstructions(),
		},
	}
}

func buildRayClusterChildComponents() []kartav1alpha1.ComponentDefinition {
	return []kartav1alpha1.ComponentDefinition{
		{
			Name: "head",
			Kind: &kartav1alpha1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
			OwnerRef: ptr.To("raycluster"),
			SpecDefinition: &kartav1alpha1.SpecDefinition{
				PodTemplateSpecPath: ptr.To(".spec.headGroupSpec.template"),
			},
			ScaleDefinition: &kartav1alpha1.ScaleDefinition{
				ReplicasPath: ptr.To("1"),
			},
			PodSelector: &kartav1alpha1.PodSelector{
				ComponentTypeSelector: &kartav1alpha1.ComponentTypeSelector{
					KeyPath: ".metadata.labels[\"ray.io/node-type\"]",
					Value:   ptr.To("head"),
				},
			},
		},
		{
			Name: "worker",
			Kind: &kartav1alpha1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
			OwnerRef: ptr.To("raycluster"),
			SpecDefinition: &kartav1alpha1.SpecDefinition{
				PodTemplateSpecPath: ptr.To(".spec.workerGroupSpecs[].template"),
			},
			ScaleDefinition: &kartav1alpha1.ScaleDefinition{
				ReplicasPath:    ptr.To(".spec.workerGroupSpecs[].replicas"),
				MinReplicasPath: ptr.To(".spec.workerGroupSpecs[].minReplicas"),
				MaxReplicasPath: ptr.To(".spec.workerGroupSpecs[].maxReplicas"),
			},
			PodSelector: &kartav1alpha1.PodSelector{
				ComponentTypeSelector: &kartav1alpha1.ComponentTypeSelector{
					KeyPath: ".metadata.labels[\"ray.io/node-type\"]",
					Value:   ptr.To("worker"),
				},
				ComponentInstanceSelector: &kartav1alpha1.ComponentInstanceSelector{
					IdPath: ".metadata.labels[\"ray.io/group\"]",
				},
			},
			InstanceIdPath: ptr.To(".spec.workerGroupSpecs[].groupName"),
		},
	}
}

func buildRayClusterInstructions() kartav1alpha1.OptimizationInstructions {
	return kartav1alpha1.OptimizationInstructions{
		GangScheduling: &kartav1alpha1.GangSchedulingInstruction{
			PodGroups: []kartav1alpha1.PodGroupDefinition{
				{
					Name: "cluster",
					Members: []kartav1alpha1.PodGroupMemberDefinition{
						{
							ComponentName: "head",
							GroupByKeyPaths: []string{
								".metadata.labels[\"ray.io/cluster\"]",
							},
						},
						{
							ComponentName: "worker",
							GroupByKeyPaths: []string{
								".metadata.labels[\"ray.io/cluster\"]",
							},
						},
					},
				},
			},
		},
	}
}
