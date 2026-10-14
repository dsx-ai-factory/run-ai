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

func GetKubeflowMPIJobRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Train",
		Priority:       "low",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildKubeflowMPIJobV1(),
			BuildKubeflowMPIJobV2beta1(),
		},
	}
}

func BuildKubeflowMPIJobV1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "mpijob",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "kubeflow.org",
						Version: "v1",
						Kind:    "MPIJob",
					},
					StatusDefinition:  buildKubeflowJobStatusDefinition(),
					SuspendDefinition: buildKubeflowSuspendDefinition(),
				},
				ChildComponents: buildKubeflowMPIJobChildComponents(".metadata.labels[\"training.kubeflow.org/replica-type\"]"),
			},
			Instructions: buildKubeflowMPIJobInstructions(),
		},
	}
}

func BuildKubeflowMPIJobV2beta1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "mpijob",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "kubeflow.org",
						Version: "v2beta1",
						Kind:    "MPIJob",
					},
					StatusDefinition:  buildKubeflowJobStatusDefinition(),
					SuspendDefinition: buildKubeflowSuspendDefinition(),
				},
				ChildComponents: buildKubeflowMPIJobChildComponents(".metadata.labels[\"training.kubeflow.org/job-role\"]"),
			},
			Instructions: buildKubeflowMPIJobInstructions(),
		},
	}
}

func buildKubeflowMPIJobChildComponents(roleLabelPath string) []kartav1alpha1.ComponentDefinition {
	return []kartav1alpha1.ComponentDefinition{
		{
			Name: "launcher",
			Kind: &kartav1alpha1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
			OwnerRef: ptr.To("mpijob"),
			SpecDefinition: &kartav1alpha1.SpecDefinition{
				PodTemplateSpecPath: ptr.To(".spec.mpiReplicaSpecs.Launcher.template"),
			},
			ScaleDefinition: &kartav1alpha1.ScaleDefinition{
				ReplicasPath: ptr.To(".spec.mpiReplicaSpecs.Launcher.replicas // 1"),
			},
			PodSelector: &kartav1alpha1.PodSelector{
				ComponentTypeSelector: &kartav1alpha1.ComponentTypeSelector{
					KeyPath: roleLabelPath,
					Value:   ptr.To("launcher"),
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
			OwnerRef: ptr.To("mpijob"),
			SpecDefinition: &kartav1alpha1.SpecDefinition{
				PodTemplateSpecPath: ptr.To(".spec.mpiReplicaSpecs.Worker.template"),
			},
			ScaleDefinition: &kartav1alpha1.ScaleDefinition{
				ReplicasPath: ptr.To(".spec.mpiReplicaSpecs.Worker.replicas // 1"),
			},
			PodSelector: &kartav1alpha1.PodSelector{
				ComponentTypeSelector: &kartav1alpha1.ComponentTypeSelector{
					KeyPath: roleLabelPath,
					Value:   ptr.To("worker"),
				},
			},
		},
	}
}

func buildKubeflowMPIJobInstructions() kartav1alpha1.OptimizationInstructions {
	return kartav1alpha1.OptimizationInstructions{
		GangScheduling: &kartav1alpha1.GangSchedulingInstruction{
			PodGroups: []kartav1alpha1.PodGroupDefinition{
				{
					Name: "job",
					Members: []kartav1alpha1.PodGroupMemberDefinition{
						{
							ComponentName: "launcher",
							GroupByKeyPaths: []string{
								".metadata.labels[\"training.kubeflow.org/job-name\"]",
							},
						},
						{
							ComponentName: "worker",
							GroupByKeyPaths: []string{
								".metadata.labels[\"training.kubeflow.org/job-name\"]",
							},
						},
					},
				},
			},
		},
	}
}
