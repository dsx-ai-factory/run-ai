/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package source

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	karta "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/utils/ptr"
)

func GetDynamoGraphDeploymentRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildDynamoGraphDeploymentV1alpha1(),
			BuildDynamoGraphDeploymentV1beta1(),
		},
	}
}

func BuildDynamoGraphDeploymentV1alpha1() models.Karta {
	return models.Karta{
		Spec: karta.KartaSpec{
			StructureDefinition: karta.StructureDefinition{
				RootComponent: karta.ComponentDefinition{
					Name: "dynamographdeployment",
					Kind: &karta.GroupVersionKind{
						Group:   "nvidia.com",
						Version: "v1alpha1",
						Kind:    "DynamoGraphDeployment",
					},
					StatusDefinition: buildDynamoGraphDeploymentStatusDefinition(),
				},
				ChildComponents: buildDynamoGraphDeploymentV1alpha1ChildComponents(),
				AdditionalChildKinds: []karta.GroupVersionKind{
					{
						Group:   "nvidia.com",
						Kind:    "DynamoComponentDeployment",
						Version: "v1alpha1",
					},
					{
						Group:   "leaderworkerset.x-k8s.io",
						Kind:    "LeaderWorkerSet",
						Version: "v1",
					},
					{
						Group:   "scheduler.grove.io",
						Kind:    "PodGang",
						Version: "v1alpha1",
					},
					{
						Group:   "grove.io",
						Kind:    "PodClique",
						Version: "v1alpha1",
					},
					{
						Group:   "grove.io",
						Kind:    "PodCliqueSet",
						Version: "v1alpha1",
					},
					{
						Group:   "grove.io",
						Kind:    "PodCliqueScalingGroup",
						Version: "v1alpha1",
					},
				},
			},
			Instructions: buildDynamoGraphDeploymentV1alpha1Instructions(),
		},
	}
}

func BuildDynamoGraphDeploymentV1beta1() models.Karta {
	return models.Karta{
		Spec: karta.KartaSpec{
			StructureDefinition: karta.StructureDefinition{
				RootComponent: karta.ComponentDefinition{
					Name: "dynamographdeployment",
					Kind: &karta.GroupVersionKind{
						Group:   "nvidia.com",
						Version: "v1beta1",
						Kind:    "DynamoGraphDeployment",
					},
					StatusDefinition: buildDynamoGraphDeploymentStatusDefinition(),
				},
				ChildComponents: buildDynamoGraphDeploymentV1beta1ChildComponents(),
				AdditionalChildKinds: []karta.GroupVersionKind{
					{
						Group:   "nvidia.com",
						Kind:    "DynamoComponentDeployment",
						Version: "v1beta1",
					},
					{
						Group:   "leaderworkerset.x-k8s.io",
						Kind:    "LeaderWorkerSet",
						Version: "v1",
					},
					{
						Group:   "scheduler.grove.io",
						Kind:    "PodGang",
						Version: "v1alpha1",
					},
					{
						Group:   "grove.io",
						Kind:    "PodClique",
						Version: "v1alpha1",
					},
					{
						Group:   "grove.io",
						Kind:    "PodCliqueSet",
						Version: "v1alpha1",
					},
					{
						Group:   "grove.io",
						Kind:    "PodCliqueScalingGroup",
						Version: "v1alpha1",
					},
				},
			},
			Instructions: buildDynamoGraphDeploymentV1beta1Instructions(),
		},
	}
}

func buildDynamoGraphDeploymentV1alpha1ChildComponents() []karta.ComponentDefinition {
	return []karta.ComponentDefinition{
		{
			Name:     "service",
			OwnerRef: ptr.To("dynamographdeployment"),
			SpecDefinition: &karta.SpecDefinition{
				FragmentedPodSpecDefinition: &karta.FragmentedPodSpecDefinition{
					LabelsPath:            ptr.To(".spec.services | .[] | .labels"),
					AnnotationsPath:       ptr.To(".spec.services | .[] | .annotations"),
					ResourcesPath:         ptr.To(".spec.services | .[] | .resources"),
					SchedulerNamePath:     ptr.To(".spec.services | .[] | .extraPodSpec.schedulerName"),
					ResourceClaimsPath:    ptr.To(".spec.services | .[] | .extraPodSpec.resourceClaims"),
					NodeAffinityPath:      ptr.To(".spec.services | .[] | .extraPodSpec.affinity.nodeAffinity"),
					PodAffinityPath:       ptr.To(".spec.services | .[] | .extraPodSpec.affinity.podAffinity"),
					PriorityClassNamePath: ptr.To(".spec.services | .[] | .extraPodSpec.priorityClassName"),
					ContainerPath:         ptr.To(".spec.services | .[] | .extraPodSpec.mainContainer"),
					ImagePath:             ptr.To(".spec.services | .[] | .extraPodSpec.mainContainer.image"),
				},
			},
			ScaleDefinition: &karta.ScaleDefinition{
				ReplicasPath:    ptr.To(".spec.services[] | (.replicas // 1) * (.multinode.nodeCount // 1)"),
				MinReplicasPath: ptr.To(".spec.services | .[] | .autoscaling.minReplicas"),
				MaxReplicasPath: ptr.To(".spec.services | .[] | .autoscaling.maxReplicas"),
			},
			PodSelector: &karta.PodSelector{
				ComponentInstanceSelector: &karta.ComponentInstanceSelector{
					IdPath: ".metadata.labels[\"nvidia.com/dynamo-component\"]",
				},
				ReplicaSelector: &karta.ReplicaSelector{
					KeyPath: ".metadata.labels[\"grove.io/podcliquescalinggroup-replica-index\"] // .metadata.labels[\"leaderworkerset.sigs.k8s.io/group-index\"]",
				},
			},
			InstanceIdPath: ptr.To(".spec.services | to_entries[] | .key"),
		},
	}
}

func buildDynamoGraphDeploymentV1beta1ChildComponents() []karta.ComponentDefinition {
	return []karta.ComponentDefinition{
		{
			Name:     "component",
			OwnerRef: ptr.To("dynamographdeployment"),
			SpecDefinition: &karta.SpecDefinition{
				FragmentedPodSpecDefinition: &karta.FragmentedPodSpecDefinition{
					LabelsPath:            ptr.To(".spec.components[] | .podTemplate.metadata.labels"),
					AnnotationsPath:       ptr.To(".spec.components[] | .podTemplate.metadata.annotations"),
					SchedulerNamePath:     ptr.To(".spec.components[] | .podTemplate.spec.schedulerName"),
					ResourceClaimsPath:    ptr.To(".spec.components[] | .podTemplate.spec.resourceClaims"),
					NodeAffinityPath:      ptr.To(".spec.components[] | .podTemplate.spec.affinity.nodeAffinity"),
					PodAffinityPath:       ptr.To(".spec.components[] | .podTemplate.spec.affinity.podAffinity"),
					PriorityClassNamePath: ptr.To(".spec.components[] | .podTemplate.spec.priorityClassName"),
					ContainerPath:         ptr.To(".spec.components[] | .podTemplate.spec.containers[] | select(.name == \"main\")"),
					ImagePath:             ptr.To(".spec.components[] | .podTemplate.spec.containers[] | select(.name == \"main\") | .image"),
				},
			},
			ScaleDefinition: &karta.ScaleDefinition{
				ReplicasPath: ptr.To(".spec.components[] | (.replicas // 1) * (.multinode.nodeCount // 1)"),
			},
			PodSelector: &karta.PodSelector{
				ComponentInstanceSelector: &karta.ComponentInstanceSelector{
					IdPath: ".metadata.labels[\"nvidia.com/dynamo-component\"]",
				},
				ReplicaSelector: &karta.ReplicaSelector{
					KeyPath: ".metadata.labels[\"grove.io/podcliquescalinggroup-replica-index\"] // .metadata.labels[\"leaderworkerset.sigs.k8s.io/group-index\"]",
				},
			},
			InstanceIdPath: ptr.To(".spec.components[] | .name"),
		},
	}
}

func buildDynamoGraphDeploymentV1beta1Instructions() karta.OptimizationInstructions {
	return karta.OptimizationInstructions{
		GangScheduling: &karta.GangSchedulingInstruction{
			PodGroups: []karta.PodGroupDefinition{
				{
					Name: "component",
					Members: []karta.PodGroupMemberDefinition{
						{
							ComponentName: "component",
							GroupByKeyPaths: []string{
								".metadata.labels[\"nvidia.com/dynamo-component\"] | ascii_downcase",
							},
						},
					},
				},
			},
		},
	}
}

func buildDynamoGraphDeploymentStatusDefinition() *karta.StatusDefinition {
	return &karta.StatusDefinition{
		PhaseDefinition: &karta.PhaseDefinition{
			Path: ".status.state",
		},
		StatusMappings: karta.StatusMappings{
			Initializing: []karta.StatusMatcher{
				{ByPhase: "pending"},
			},
			Running: []karta.StatusMatcher{
				{ByPhase: "successful"},
			},
			Failed: []karta.StatusMatcher{
				{ByPhase: "failed"},
			},
		},
	}
}

func buildDynamoGraphDeploymentV1alpha1Instructions() karta.OptimizationInstructions {
	return karta.OptimizationInstructions{
		GangScheduling: &karta.GangSchedulingInstruction{
			PodGroups: []karta.PodGroupDefinition{
				{
					Name: "service",
					Members: []karta.PodGroupMemberDefinition{
						{
							ComponentName: "service",
							GroupByKeyPaths: []string{
								".metadata.labels[\"nvidia.com/dynamo-component\"]",
							},
						},
					},
				},
			},
		},
	}
}
