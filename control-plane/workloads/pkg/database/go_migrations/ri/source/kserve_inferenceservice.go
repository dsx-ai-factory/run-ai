/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package source

import (
	karta "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

func GetKserveInferenceServiceRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildKServeInferenceServiceV1beta1(),
		},
	}
}

func BuildKServeInferenceServiceV1beta1() models.Karta {
	return models.Karta{
		Spec: karta.KartaSpec{
			StructureDefinition: karta.StructureDefinition{
				RootComponent: karta.ComponentDefinition{
					Name: "inferenceservice",
					Kind: &karta.GroupVersionKind{
						Group:   "serving.kserve.io",
						Version: "v1beta1",
						Kind:    "InferenceService",
					},
					StatusDefinition: &karta.StatusDefinition{
						ConditionsDefinition: &karta.ConditionsDefinition{
							Path:             ".status.conditions",
							TypeFieldName:    "type",
							StatusFieldName:  "status",
							MessageFieldName: ptr.To("message"),
						},
						StatusMappings: karta.StatusMappings{
							Running: []karta.StatusMatcher{
								{
									ByConditions: []karta.ExpectedCondition{
										{Type: "PredictorReady", Status: ptr.To("True")},
										{Type: "RoutesReady", Status: ptr.To("True")},
										{Type: "LatestDeploymentReady", Status: ptr.To("True")},
									},
								},
							},
							Failed: []karta.StatusMatcher{
								{
									ByConditions: []karta.ExpectedCondition{
										{Type: "PredictorReady", Status: ptr.To("False")},
										{Type: "PredictorConfigurationReady", Status: ptr.To("False")},
										{Type: "RoutesReady", Status: ptr.To("False")},
									},
								},
							},
						},
					},
				},
				ChildComponents: []karta.ComponentDefinition{
					{
						Name: "predictor",
						Kind: &karta.GroupVersionKind{
							Group:   "apps",
							Version: "v1",
							Kind:    "Deployment",
						},
						OwnerRef: ptr.To("inferenceservice"),
						SpecDefinition: &karta.SpecDefinition{
							FragmentedPodSpecDefinition: &karta.FragmentedPodSpecDefinition{
								NodeAffinityPath:      ptr.To(".spec.predictor.affinity.nodeAffinity"),
								PodAffinityPath:       ptr.To(".spec.predictor.affinity.podAffinity"),
								PriorityClassNamePath: ptr.To(".spec.predictor.priorityClassName"),
								SchedulerNamePath:     ptr.To(".spec.predictor.schedulerName"),
								LabelsPath:            ptr.To(".spec.predictor.labels"),
								AnnotationsPath:       ptr.To(".spec.predictor.annotations"),
								ContainerPath:         ptr.To(".spec.predictor | ( (.[]?  | select(type ==\"object\" and .storageUri )))"),
								ContainersPath:        ptr.To(".spec.predictor.containers"),
							},
						},
						ScaleDefinition: &karta.ScaleDefinition{
							MinReplicasPath: ptr.To(".spec.predictor.minReplicas"),
							MaxReplicasPath: ptr.To(".spec.predictor.maxReplicas"),
						},
						PodSelector: &karta.PodSelector{
							ComponentTypeSelector: &karta.ComponentTypeSelector{
								KeyPath: ".metadata.labels[\"component\"]",
								Value:   ptr.To("predictor"),
							},
						},
					},
					{
						Name: "transformer",
						Kind: &karta.GroupVersionKind{
							Group:   "apps",
							Version: "v1",
							Kind:    "Deployment",
						},
						OwnerRef: ptr.To("inferenceservice"),
						SpecDefinition: &karta.SpecDefinition{
							PodSpecPath:  ptr.To(".spec.transformer"),
							MetadataPath: ptr.To(".spec.transformer"),
						},
						ScaleDefinition: &karta.ScaleDefinition{
							MinReplicasPath: ptr.To(".spec.transformer.minReplicas"),
							MaxReplicasPath: ptr.To(".spec.transformer.maxReplicas"),
						},
						PodSelector: &karta.PodSelector{
							ComponentTypeSelector: &karta.ComponentTypeSelector{
								KeyPath: ".metadata.labels[\"component\"]",
								Value:   ptr.To("transformer"),
							},
						},
					},
				},
			},
			Instructions: karta.OptimizationInstructions{
				GangScheduling: &karta.GangSchedulingInstruction{
					PodGroups: []karta.PodGroupDefinition{
						{
							Name: "service",
							Members: []karta.PodGroupMemberDefinition{
								{
									ComponentName: "predictor",
									GroupByKeyPaths: []string{
										".metadata.labels[\"serving.kserve.io/inferenceservice\"]",
									},
								},
								{
									ComponentName: "transformer",
									GroupByKeyPaths: []string{
										".metadata.labels[\"serving.kserve.io/inferenceservice\"]",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
