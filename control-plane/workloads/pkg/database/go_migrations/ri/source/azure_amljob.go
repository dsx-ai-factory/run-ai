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

func GetAzureAmlJobRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Train",
		Priority:       "low",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildAzureAmlJobV1alpha1(),
		},
	}
}

func BuildAzureAmlJobV1alpha1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "amljob",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "amlarc.azureml.com",
						Version: "v1alpha1",
						Kind:    "AmlJob",
					},
					StatusDefinition: buildAmlJobStatusDefinition(),
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To("1"),
					},
				},
			},
		},
	}
}

func buildAmlJobStatusDefinition() *kartav1alpha1.StatusDefinition {
	return &kartav1alpha1.StatusDefinition{
		ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
			Path:             ".status.conditions",
			TypeFieldName:    "type",
			StatusFieldName:  "status",
			MessageFieldName: ptr.To("message"),
		},
		StatusMappings: kartav1alpha1.StatusMappings{},
	}
}
