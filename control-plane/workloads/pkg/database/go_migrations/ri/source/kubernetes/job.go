package kubernetes

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/go_migrations/ri/source"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

func GetK8sJobRIConfiguration() source.KartaConfiguration {
	return source.KartaConfiguration{
		Category:       "Train",
		Priority:       "low",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildK8sJobV1(),
		},
	}
}

func BuildK8sJobV1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "job",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "batch",
						Version: "v1",
						Kind:    "Job",
					},
					StatusDefinition: buildJobStatusDefinition(),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: ptr.To(".spec.template"),
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To(".spec.parallelism // 1"),
					},
					SuspendDefinition: buildK8sSuspendDefinition(),
				},
			},
		},
	}
}

func buildK8sSuspendDefinition() *kartav1alpha1.SuspendDefinition {
	return &kartav1alpha1.SuspendDefinition{
		SuspendActions: []kartav1alpha1.SuspendAction{
			{Path: ".spec.suspend", Value: "true"},
		},
		ResumeActions: []kartav1alpha1.SuspendAction{
			{Path: ".spec.suspend", Value: "false"},
		},
	}
}

func buildJobStatusDefinition() *kartav1alpha1.StatusDefinition {
	return &kartav1alpha1.StatusDefinition{
		ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
			Path:             ".status.conditions",
			TypeFieldName:    "type",
			StatusFieldName:  "status",
			MessageFieldName: ptr.To("message"),
		},
		StatusMappings: kartav1alpha1.StatusMappings{
			Completed: []kartav1alpha1.StatusMatcher{
				{
					ByConditions: []kartav1alpha1.ExpectedCondition{
						{Type: "Complete", Status: ptr.To("True")},
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
			Running: []kartav1alpha1.StatusMatcher{
				{
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     "(.status.active // 0) > 0",
						ExpectedResult: "true",
					},
				},
				{
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     "(.status.active // 0) > 0 and (.status.ready // 0) > 0",
						ExpectedResult: "true",
					},
				},
			},
			Initializing: []kartav1alpha1.StatusMatcher{
				{
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     "(.status.active // 0) > 0 and (.status.ready == 0)",
						ExpectedResult: "true",
					},
				},
			},
			Suspended: []kartav1alpha1.StatusMatcher{
				{
					ByConditions: []kartav1alpha1.ExpectedCondition{
						{Type: "Suspended", Status: ptr.To("True")},
					},
				},
			},
			Suspending: []kartav1alpha1.StatusMatcher{
				{
					ByConditions: []kartav1alpha1.ExpectedCondition{
						{Type: "Suspended", Status: ptr.To("False")},
					},
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     ".spec.suspend == true and (.status.active // 0) > 0",
						ExpectedResult: "true",
					},
				},
			},
		},
	}
}
