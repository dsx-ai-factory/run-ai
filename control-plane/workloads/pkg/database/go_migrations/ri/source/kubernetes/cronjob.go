/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kubernetes

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/go_migrations/ri/source"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

func GetK8sCronJobRIConfiguration() source.KartaConfiguration {
	return source.KartaConfiguration{
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildK8sCronJobV1(),
		},
	}
}

func BuildK8sCronJobV1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "cronjob",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "batch",
						Version: "v1",
						Kind:    "CronJob",
					},
					StatusDefinition: buildCronJobStatusDefinition(),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: ptr.To(".spec.jobTemplate.spec.template"),
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To(".spec.jobTemplate.spec.parallelism // 1"),
					},
					SuspendDefinition: buildK8sSuspendDefinition(),
				},
				ChildComponents: []kartav1alpha1.ComponentDefinition{
					{
						Name: "job",
						Kind: &kartav1alpha1.GroupVersionKind{
							Group:   "batch",
							Version: "v1",
							Kind:    "Job",
						},
						OwnerRef: ptr.To("cronjob"),
					},
				},
			},
		},
	}
}

func buildCronJobStatusDefinition() *kartav1alpha1.StatusDefinition {
	return &kartav1alpha1.StatusDefinition{
		StatusMappings: kartav1alpha1.StatusMappings{
			Running: []kartav1alpha1.StatusMatcher{
				{
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     "(.status.active // [] | length) > 0",
						ExpectedResult: "true",
					},
				},
			},
			Suspended: []kartav1alpha1.StatusMatcher{
				{
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     ".spec.suspend == true and ((.status.active // [] | length) == 0)",
						ExpectedResult: "true",
					},
				},
			},
			Suspending: []kartav1alpha1.StatusMatcher{
				{
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     ".spec.suspend == true and ((.status.active // [] | length) > 0)",
						ExpectedResult: "true",
					},
				},
			},
		},
	}
}
