package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

// BuildSystemSelectorFormatterFields builds the runai-selector system formatter,
// shared across workload mappings whose spec carries a top-level LabelSelector
// at .spec.selector (e.g. apps/Deployment/v1, batch/Job/v1).
//
// The setExprs are rooted at the spec-selector anchor (the workload .spec).
func BuildSystemSelectorFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	////////////////////////////////////////////////////////
	// matchExprSelector
	////////////////////////////////////////////////////////
	addSelectorField(m, "matchExprSelector", `{
	  "selector": {
		"matchExpressions": [
		  {
			"key": $ARGS.named.key,
			"operator": $ARGS.named.operator,
			"values": $ARGS.named.values
		  }
		]
	  }
	}`, nil,
		&api.FormatterArrayKey{ArgName: "key"},
		map[string]api.MappingFieldType{
			"values": api.MAPPINGFIELDTYPE_STRINGS,
		},
		map[string][]string{
			"operator": {"In", "NotIn", "Exists", "DoesNotExist"},
		})

	////////////////////////////////////////////////////////
	// matchLabelsSelector read companion
	//
	// matchLabelsSelector itself is a KeyValueMap (set-only); to read /
	// validate map values, policies use this GetExpr-only companion paired
	// with the matchLabelByKey item selector which already resolves to the
	// final scalar — so the formatter just returns ".".
	////////////////////////////////////////////////////////
	addSelectorGetterField(m, "matchLabelsSelector-value")

	return m
}

// addSelectorGetterField registers a GET-only companion (no SetExpr): the
// item selector does the navigation to the final scalar and the formatter
// just returns it as-is.
func addSelectorGetterField(dst map[string]api.FormatterField, name string) {
	getExpr := "."
	dst[name] = api.FormatterField{GetExpr: &getExpr}
}

func addSelectorField(
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
