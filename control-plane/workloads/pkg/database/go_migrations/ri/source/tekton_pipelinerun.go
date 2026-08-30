package source

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

func GetTektonPipelineRunRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildTektonPipelineRunV1(),
		},
	}
}

func BuildTektonPipelineRunV1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "pipelinerun",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "tekton.dev",
						Version: "v1",
						Kind:    "PipelineRun",
					},
					StatusDefinition: buildTektonPipelineRunStatusDefinition(),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: ptr.To(".spec.podTemplate"),
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To("1"),
					},
				},
			},
		},
	}
}

func buildTektonPipelineRunStatusDefinition() *kartav1alpha1.StatusDefinition {
	return buildTektonStatusDefinition()
}
