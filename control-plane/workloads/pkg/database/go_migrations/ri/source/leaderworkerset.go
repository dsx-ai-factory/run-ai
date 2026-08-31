/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package source

import (
	karta "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

func GetLeaderWorkerSetRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildLeaderWorkerSetV1(),
		},
	}
}

func BuildLeaderWorkerSetV1() models.Karta {
	return models.Karta{
		Spec: karta.KartaSpec{
			StructureDefinition: karta.StructureDefinition{
				RootComponent: karta.ComponentDefinition{
					Name: "leaderworkerset",
					Kind: &karta.GroupVersionKind{
						Group:   "leaderworkerset.x-k8s.io",
						Version: "v1",
						Kind:    "LeaderWorkerSet",
					},
					StatusDefinition: &karta.StatusDefinition{
						ConditionsDefinition: &karta.ConditionsDefinition{
							Path:             ".status.conditions",
							TypeFieldName:    "type",
							StatusFieldName:  "status",
							MessageFieldName: ptr.To("message"),
						},
						StatusMappings: karta.StatusMappings{
							Initializing: []karta.StatusMatcher{
								{
									ByConditions: []karta.ExpectedCondition{
										{Type: "Progressing", Status: ptr.To("True")},
									},
								},
							},
							Running: []karta.StatusMatcher{
								{
									ByConditions: []karta.ExpectedCondition{
										{Type: "Available", Status: ptr.To("True")},
										{Type: "Progressing", Status: ptr.To("False")},
									},
								},
							},
							Failed: []karta.StatusMatcher{
								{
									ByConditions: []karta.ExpectedCondition{
										{Type: "Available", Status: ptr.To("False")},
										{Type: "Progressing", Status: ptr.To("False")},
										{Type: "UpdateInProgress", Status: ptr.To("False")},
									},
								},
							},
						},
					},
				},
				ChildComponents: []karta.ComponentDefinition{
					{
						Name:     "group",
						OwnerRef: ptr.To("leaderworkerset"),
						ScaleDefinition: &karta.ScaleDefinition{
							ReplicasPath: ptr.To(".spec.leaderWorkerTemplate.size"),
						},
						PodSelector: &karta.PodSelector{
							ReplicaSelector: &karta.ReplicaSelector{
								KeyPath: ".metadata.labels[\"leaderworkerset.sigs.k8s.io/group-index\"]",
							},
						},
					},
					{
						Name:     "leader",
						OwnerRef: ptr.To("group"),
						SpecDefinition: &karta.SpecDefinition{
							PodTemplateSpecPath: ptr.To(".spec.leaderWorkerTemplate.leaderTemplate"),
						},
						ScaleDefinition: &karta.ScaleDefinition{
							ReplicasPath: ptr.To(".spec.replicas // 1"),
						},
						PodSelector: &karta.PodSelector{
							ComponentTypeSelector: &karta.ComponentTypeSelector{
								KeyPath: ".metadata.labels[\"leaderworkerset.sigs.k8s.io/worker-index\"]",
								Value:   ptr.To("0"),
							},
						},
					},
					{
						Name: "worker",
						Kind: &karta.GroupVersionKind{
							Group:   "apps",
							Version: "v1",
							Kind:    "StatefulSet",
						},
						OwnerRef: ptr.To("group"),
						SpecDefinition: &karta.SpecDefinition{
							PodTemplateSpecPath: ptr.To(".spec.leaderWorkerTemplate.workerTemplate"),
						},
						ScaleDefinition: &karta.ScaleDefinition{
							ReplicasPath: ptr.To("(.spec.replicas // 1) * ((.spec.leaderWorkerTemplate.size // 1) - 1)"),
						},
						PodSelector: &karta.PodSelector{
							ComponentTypeSelector: &karta.ComponentTypeSelector{
								KeyPath: ".metadata.annotations[\"leaderworkerset.sigs.k8s.io/leader-name\"]",
							},
						},
					},
				},
			},
			Instructions: karta.OptimizationInstructions{
				GangScheduling: &karta.GangSchedulingInstruction{
					PodGroups: []karta.PodGroupDefinition{
						{
							Name: "group",
							// Gang scheduling for leader-worker coordination
							// Multi-group pattern: podgroup-lws-{name}-group-{index}
							// Uses default group-index "0" for single-replica LWS
							Members: []karta.PodGroupMemberDefinition{
								{
									ComponentName: "group",
									GroupByKeyPaths: []string{
										".metadata.labels[\"leaderworkerset.sigs.k8s.io/name\"]",
										".metadata.labels[\"leaderworkerset.sigs.k8s.io/group-index\"] // \"0\"",
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
