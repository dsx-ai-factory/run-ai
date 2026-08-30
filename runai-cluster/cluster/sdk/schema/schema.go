package schema

import (
	"reflect"

	"github.com/invopop/jsonschema"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const description = "JSON Schema for RunAI Helm Chart Values based on RunaiConfigSpec"

// ValuesSchema is the root structure for Helm values
type ValuesSchema struct {
	ClusterConfig *runaiv1.RunaiConfigSpec `json:"clusterConfig,omitempty" jsonschema:"title=RunAI Cluster Configuration,description=RunAI cluster configuration"`
}

// Generate builds the JSON Schema for Helm values using reflection and custom mappings
func Generate() *jsonschema.Schema {
	r := jsonschema.Reflector{
		DoNotReference:            false,
		AllowAdditionalProperties: true,
	}

	// Map common Kubernetes wrapper types to their JSON representations
	qtyType := reflect.TypeOf(k8sresource.Quantity{})
	timeType := reflect.TypeOf(metav1.Time{})
	r.Mapper = func(t reflect.Type) *jsonschema.Schema {
		if t == qtyType {
			return &jsonschema.Schema{Type: "string", Pattern: "^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$"}
		}
		if t == timeType {
			return &jsonschema.Schema{Type: "string", Format: "date-time"}
		}
		return nil
	}

	schema := r.Reflect(&ValuesSchema{})
	schema.Title = "clusterConfig Helm Chart Values Schema"
	schema.Description = description
	schema.Version = "https://json-schema.org/draft/2020-12/schema"
	return schema
}
