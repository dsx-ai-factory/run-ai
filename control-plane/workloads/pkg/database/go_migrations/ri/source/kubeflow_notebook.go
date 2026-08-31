/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package source

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

func GetKubeflowNotebookRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Build",
		Priority:       "high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildKubeflowNotebookV1beta1(),
		},
	}
}

func BuildKubeflowNotebookV1beta1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "notebook",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "kubeflow.org",
						Version: "v1beta1",
						Kind:    "Notebook",
					},
					StatusDefinition: buildKubeflowJobStatusDefinition(),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: ptr.To(".spec.template"),
					},
				},
			},
		},
	}
}
