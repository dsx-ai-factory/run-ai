package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

// BuildSystemCronJobSpecFormatterFields builds the runai-cronJobSpec system
// formatter, holding fields whose argsEnum we want enforced at write time
// (string fields whose K8s values are restricted to a closed set).
//
// The setExprs are rooted at the spec-selector anchor (the workload .spec).
func BuildSystemCronJobSpecFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	////////////////////////////////////////////////////////
	// concurrencyPolicy
	////////////////////////////////////////////////////////
	addCronJobSpecField(m, "concurrencyPolicy", `{
	  "concurrencyPolicy": $ARGS.named.value
	}`, nil, nil, nil,
		map[string][]string{
			"value": {"Allow", "Forbid", "Replace"},
		})

	return m
}

func addCronJobSpecField(
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
