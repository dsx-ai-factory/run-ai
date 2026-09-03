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

func GetKubeflowPyTorchJobRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Train",
		Priority:       "low",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildKubeflowPyTorchJobV1(),
		},
	}
}

func BuildKubeflowPyTorchJobV1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "pytorchjob",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "kubeflow.org",
						Version: "v1",
						Kind:    "PyTorchJob",
					},
					StatusDefinition:  buildKubeflowJobStatusDefinition(),
					SuspendDefinition: buildKubeflowSuspendDefinition(),
				},
				ChildComponents: buildKubeflowPyTorchJobChildComponents(),
			},
			Instructions: buildKubeflowPyTorchJobInstructions(),
		},
	}
}

func buildKubeflowPyTorchJobChildComponents() []kartav1alpha1.ComponentDefinition {
	return []kartav1alpha1.ComponentDefinition{
		{
			Name: "master",
			Kind: &kartav1alpha1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
			OwnerRef: ptr.To("pytorchjob"),
			SpecDefinition: &kartav1alpha1.SpecDefinition{
				PodTemplateSpecPath: ptr.To(".spec.pytorchReplicaSpecs.Master.template"),
			},
			ScaleDefinition: &kartav1alpha1.ScaleDefinition{
				ReplicasPath: ptr.To(".spec.pytorchReplicaSpecs.Master.replicas // 1"),
			},
			PodSelector: &kartav1alpha1.PodSelector{
				ComponentTypeSelector: &kartav1alpha1.ComponentTypeSelector{
					KeyPath: ".metadata.labels[\"training.kubeflow.org/replica-type\"]",
					Value:   ptr.To("master"),
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
			OwnerRef: ptr.To("pytorchjob"),
			SpecDefinition: &kartav1alpha1.SpecDefinition{
				PodTemplateSpecPath: ptr.To(".spec.pytorchReplicaSpecs.Worker.template"),
			},
			ScaleDefinition: &kartav1alpha1.ScaleDefinition{
				ReplicasPath:    ptr.To(".spec.pytorchReplicaSpecs.Worker.replicas // 1"),
				MinReplicasPath: ptr.To(".spec.elasticPolicy.minReplicas"),
				MaxReplicasPath: ptr.To(".spec.elasticPolicy.maxReplicas"),
			},
			PodSelector: &kartav1alpha1.PodSelector{
				ComponentTypeSelector: &kartav1alpha1.ComponentTypeSelector{
					KeyPath: ".metadata.labels[\"training.kubeflow.org/replica-type\"]",
					Value:   ptr.To("worker"),
				},
			},
		},
	}
}

func buildKubeflowPyTorchJobInstructions() kartav1alpha1.OptimizationInstructions {
	return kartav1alpha1.OptimizationInstructions{
		GangScheduling: &kartav1alpha1.GangSchedulingInstruction{
			PodGroups: []kartav1alpha1.PodGroupDefinition{
				{
					Name: "job",
					Members: []kartav1alpha1.PodGroupMemberDefinition{
						{
							ComponentName: "master",
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
