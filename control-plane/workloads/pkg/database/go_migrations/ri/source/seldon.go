package source

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

func GetSeldonDeploymentRIConfiguration() KartaConfiguration {
	return KartaConfiguration{
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		Kartas: []models.Karta{
			BuildSeldonDeploymentV1(),
			BuildSeldonDeploymentV1alpha2(),
		},
	}
}

func BuildSeldonDeploymentV1() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "seldondeployment",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "machinelearning.seldon.io",
						Version: "v1",
						Kind:    "SeldonDeployment",
					},
					StatusDefinition: buildSeldonStatusDefinition(),
					ScaleDefinition:  &kartav1alpha1.ScaleDefinition{ReplicasPath: ptr.To(".spec.replicas")},
				},
				ChildComponents: buildSeldonChildComponents(),
			},
		},
	}
}

func BuildSeldonDeploymentV1alpha2() models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "seldondeployment",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "machinelearning.seldon.io",
						Version: "v1alpha2",
						Kind:    "SeldonDeployment",
					},
					StatusDefinition: buildSeldonStatusDefinition(),
				},
				ChildComponents: buildSeldonChildComponents(),
			},
		},
	}
}

func buildSeldonChildComponents() []kartav1alpha1.ComponentDefinition {
	return []kartav1alpha1.ComponentDefinition{
		{
			Name:     "component",
			OwnerRef: ptr.To("seldondeployment"),
			ScaleDefinition: &kartav1alpha1.ScaleDefinition{
				ReplicasPath:    ptr.To(".spec.predictors[].componentSpecs[].replicas"),
				MinReplicasPath: ptr.To(".spec.predictors[].componentSpecs[].hpaSpec.minReplicas"),
				MaxReplicasPath: ptr.To(".spec.predictors[].componentSpecs[].hpaSpec.maxReplicas"),
			},
			SpecDefinition: &kartav1alpha1.SpecDefinition{
				PodSpecPath:  ptr.To(".spec.predictors[].componentSpecs[].spec"),
				MetadataPath: ptr.To(".spec.predictors[].componentSpecs[].metadata"),
			},
			// currently we do not support owner ref for 3rd level components,
			// so all the components specs will point to the Seldon deployment, instead of the predicators.
			InstanceIdPath: ptr.To(".spec.predictors[] | .name as $pname | .componentSpecs | to_entries[] | \"\\($pname)-\\(.key)\""),

			PodSelector: &kartav1alpha1.PodSelector{
				ComponentInstanceSelector: &kartav1alpha1.ComponentInstanceSelector{
					// The predicator name is the value of the version label of the pod.
					// inorder to generate the same instance id, we extract the predicator name,
					// and it helps us to extract the insatnce id for the app label.
					// example to pod labels: app=ab-test-model-new-model-v2-0-v2-classifier, version=new-model-v2
					IdPath: ".metadata.labels[\"version\"] as $predicator | .metadata.labels[\"app\"] | split($predicator + \"-\")[1] | split(\"-\")[0] as $component_index | \"\\($predicator)-\\($component_index)\"",
				},
			},
		},
	}
}

// buildSeldonStatusDefinition creates the status definition for SeldonDeployment.
// SeldonDeployment uses .status.state with values: Available, Creating, Failed
func buildSeldonStatusDefinition() *kartav1alpha1.StatusDefinition {
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
			Running: []kartav1alpha1.StatusMatcher{
				{ByPhase: "Available"},
			},
			Initializing: []kartav1alpha1.StatusMatcher{
				{ByPhase: "Creating"},
			},
			Failed: []kartav1alpha1.StatusMatcher{
				{ByPhase: "Failed"},
			},
		},
	}
}
