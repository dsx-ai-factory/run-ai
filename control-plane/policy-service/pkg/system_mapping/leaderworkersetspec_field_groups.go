package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

const (
	SystemLeaderWorkerSetSpecFgName    = "leaderWorkerSetSpec"
	SystemLeaderWorkerTemplateFgName   = "leaderWorkerTemplate"
	SystemLeaderWorkerSetFormatterName = "runai-leaderWorkerSetSpec"
)

// BuildSystemLeaderWorkerSetSpecFg returns the workload-level field group for LeaderWorkerSet.
// Fields: replicas, startupPolicy, networkConfig-SubdomainPolicy, rolloutStrategy-*
func BuildSystemLeaderWorkerSetSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	return map[string]api.FieldGroupField{
		"replicas": {
			Formatter: SystemLeaderWorkerSetFormatterName,
			SpecRef:   "replicas",
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		"startupPolicy": {
			Formatter: SystemLeaderWorkerSetFormatterName,
			SpecRef:   "startupPolicy",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"networkConfig-SubdomainPolicy": {
			Formatter: SystemLeaderWorkerSetFormatterName,
			SpecRef:   "networkConfig.subdomainPolicy",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"rolloutStrategy-Type": {
			Formatter: SystemLeaderWorkerSetFormatterName,
			SpecRef:   "rolloutStrategy.type",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"rolloutStrategy-MaxSurge": {
			Formatter: SystemLeaderWorkerSetFormatterName,
			SpecRef:   "rolloutStrategy.rollingUpdateConfiguration.maxSurge",
			Type:      api.MAPPINGFIELDTYPE_STRING, // IntOrString
		},
		"rolloutStrategy-MaxUnavailable": {
			Formatter: SystemLeaderWorkerSetFormatterName,
			SpecRef:   "rolloutStrategy.rollingUpdateConfiguration.maxUnavailable",
			Type:      api.MAPPINGFIELDTYPE_STRING, // IntOrString
		},
	}, map[string]api.ItemSelector{}
}

// BuildSystemLeaderWorkerTemplateFg returns the template-level field group for LeaderWorkerSet.
// Fields: size, restartPolicy, subGroupPolicy-SubGroupSize
func BuildSystemLeaderWorkerTemplateFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	return map[string]api.FieldGroupField{
		"size": {
			Formatter: SystemLeaderWorkerSetFormatterName,
			SpecRef:   "size",
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		"restartPolicy": {
			Formatter: SystemLeaderWorkerSetFormatterName,
			SpecRef:   "restartPolicy",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"subGroupPolicy-SubGroupSize": {
			Formatter: SystemLeaderWorkerSetFormatterName,
			SpecRef:   "subGroupPolicy.subGroupSize",
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
	}, map[string]api.ItemSelector{}
}
