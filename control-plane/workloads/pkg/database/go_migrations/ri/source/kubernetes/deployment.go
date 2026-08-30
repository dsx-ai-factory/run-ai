package kubernetes

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/go_migrations/ri/source"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

func GetK8sDeploymentRIConfiguration() source.KartaConfiguration {
	return source.KartaConfiguration{
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildK8sDeploymentV1(),
		},
	}
}

func BuildK8sDeploymentV1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "deployment",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "apps",
						Version: "v1",
						Kind:    "Deployment",
					},
					StatusDefinition: buildDeploymentStatusDefinition(),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: ptr.To(".spec.template"),
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To(".spec.replicas // 1"),
					},
				},
				ChildComponents: []kartav1alpha1.ComponentDefinition{
					{
						Name: "replicaset",
						Kind: &kartav1alpha1.GroupVersionKind{
							Group:   "apps",
							Version: "v1",
							Kind:    "ReplicaSet",
						},
						OwnerRef: ptr.To("deployment"),
					},
				},
			},
		},
	}
}

func buildDeploymentStatusDefinition() *kartav1alpha1.StatusDefinition {
	return &kartav1alpha1.StatusDefinition{
		ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
			Path:             ".status.conditions",
			TypeFieldName:    "type",
			StatusFieldName:  "status",
			MessageFieldName: ptr.To("message"),
			ReasonFieldName:  ptr.To("reason"),
		},
		StatusMappings: kartav1alpha1.StatusMappings{
			Running: []kartav1alpha1.StatusMatcher{
				{
					ByConditions: []kartav1alpha1.ExpectedCondition{
						{Type: "Progressing", Status: ptr.To("True"), Reason: ptr.To("NewReplicaSetAvailable")},
					},
				},
			},
			Failed: []kartav1alpha1.StatusMatcher{
				{
					ByConditions: []kartav1alpha1.ExpectedCondition{
						{Type: "Progressing", Status: ptr.To("False"), Reason: ptr.To("ProgressDeadlineExceeded")},
					},
				},
			},
			Initializing: []kartav1alpha1.StatusMatcher{
				{
					ByConditions: []kartav1alpha1.ExpectedCondition{
						{Type: "Progressing", Status: ptr.To("True")},
						{Type: "Available", Status: ptr.To("False")},
					},
				},
			},
		},
	}
}
