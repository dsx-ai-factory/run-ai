/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package source

import (
	"strings"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/utils/ptr"
)

func GetKubeflowJAXJobRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Train",
		Priority:       "low",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildKubeflowJAXJobV1(),
		},
	}
}

func BuildKubeflowJAXJobV1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "jaxjob",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "kubeflow.org",
						Version: "v1",
						Kind:    "JAXJob",
					},
					StatusDefinition:  buildKubeflowJobStatusDefinition(),
					SuspendDefinition: buildKubeflowSuspendDefinition(),
				},
				ChildComponents: buildKubeflowJAXJobChildComponents(),
			},
			Instructions: buildKubeflowJAXJobInstructions(),
		},
	}
}

func buildKubeflowJAXJobChildComponents() []kartav1alpha1.ComponentDefinition {
	replicaTypes := []string{"Worker"}
	components := make([]kartav1alpha1.ComponentDefinition, 0, len(replicaTypes))

	for _, replicaType := range replicaTypes {
		components = append(components, kartav1alpha1.ComponentDefinition{
			Name: replicaType,
			Kind: &kartav1alpha1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
			OwnerRef: ptr.To("jaxjob"),
			SpecDefinition: &kartav1alpha1.SpecDefinition{
				PodTemplateSpecPath: ptr.To(".spec.jaxReplicaSpecs." + replicaType + ".template"),
			},
			ScaleDefinition: &kartav1alpha1.ScaleDefinition{
				ReplicasPath: ptr.To(".spec.jaxReplicaSpecs." + replicaType + ".replicas // 1"),
			},
			PodSelector: &kartav1alpha1.PodSelector{
				ComponentTypeSelector: &kartav1alpha1.ComponentTypeSelector{
					KeyPath: ".metadata.labels[\"training.kubeflow.org/replica-type\"]",
					Value:   ptr.To(strings.ToLower(replicaType)),
				},
			},
		})
	}

	return components
}

func buildKubeflowJAXJobInstructions() kartav1alpha1.OptimizationInstructions {
	return kartav1alpha1.OptimizationInstructions{
		GangScheduling: &kartav1alpha1.GangSchedulingInstruction{
			PodGroups: []kartav1alpha1.PodGroupDefinition{
				{
					Name: "job",
					Members: []kartav1alpha1.PodGroupMemberDefinition{
						{
							ComponentName: "Worker",
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
