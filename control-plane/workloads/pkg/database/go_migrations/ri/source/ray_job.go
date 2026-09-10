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

func GetRayJobRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Train",
		Priority:       "low",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildRayJobV1alpha1(),
			BuildRayJobV1(),
		},
	}
}

func BuildRayJobV1alpha1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "rayjob",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "ray.io",
						Version: "v1alpha1",
						Kind:    "RayJob",
					},
					StatusDefinition:  buildRayJobStatusDefinition(),
					SuspendDefinition: buildRayJobSuspendDefinition(),
				},
				ChildComponents: buildRayJobChildComponents(),
			},
			Instructions: buildRayJobInstructions(),
		},
	}
}

func BuildRayJobV1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "rayjob",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "ray.io",
						Version: "v1",
						Kind:    "RayJob",
					},
					StatusDefinition:  buildRayJobStatusDefinition(),
					SuspendDefinition: buildRayJobSuspendDefinition(),
				},
				ChildComponents: buildRayJobChildComponents(),
			},
			Instructions: buildRayJobInstructions(),
		},
	}
}

func buildRayJobChildComponents() []kartav1alpha1.ComponentDefinition {
	return []kartav1alpha1.ComponentDefinition{
		{
			Name: "head",
			Kind: &kartav1alpha1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
			OwnerRef: ptr.To("rayjob"),
			SpecDefinition: &kartav1alpha1.SpecDefinition{
				PodTemplateSpecPath: ptr.To(".spec.rayClusterSpec.headGroupSpec.template"),
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
			OwnerRef: ptr.To("rayjob"),
			SpecDefinition: &kartav1alpha1.SpecDefinition{
				PodTemplateSpecPath: ptr.To(".spec.rayClusterSpec.workerGroupSpecs[]?.template"),
			},
			ScaleDefinition: &kartav1alpha1.ScaleDefinition{
				ReplicasPath: ptr.To(".spec.rayClusterSpec.workerGroupSpecs[]? | (.replicas // 1)"),
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
			InstanceIdPath: ptr.To(".spec.rayClusterSpec.workerGroupSpecs[]?.groupName"),
		},
	}
}

func buildRayJobInstructions() kartav1alpha1.OptimizationInstructions {
	return kartav1alpha1.OptimizationInstructions{
		GangScheduling: &kartav1alpha1.GangSchedulingInstruction{
			PodGroups: []kartav1alpha1.PodGroupDefinition{
				{
					Name: "job",
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
