package common

const (
	BEARER                = "BEARER "
	PolicyTypeInteractive = "InteractivePolicy"
	PolicyTypeTraining    = "TrainingPolicy"
	PolicyTypeDistributed = "DistributedPolicy"
	PolicyTypeInference   = "InferencePolicy"

	WorkloadTypeInteractive          = "InteractiveWorkload"
	WorkloadTypeTraining             = "TrainingWorkload"
	WorkloadTypeDistributed          = "DistributedWorkload"
	WorkloadTypeInference            = "InferenceWorkload"
	WorkloadTypeDistributedInference = "DistributedInferenceWorkload"

	WorkloadsV2Type  = "WorkloadsV2"
	ApplicationsType = "Applications"

	DatasourceAssets = "DatasourceAssets"
	CredentialAssets = "CredentialAssets"
)

func GetStringVal(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

func GetFloatVal(num *float64) float64 {
	return GetFloatValDefault(num, 0)
}

func GetFloatValDefault(num *float64, defaultVal float64) float64 {
	if num == nil {
		return defaultVal
	}
	return *num
}
