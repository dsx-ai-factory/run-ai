package nvidia

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/go_migrations/ri/source"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

func GetNIMCacheRIConfiguration() source.KartaConfiguration {
	return source.KartaConfiguration{
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildNIMCacheV1alpha1(),
		},
	}
}

func BuildNIMCacheV1alpha1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "nimcache",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "apps.nvidia.com",
						Version: "v1alpha1",
						Kind:    "NIMCache",
					},
					StatusDefinition: buildNIMCacheStatusDefinition(),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
							ResourcesPath: ptr.To(".spec.resources"),
						},
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To("1"),
					},
				},
			},
		},
	}
}

func buildNIMCacheStatusDefinition() *kartav1alpha1.StatusDefinition {
	return &kartav1alpha1.StatusDefinition{
		PhaseDefinition: &kartav1alpha1.PhaseDefinition{
			Path: ".status.state",
		},
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
						{Type: "NIM_CACHE_JOB_COMPLETED", Status: ptr.To("True")},
					},
					ByPhase: "Ready",
				},
			},
			Running: []kartav1alpha1.StatusMatcher{
				{ByPhase: "InProgress"},
			},
			Initializing: []kartav1alpha1.StatusMatcher{
				{ByPhase: "PVC-Created"},
				{ByPhase: "Started"},
				{
					ByConditions: []kartav1alpha1.ExpectedCondition{
						{Type: "NIM_CACHE_JOB_CREATED", Status: ptr.To("True")},
					},
				},
			},
			Failed: []kartav1alpha1.StatusMatcher{
				{ByPhase: "Failed"},
			},
		},
	}
}
