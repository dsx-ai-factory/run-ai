package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
)

// BuildSystemKubeflowRunPolicyFg builds the field group for Kubeflow Training Operator's
// common runPolicy fields. These are shared across all Kubeflow training jobs
// (PyTorchJob, MPIJob, TFJob, XGBoostJob, JAXJob).
func BuildSystemKubeflowRunPolicyFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(kubeflowRunPolicyFgFields())
	return fields, nil
}

// BuildSystemKubeflowReplicaSpecFg builds the field group for Kubeflow Training Operator's
// common replica spec fields (replicas, restartPolicy). These are shared across all
// Kubeflow training jobs.
func BuildSystemKubeflowReplicaSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(kubeflowReplicaSpecFgFields())
	return fields, nil
}

func kubeflowRunPolicyFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	kubeflowRunPolicy := SystemKubeflowRunPolicyFormatterName

	return []SystemFgField{
		{
			Name:      "runPolicy-CleanPodPolicy",
			Formatter: kubeflowRunPolicy,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "runPolicy-TTLSecondsAfterFinished",
			Formatter: keyValue,
			SpecRef:   strRef("runPolicy.ttlSecondsAfterFinished"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "runPolicy-ActiveDeadlineSeconds",
			Formatter: keyValue,
			SpecRef:   strRef("runPolicy.activeDeadlineSeconds"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "runPolicy-BackoffLimit",
			Formatter: keyValue,
			SpecRef:   strRef("runPolicy.backoffLimit"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "runPolicy-Suspend",
			Formatter: keyValue,
			SpecRef:   strRef("runPolicy.suspend"),
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
	}
}

func kubeflowReplicaSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	kubeflowReplicaSpec := SystemKubeflowReplicaSpecFormatterName

	return []SystemFgField{
		{
			Name:      "replicas",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "restartPolicy",
			Formatter: kubeflowReplicaSpec,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
	}
}
