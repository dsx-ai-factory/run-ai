/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package source

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

func GetTektonTaskRunRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Build",
		Priority:       "high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildTektonTaskRunV1(),
		},
	}
}

func BuildTektonTaskRunV1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "taskrun",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "tekton.dev",
						Version: "v1",
						Kind:    "TaskRun",
					},
					StatusDefinition: buildTektonTaskRunStatusDefinition(),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: ptr.To(".spec.podTemplate"),
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To("1"),
					},
				},
			},
		},
	}
}

func buildTektonTaskRunStatusDefinition() *kartav1alpha1.StatusDefinition {
	return buildTektonStatusDefinition()
}
