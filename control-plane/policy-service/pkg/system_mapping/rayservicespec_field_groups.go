package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
)

// BuildSystemRayServiceSpecFg builds the field group for ray.io/RayService/v1
// workload-level scalars under .spec (excluding the embedded rayClusterConfig,
// which uses the shared SystemRayClusterSpecFgName).
func BuildSystemRayServiceSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(rayServiceSpecFgFields())
	return fields, nil
}

func rayServiceSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	rayServiceSpec := SystemRayServiceSpecFormatterName

	return []SystemFgField{
		{
			Name:      "serveConfigV2",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "upgradeStrategy-Type",
			Formatter: rayServiceSpec,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "deploymentUnhealthySecondThreshold",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "serviceUnhealthySecondThreshold",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "excludeHeadPodFromServeSvc",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
	}
}
