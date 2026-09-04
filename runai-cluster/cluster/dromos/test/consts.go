/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// kartaLabelsForGVK stamps the GVK-routing labels dromos's karta lookup
// matches on (via GetKartaPerGvkLabelSelectors). Inlined here rather than
// imported from pkg/utils/karta to avoid an import cycle
// (test → kartautils → internaltest → test).
func kartaLabelsForGVK(group, version, kind string) map[string]string {
	return map[string]string{
		"run.ai/karta-group":   group,
		"run.ai/karta-version": version,
		"run.ai/karta-kind":    kind,
	}
}

var gvk = metav1.GroupVersionKind{
	Group:   "jobset.x-k8s.io",
	Version: "v1alpha2",
	Kind:    "JobSet",
}

func NewValidJobSetKarta() *kartav1alpha1.Karta {
	return validJobSetKartaWithPodTemplateSpec.DeepCopy()
}

func NewValidJobSetKartaWithPodTemplateSpec() *kartav1alpha1.Karta {
	return validJobSetKartaWithPodTemplateSpec.DeepCopy()
}

func NewValidJobSetKartaWithPodSpec() *kartav1alpha1.Karta {
	return validJobSetKartaWithPodSpec.DeepCopy()
}

func NewValidJobSetKartaWithFragmentedPodSpec() *kartav1alpha1.Karta {
	return validJobSetKartaWithFragmentedPodSpec.DeepCopy()
}

var validJobSetKartaWithPodTemplateSpec = &kartav1alpha1.Karta{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "valid-jobset-karta-podtemplatespec",
		Labels: kartaLabelsForGVK(gvk.Group, gvk.Version, gvk.Kind),
	},
	Spec: kartav1alpha1.KartaSpec{
		StructureDefinition: kartav1alpha1.StructureDefinition{
			RootComponent: kartav1alpha1.ComponentDefinition{
				Name: "root",
				Kind: &kartav1alpha1.GroupVersionKind{
					Group:   gvk.Group,
					Version: gvk.Version,
					Kind:    gvk.Kind,
				},
				StatusDefinition: &kartav1alpha1.StatusDefinition{
					ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
						Path:            ".status.conditions",
						TypeFieldName:   "type",
						StatusFieldName: "status",
					},
					StatusMappings: kartav1alpha1.StatusMappings{
						Initializing: []kartav1alpha1.StatusMatcher{
							{
								ByConditions: []kartav1alpha1.ExpectedCondition{
									{Type: "StartupPolicyInProgress", Status: ptr.To("True")},
								},
							},
						},
						Running: []kartav1alpha1.StatusMatcher{
							{
								ByConditions: []kartav1alpha1.ExpectedCondition{
									{Type: "StartupPolicyCompleted", Status: ptr.To("True")},
									{Type: "Completed", Status: ptr.To("False")},
									{Type: "Failed", Status: ptr.To("False")},
								},
							},
						},
						Completed: []kartav1alpha1.StatusMatcher{
							{
								ByConditions: []kartav1alpha1.ExpectedCondition{
									{Type: "Completed", Status: ptr.To("True")},
								},
							},
						},
						Failed: []kartav1alpha1.StatusMatcher{
							{
								ByConditions: []kartav1alpha1.ExpectedCondition{
									{Type: "Failed", Status: ptr.To("True")},
								},
							},
						},
					},
				},
			},
			ChildComponents: []kartav1alpha1.ComponentDefinition{
				{
					Name: "replicatedjob",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "batch",
						Version: "v1",
						Kind:    "Job",
					},
					OwnerRef: ptr.To("jobset"),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: ptr.To(".spec.replicatedJobs[].template.spec.template"),
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To(".spec.replicatedJobs[].replicas"),
					},
				},
			},
		},
	},
}

var validJobSetKartaWithPodSpec = &kartav1alpha1.Karta{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "valid-jobset-karta-podspec",
		Labels: kartaLabelsForGVK(gvk.Group, gvk.Version, gvk.Kind),
	},
	Spec: kartav1alpha1.KartaSpec{
		StructureDefinition: kartav1alpha1.StructureDefinition{
			RootComponent: kartav1alpha1.ComponentDefinition{
				Name: "root",
				Kind: &kartav1alpha1.GroupVersionKind{
					Group:   gvk.Group,
					Version: gvk.Version,
					Kind:    gvk.Kind,
				},
				StatusDefinition: &kartav1alpha1.StatusDefinition{
					ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
						Path:            ".status.conditions",
						TypeFieldName:   "type",
						StatusFieldName: "status",
					},
					StatusMappings: kartav1alpha1.StatusMappings{
						Initializing: []kartav1alpha1.StatusMatcher{
							{
								ByConditions: []kartav1alpha1.ExpectedCondition{
									{Type: "StartupPolicyInProgress", Status: ptr.To("True")},
								},
							},
						},
						Running: []kartav1alpha1.StatusMatcher{
							{
								ByConditions: []kartav1alpha1.ExpectedCondition{
									{Type: "StartupPolicyCompleted", Status: ptr.To("True")},
									{Type: "Completed", Status: ptr.To("False")},
									{Type: "Failed", Status: ptr.To("False")},
								},
							},
						},
						Completed: []kartav1alpha1.StatusMatcher{
							{
								ByConditions: []kartav1alpha1.ExpectedCondition{
									{Type: "Completed", Status: ptr.To("True")},
								},
							},
						},
						Failed: []kartav1alpha1.StatusMatcher{
							{
								ByConditions: []kartav1alpha1.ExpectedCondition{
									{Type: "Failed", Status: ptr.To("True")},
								},
							},
						},
					},
				},
			},
			ChildComponents: []kartav1alpha1.ComponentDefinition{
				{
					Name: "replicatedjob",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "batch",
						Version: "v1",
						Kind:    "Job",
					},
					OwnerRef: ptr.To("jobset"),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodSpecPath:  ptr.To(".spec.replicatedJobs[].template.spec.template.spec"),
						MetadataPath: ptr.To(".spec.replicatedJobs[].template.spec.template.metadata"),
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To(".spec.replicatedJobs[].replicas"),
					},
				},
			},
		},
	},
}

var validJobSetKartaWithFragmentedPodSpec = &kartav1alpha1.Karta{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "valid-jobset-karta-fragmented",
		Labels: kartaLabelsForGVK(gvk.Group, gvk.Version, gvk.Kind),
	},
	Spec: kartav1alpha1.KartaSpec{
		StructureDefinition: kartav1alpha1.StructureDefinition{
			RootComponent: kartav1alpha1.ComponentDefinition{
				Name: "root",
				Kind: &kartav1alpha1.GroupVersionKind{
					Group:   gvk.Group,
					Version: gvk.Version,
					Kind:    gvk.Kind,
				},
				StatusDefinition: &kartav1alpha1.StatusDefinition{
					ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
						Path:            ".status.conditions",
						TypeFieldName:   "type",
						StatusFieldName: "status",
					},
					StatusMappings: kartav1alpha1.StatusMappings{
						Initializing: []kartav1alpha1.StatusMatcher{
							{
								ByConditions: []kartav1alpha1.ExpectedCondition{
									{Type: "StartupPolicyInProgress", Status: ptr.To("True")},
								},
							},
						},
						Running: []kartav1alpha1.StatusMatcher{
							{
								ByConditions: []kartav1alpha1.ExpectedCondition{
									{Type: "StartupPolicyCompleted", Status: ptr.To("True")},
									{Type: "Completed", Status: ptr.To("False")},
									{Type: "Failed", Status: ptr.To("False")},
								},
							},
						},
						Completed: []kartav1alpha1.StatusMatcher{
							{
								ByConditions: []kartav1alpha1.ExpectedCondition{
									{Type: "Completed", Status: ptr.To("True")},
								},
							},
						},
						Failed: []kartav1alpha1.StatusMatcher{
							{
								ByConditions: []kartav1alpha1.ExpectedCondition{
									{Type: "Failed", Status: ptr.To("True")},
								},
							},
						},
					},
				},
			},
			ChildComponents: []kartav1alpha1.ComponentDefinition{
				{
					Name: "replicatedjob",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "batch",
						Version: "v1",
						Kind:    "Job",
					},
					OwnerRef: ptr.To("jobset"),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
							ContainersPath:     ptr.To(".spec.replicatedJobs[].template.spec.template.spec.containers"),
							ResourceClaimsPath: ptr.To(".spec.replicatedJobs[].template.spec.template.spec.resourceClaims"),
							NodeAffinityPath:   ptr.To(".spec.replicatedJobs[].template.spec.template.spec.affinity.nodeAffinity"),
							PodAffinityPath:    ptr.To(".spec.replicatedJobs[].template.spec.template.spec.affinity.podAffinity"),
							LabelsPath:         ptr.To(".spec.replicatedJobs[].template.spec.template.metadata.labels"),
						},
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To(".spec.replicatedJobs[].replicas"),
					},
				},
			},
		},
	},
}
