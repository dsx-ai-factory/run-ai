package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// BuildSystemStatefulSetSpecFormatterFields builds the runai-statefulSetSpec
// system formatter, holding fields whose argsEnum we want enforced at write
// time (string fields whose K8s values are restricted to a closed set).
//
// The setExprs are rooted at the spec-selector anchor (the workload .spec).
func BuildSystemStatefulSetSpecFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	////////////////////////////////////////////////////////
	// podManagementPolicy
	////////////////////////////////////////////////////////
	addStatefulSetSpecField(m, "podManagementPolicy", `{
	  "podManagementPolicy": $ARGS.named.value
	}`, nil, nil, nil,
		map[string][]string{
			"value": {"OrderedReady", "Parallel"},
		})

	////////////////////////////////////////////////////////
	// updateStrategy-Type
	////////////////////////////////////////////////////////
	addStatefulSetSpecField(m, "updateStrategy-Type", `{
	  "updateStrategy": {
		"type": $ARGS.named.value
	  }
	}`, nil, nil, nil,
		map[string][]string{
			"value": {"OnDelete", "RollingUpdate"},
		})

	////////////////////////////////////////////////////////
	// persistentVolumeClaimRetentionPolicy-WhenDeleted
	////////////////////////////////////////////////////////
	addStatefulSetSpecField(m, "persistentVolumeClaimRetentionPolicy-WhenDeleted", `{
	  "persistentVolumeClaimRetentionPolicy": {
		"whenDeleted": $ARGS.named.value
	  }
	}`, nil, nil, nil,
		map[string][]string{
			"value": {"Retain", "Delete"},
		})

	////////////////////////////////////////////////////////
	// persistentVolumeClaimRetentionPolicy-WhenScaled
	////////////////////////////////////////////////////////
	addStatefulSetSpecField(m, "persistentVolumeClaimRetentionPolicy-WhenScaled", `{
	  "persistentVolumeClaimRetentionPolicy": {
		"whenScaled": $ARGS.named.value
	  }
	}`, nil, nil, nil,
		map[string][]string{
			"value": {"Retain", "Delete"},
		})

	////////////////////////////////////////////////////////
	// volumeClaimTemplates
	//
	// Rule-builder rooted at the workload .spec. Produces one PersistentVolumeClaim
	// entry; the engine merges it into the volumeClaimTemplates array by
	// metadata.name (via arrayKey.specRef).
	////////////////////////////////////////////////////////
	addStatefulSetSpecField(m, "volumeClaimTemplates", `{
	  "volumeClaimTemplates": [
		{
		  "metadata": {
			"name": $ARGS.named.name
		  },
		  "spec": {
			"accessModes": $ARGS.named.accessModes,
			"storageClassName": $ARGS.named.storageClassName,
			"volumeMode": $ARGS.named.volumeMode,
			"resources": {
			  "requests": {
				"storage": $ARGS.named.storageRequest
			  }
			}
		  }
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{
			ArgName: "name",
			SpecRef: convutils.Ref("metadata.name"),
		},
		map[string]api.MappingFieldType{
			"accessModes":    api.MAPPINGFIELDTYPE_STRINGS,
			"storageRequest": api.MAPPINGFIELDTYPE_QUANTITY,
		},
		map[string][]string{
			"volumeMode": {"Filesystem", "Block"},
		})

	////////////////////////////////////////////////////////
	// volumeClaimTemplates-VolumeMode
	//
	// Rooted at a single PVC entry (selected via the field's item selectors).
	// Sets/defaults the entry's spec.volumeMode with enum enforcement.
	////////////////////////////////////////////////////////
	addStatefulSetSpecField(m, "volumeClaimTemplates-VolumeMode", `{
	  "spec": {
		"volumeMode": $ARGS.named.value
	  }
	}`, nil, nil, nil,
		map[string][]string{
			"value": {"Filesystem", "Block"},
		})

	return m
}

func addStatefulSetSpecField(
	dst map[string]api.FormatterField,
	name string,
	setExpr string,
	getExpr *string,
	arrayKey *api.FormatterArrayKey,
	argsInfo map[string]api.MappingFieldType,
	argsEnum map[string][]string,
) {
	ff := api.FormatterField{
		SetExpr:  &setExpr,
		GetExpr:  getExpr,
		ArrayKey: arrayKey,
	}
	if len(argsInfo) > 0 {
		ff.ArgsInfo = argsInfo
	}
	if len(argsEnum) > 0 {
		ff.ArgsEnum = argsEnum
	}
	dst[name] = ff
}
