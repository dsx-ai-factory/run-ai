package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

// BuildSystemPyTorchJobSpecFormatterFields returns custom formatter fields for
// kubeflow.org/PyTorchJob/v1 workload-level fields that need enum validation.
//
// Note: runPolicy-CleanPodPolicy is provided by the shared kubeflowRunPolicy formatter.
func BuildSystemPyTorchJobSpecFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	////////////////////////////////////////////////////////
	// elasticPolicy-RDZVBackend (PyTorchJob-specific)
	////////////////////////////////////////////////////////
	addPyTorchJobSpecField(m, "elasticPolicy-RDZVBackend", `{
	  "elasticPolicy": {
		"rdzvBackend": $ARGS.named.value
	  }
	}`, nil,
		map[string][]string{
			"value": {"c10d", "etcd", "etcd-v2"},
		})

	return m
}

func addPyTorchJobSpecField(
	dst map[string]api.FormatterField,
	name string,
	setExpr string,
	argsInfo map[string]api.MappingFieldType,
	argsEnum map[string][]string,
) {
	ff := api.FormatterField{
		SetExpr: &setExpr,
	}
	if len(argsInfo) > 0 {
		ff.ArgsInfo = argsInfo
	}
	if len(argsEnum) > 0 {
		ff.ArgsEnum = argsEnum
	}
	dst[name] = ff
}
