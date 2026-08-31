/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nvidia

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/go_migrations/ri/source"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

func GetNVCFSPOTRequestRIConfiguration() source.KartaConfiguration {
	return source.KartaConfiguration{
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildNVCFSPOTRequestV1(),
		},
	}
}

func BuildNVCFSPOTRequestV1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "spotrequest",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "egx.nvidia.io",
						Version: "v1",
						Kind:    "SPOTRequest",
					},
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						MetadataPath: ptr.To(".metadata"),
					},
					StatusDefinition: &kartav1alpha1.StatusDefinition{
						StatusMappings: kartav1alpha1.StatusMappings{},
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To("1"),
					},
					PodSelector: &kartav1alpha1.PodSelector{
						ComponentTypeSelector: &kartav1alpha1.ComponentTypeSelector{
							KeyPath: ".metadata.labels[\"function-version-id\"]",
						},
					},
				},
			},
		},
	}
}
