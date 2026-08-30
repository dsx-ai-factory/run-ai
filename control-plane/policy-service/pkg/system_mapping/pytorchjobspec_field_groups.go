package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
)

// BuildSystemPyTorchJobSpecFg builds the field group for kubeflow.org/PyTorchJob/v1
// workload-level fields. This includes:
//   - elasticPolicy fields (PyTorchJob-specific)
//
// Note: runPolicy fields are provided by the shared kubeflowRunPolicy field group.
func BuildSystemPyTorchJobSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(pytorchJobSpecFgFields())
	return fields, nil
}

func pytorchJobSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	pytorchJobSpec := SystemPyTorchJobSpecFormatterName

	return []SystemFgField{
		////////////////////////////////////////////////////////
		// elasticPolicy fields (PyTorchJob-specific)
		////////////////////////////////////////////////////////
		{
			Name:      "elasticPolicy-MinReplicas",
			Formatter: keyValue,
			SpecRef:   strRef("elasticPolicy.minReplicas"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "elasticPolicy-MaxReplicas",
			Formatter: keyValue,
			SpecRef:   strRef("elasticPolicy.maxReplicas"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "elasticPolicy-RDZVBackend",
			Formatter: pytorchJobSpec,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "elasticPolicy-NProcPerNode",
			Formatter: keyValue,
			SpecRef:   strRef("elasticPolicy.nProcPerNode"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "elasticPolicy-MaxRestarts",
			Formatter: keyValue,
			SpecRef:   strRef("elasticPolicy.maxRestarts"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "elasticPolicy-RDZVPort",
			Formatter: keyValue,
			SpecRef:   strRef("elasticPolicy.rdzvPort"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "elasticPolicy-RDZVHost",
			Formatter: keyValue,
			SpecRef:   strRef("elasticPolicy.rdzvHost"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "elasticPolicy-RDZVID",
			Formatter: keyValue,
			SpecRef:   strRef("elasticPolicy.rdzvId"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
	}
}

func strRef(s string) *string {
	return &s
}
