package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

// BuildSystemKubeflowRunPolicyFormatterFields returns formatter fields for the
// Kubeflow Training Operator's common runPolicy fields. These are shared across
// all Kubeflow training jobs (PyTorchJob, MPIJob, TFJob, XGBoostJob, JAXJob).
func BuildSystemKubeflowRunPolicyFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	// runPolicy-CleanPodPolicy is an enum with the same values across all job types.
	addKubeflowFormatterField(m, "runPolicy-CleanPodPolicy", `{
	  "runPolicy": {
		"cleanPodPolicy": $ARGS.named.value
	  }
	}`, nil,
		map[string][]string{
			"value": {"All", "Running", "None"},
		})

	return m
}

// BuildSystemKubeflowReplicaSpecFormatterFields returns formatter fields for the
// Kubeflow Training Operator's common replica spec fields. These are shared across
// all Kubeflow training jobs.
func BuildSystemKubeflowReplicaSpecFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	// restartPolicy is an enum with the same values across all job types.
	addKubeflowFormatterField(m, "restartPolicy", `{
	  "restartPolicy": $ARGS.named.value
	}`, nil,
		map[string][]string{
			"value": {"Always", "OnFailure", "Never", "ExitCode"},
		})

	return m
}

func addKubeflowFormatterField(
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
